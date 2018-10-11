package n1qlmonitor

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/elfido/n1qlExporter/cbapi"
)

// Monitor N1ql monitoring agent
type Monitor struct {
	ClusterName       string
	Servers           []string
	HTTPAuth          cbapi.Auth
	lastRecordedQuery time.Time
	scrapCount        int64
	protocol          string
	datelayout        string
}

type completedQueryResponse struct {
	ElapsedTime         int64
	ExecutionTime       int64
	WaitingTime         int64
	QueryType           string
	RequestTimeDate     time.Time
	ElapsedTimeString   string `json:"elapsedTime"`
	ExecutionTimeString string `json:"serviceTime"`
	ErrorCount          int    `json:"errorCount"`
	PhaseCounts         struct {
		IndexScan   int `json:"IndexScan"`
		Fetch       int `json:"Fetch"`
		PrimaryScan int `json:"PrimaryScan"`
		Sort        int `json:"Sort"`
	} `json:"phaseCounts"`
	PhaseOperators struct {
		IndexScan   int `json:"IndexScan"`
		Fetch       int `json:"Fetch"`
		PrimaryScan int `json:"PrimaryScan"`
		Sort        int `json:"Sort"`
	} `json:"phaseOperators"`
	RequestTime string `json:"requestTime"`
	ResultCount int    `json:"resultCount"`
	ResultSize  int    `json:"resultSize"`
	Statement   string `json:"statement"`
	State       string `json:"state"`
}

// ActiveQueryResponse from Couchbase
type activeQueryResponse struct {
	ElapsedTimeString   string `json:"elapsedTime"`
	ExecutionTimeString string `json:"executionTime"`
	ScanConsistency     string `json:"scanConsistency"`
	Statement           string `json:"statement"`
	ElapsedTime         int64
	ExecutionTime       int64
	WaitingTime         int64
	QueryType           string
}

type completedQueriesSnapshot struct {
	lastRecordTime time.Time
	completed      []completedQueryResponse
}

// ServerResponse Full server response aggregated
type ServerResponse struct {
	Node                  string
	Active                []activeQueryResponse
	Completed             []completedQueryResponse
	CompletedQueriesCount int64
	CPUUser               float64
	CPUSystem             float64
	lastRecordTime        time.Time
}

// ClusterResponse Collection of server metrics
type ClusterResponse struct {
	ClusterName     string
	ServerResponses []ServerResponse
}

type vitalsResponse struct {
	CompletedCount int64   `json:"request.completed.count"`
	CPUUser        float64 `json:"cpu.user.percent"`
	CPUSystem      float64 `json:"cpu.sys.percent"`
}

func getQueryType(statement string) string {
	components := strings.Split(strings.ToUpper(statement), " ")
	if len(components) > 0 {
		queryType := components[0]
		if len(components) > 1 && components[1] == "RAW" {
			queryType = queryType + "_RAW"
		}
		return queryType
	}
	return "UNK"
}

func getVitalsInformation(server string, serverAuht *cbapi.Auth, c chan vitalsResponse) {
	url := server + "/admin/vitals"
	var serverVitals vitalsResponse
	bytes := cbapi.GetAPI(url, serverAuht)
	err := json.Unmarshal(bytes, &serverVitals)
	if err == nil {
		c <- serverVitals
	} else {
		fmt.Printf("Server: %s\n%s\nError (vitals):\n%s\n", url, string(bytes), err.Error())
		c <- vitalsResponse{}
	}
}

func getActiveQueries(server string, serverAuth *cbapi.Auth, c chan []activeQueryResponse) {
	url := server + "/admin/active_requests"
	var inProgress []activeQueryResponse
	bytes := cbapi.GetAPI(url, serverAuth)
	err := json.Unmarshal(bytes, &inProgress)
	if err == nil {
		for ndx, q := range inProgress {
			inProgress[ndx].ElapsedTime = cbapi.ToMillis(q.ElapsedTimeString)
			inProgress[ndx].ExecutionTime = cbapi.ToMillis(q.ExecutionTimeString)
			inProgress[ndx].WaitingTime = inProgress[ndx].ElapsedTime - inProgress[ndx].ExecutionTime
			inProgress[ndx].QueryType = getQueryType(q.Statement)
			inProgress[ndx].ElapsedTimeString = ""
			inProgress[ndx].ExecutionTimeString = ""
			inProgress[ndx].Statement = ""
		}
		c <- inProgress
	} else {
		fmt.Printf("Server: %s\\nError getting active queries:\n%s\n", url, err.Error())
		c <- []activeQueryResponse{}
	}
}

