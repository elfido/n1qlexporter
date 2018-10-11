package datamonitor

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elfido/n1qlExporter/cbapi"
)

// Monitor for data nodes
type Monitor struct {
}

// ClusterMap Couchbase cluster summary
type ClusterMap struct {
	Name       string
	QueryNodes []string
	IndexNodes []string
	DataNodes  []string
	TotalNodes int // Different since nodes can share roles
	Version    string
	Buckets    []string // Do I really need this?
}

type couchbaseDefaultResponse struct {
	Name  string `json:"name"`
	Nodes []struct {
		Hostname string   `json:"hostname"`
		Services []string `json:"services"`
		Version  string   `json:"version"`
	} `json:"nodes"`
	SystemStats struct {
		CPUUtilization float64 `json:"cpu_utilization_rate"`
	}
	MemoryTotal int64 `json:"memoryTotal"`
	MemoryFree  int64 `json:"memoryFree"`
	CPUCount    int   `json:"cpuCount"`
}

// Execute calls the monitoring APIs in data nodes
func (m *Monitor) Execute() {
	// url := "http://${user}:${password}@${clusterAddress}:8091/pools/default"

}

// NewDataMonitor Creates a data node monitor
func NewDataMonitor(clusterName string, servers []string, serverAuth cbapi.Auth, useHTTPS bool) {

}

// GetClusterMap Discovers the nodes of a Couchbase cluster
func GetClusterMap(server string, auth cbapi.Auth) (ClusterMap, error) {
	fmt.Printf("Looking for new nodes for cluster %s\n", server)
	url := server + ":8091/pools/default"
	var response couchbaseDefaultResponse
	version := ""
	bytes := cbapi.GetAPI(url, &auth)
	err := json.Unmarshal(bytes, &response)
	if err == nil {
		kvNodes := make([]string, 0, 0)
		n1qlNodes := make([]string, 0, 0)
		indexNodes := make([]string, 0, 0)
		for _, node := range response.Nodes {
			for _, service := range node.Services {
				simpleHostName := strings.Replace(node.Hostname, ":8091", "", -1)
				if service == "kv" {
					kvNodes = append(kvNodes, simpleHostName)
				}
				if service == "n1ql" {
					n1qlNodes = append(n1qlNodes, simpleHostName)
				}
				if service == "index" {
					indexNodes = append(indexNodes, simpleHostName)
				}
			}
			version = node.Version
		}
		return ClusterMap{
			Name:       response.Name,
			TotalNodes: len(response.Nodes),
			QueryNodes: n1qlNodes,
			DataNodes:  kvNodes,
			Version:    version,
			IndexNodes: indexNodes,
		}, nil
	}
	return ClusterMap{}, err
}
