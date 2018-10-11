package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/elfido/n1qlExporter/cbapi"

	"github.com/elfido/n1qlExporter/datamonitor"
	"github.com/elfido/n1qlExporter/n1qlmonitor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var listenAddr = flag.String("listen", ":8380", "Address to listen for HTTP requests")

func init() {
	initN1QLMetrics()
}

func discoverCluster(hostName string, auth cbapi.Auth) (datamonitor.ClusterMap, error) {
	return datamonitor.GetClusterMap(hostName, auth)
}

func getMonitors() []n1qlmonitor.Monitor {
	monitorDefinitions := getConfigurationDefs()
	monitors := make([]n1qlmonitor.Monitor, len(monitorDefinitions), len(monitorDefinitions))
	for ndx, definition := range monitorDefinitions {
		if len(definition.hosts) > 0 {
			protocol := "http"
			if definition.useHTTPS {
				protocol = "https"
			}
			server := protocol + "://" + definition.hosts[0]
			clusterMap, err := discoverCluster(server, definition.auth)
			if err == nil {
				definition.clusterName = strings.ToUpper(definition.clusterName)
				log.Printf("Registering monitor %s for hosts: %v\n", definition.clusterName, clusterMap.QueryNodes)
				datelayout := time.RFC3339Nano
				versionsplit := strings.Split(clusterMap.Version, ".")
				if versionsplit[0] == "5" {
					datelayout = "2006-01-02 15:04:05.999999999 -0700 MST"
				}
				mon := n1qlmonitor.New(definition.clusterName, clusterMap.QueryNodes, definition.auth, definition.useHTTPS, datelayout)
				monitors[ndx] = mon
			} else {
				fmt.Printf("Cannot discover cluster %s: %s\n", definition.clusterName, err.Error())
			}
		}
	}
	return monitors
}

func reportMetrics(metrics *n1qlmonitor.ClusterResponse) {
	for _, server := range metrics.ServerResponses {
		// Active queries report
		for _, query := range server.Active {
			activeExecutionTime.WithLabelValues(metrics.ClusterName, server.Node, query.QueryType).Observe(float64(query.ExecutionTime))
			activeWaitingTime.WithLabelValues(metrics.ClusterName, server.Node, query.QueryType).Observe(float64(query.WaitingTime))
			activeScanConsistency.WithLabelValues(metrics.ClusterName, query.ScanConsistency).Inc()
		}
		activeAccumulation.WithLabelValues(metrics.ClusterName, server.Node).Observe(float64(len(server.Active)))

		// Completed queries report
		for _, query := range server.Completed {
			completedResultCount.WithLabelValues(metrics.ClusterName, query.QueryType).Observe(float64(query.ResultCount))
			completedResultSize.WithLabelValues(metrics.ClusterName, query.QueryType).Observe(float64(query.ResultSize))
			completedExecutionTime.WithLabelValues(metrics.ClusterName, server.Node, query.QueryType, query.State).Observe(float64(query.ExecutionTime))
			completedWaitingTime.WithLabelValues(metrics.ClusterName, server.Node, query.QueryType).Observe(float64(query.WaitingTime))
			if query.PhaseCounts.PrimaryScan > 0 || query.PhaseOperators.PrimaryScan > 0 {
				completedPrimaryIndexUse.WithLabelValues(metrics.ClusterName, query.QueryType).Inc()
			}
		}

		// Vitals report
		completedVitals.WithLabelValues(metrics.ClusterName, server.Node).Set(float64(server.CompletedQueriesCount))
		cpuVitals.WithLabelValues(metrics.ClusterName, server.Node, "user").Set(float64(server.CPUUser))
		cpuVitals.WithLabelValues(metrics.ClusterName, server.Node, "system").Set(float64(server.CPUSystem))
	}
}

func main() {
	flag.Parse()
	monitors := getMonitors()
	renewCounter := 0
	go func() {
		for {
			for ndx := range monitors {
				metrics := monitors[ndx].Execute()
				reportMetrics(&metrics)
			}
			renewCounter++
			if renewCounter%10 == 0 {
				tmpMonitors := getMonitors()
				for ndx, mon := range tmpMonitors {
					if mon.ClusterName == monitors[ndx].ClusterName || monitors[ndx].ClusterName == "" {
						fmt.Printf("Renewing nodes for %s\n", monitors[ndx].ClusterName)
						if monitors[ndx].ClusterName == "" || len(monitors[ndx].Servers) == 0 {
							monitors[ndx] = mon
						} else {
							monitors[ndx].Servers = mon.Servers
						}
					}
				}
				renewCounter = 0
			}
			time.Sleep(15 * time.Second)
		}
	}()
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Serving at %s", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