func getCompletedQueries(server string, serverAuth *cbapi.Auth, lastScrapped time.Time, isFirstRun bool, datelayout string, c chan completedQueriesSnapshot) {
	url := server + "/admin/completed_requests"
	var completed []completedQueryResponse
	bytes := cbapi.GetAPI(url, serverAuth)
	err := json.Unmarshal(bytes, &completed)
	if err == nil {
		completedFiltered := []completedQueryResponse{}
		for ndx, q := range completed {
			completed[ndx].ElapsedTime = cbapi.ToMillis(q.ElapsedTimeString)
			completed[ndx].ExecutionTime = cbapi.ToMillis(q.ExecutionTimeString)
			completed[ndx].WaitingTime = completed[ndx].ElapsedTime - completed[ndx].ExecutionTime
			completed[ndx].QueryType = getQueryType(q.Statement)

			if q.RequestTime != "" {
				parsedDate, dateErr := time.Parse(datelayout, q.RequestTime)
				if dateErr == nil {
					completed[ndx].RequestTimeDate = parsedDate
				} else {
					log.Printf("Error formatting %s %s", q.RequestTime, dateErr)
				}
			}
			if isFirstRun == true || completed[ndx].RequestTimeDate.After(lastScrapped) {
				completed[ndx].Statement = ""
				completed[ndx].ElapsedTimeString = ""
				completed[ndx].ExecutionTimeString = ""
				completedFiltered = append(completedFiltered, completed[ndx])
				lastScrapped = completed[ndx].RequestTimeDate
			}
		}
		c <- completedQueriesSnapshot{
			lastRecordTime: lastScrapped,
			completed:      completedFiltered,
		}
	} else {
		fmt.Printf("Server: %s\nError getting completed queries:\n%s\n", url, err.Error())
		c <- completedQueriesSnapshot{
			lastRecordTime: lastScrapped,
			completed:      []completedQueryResponse{},
		}
	}
}

// should return a channel with a server wrapper
func getServerRecords(url string, serverAuth *cbapi.Auth, lastScrapped time.Time, isFirstRun bool, datelayout string, c chan ServerResponse) {
	activeQueriesChannel := make(chan []activeQueryResponse)
	completedQueriesChannel := make(chan completedQueriesSnapshot)
	vitalsChannel := make(chan vitalsResponse)
	go getActiveQueries(url, serverAuth, activeQueriesChannel)
	go getCompletedQueries(url, serverAuth, lastScrapped, isFirstRun, datelayout, completedQueriesChannel)
	go getVitalsInformation(url, serverAuth, vitalsChannel)
	activeQueries := <-activeQueriesChannel
	completedQueries := <-completedQueriesChannel
	vitalsInformation := <-vitalsChannel
	serverRecord := ServerResponse{
		Active:                activeQueries,
		Completed:             completedQueries.completed,
		lastRecordTime:        completedQueries.lastRecordTime,
		CompletedQueriesCount: vitalsInformation.CompletedCount,
		CPUUser:               vitalsInformation.CPUUser,
		CPUSystem:             vitalsInformation.CPUSystem,
	}
	c <- serverRecord
}

// Execute Retrieves server status
func (m *Monitor) Execute() ClusterResponse {
	serversChannel := make(chan ServerResponse, len(m.Servers))
	if len(m.Servers) > 0 {
		log.Printf("Collecting metrics from cluster %s \n", m.ClusterName)
		lastScrapper := time.Now()
		isFirst := true
		if m.scrapCount > 0 {
			isFirst = false
		}
		for _, s := range m.Servers {
			url := m.protocol + "://" + s + ":8093"
			if m.scrapCount > 0 {
				lastScrapper = m.lastRecordedQuery
			}
			go getServerRecords(url, &m.HTTPAuth, lastScrapper, isFirst, m.datelayout, serversChannel)
		}
		serverResponses := make([]ServerResponse, len(m.Servers), len(m.Servers))
		if len(m.Servers) > 0 {

		}
		for ndx, s := range m.Servers {
			serverRecord := <-serversChannel
			serverRecord.Node = s
			serverResponses[ndx] = serverRecord
			if serverRecord.lastRecordTime.After(m.lastRecordedQuery) {
				m.lastRecordedQuery = serverRecord.lastRecordTime
			}
		}
		m.scrapCount = m.scrapCount + 1
		return ClusterResponse{
			ClusterName:     m.ClusterName,
			ServerResponses: serverResponses,
		}
	}
	log.Printf("Skipping monitor for cluster %s since it has no servers\n", m.ClusterName)
	return ClusterResponse{
		ClusterName:     m.ClusterName,
		ServerResponses: []ServerResponse{},
	}
}

// New creates a new cluster monitor
func New(clusterName string, servers []string, serverAuth cbapi.Auth, useHTTPS bool, datelayout string) Monitor {
	protocol := "http"
	if useHTTPS == true {
		protocol = "https"
	}
	return Monitor{
		ClusterName: clusterName,
		Servers:     servers,
		HTTPAuth:    serverAuth,
		protocol:    protocol,
		datelayout:  datelayout,
	}
}
