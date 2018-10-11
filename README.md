# Prometheus N1QL Exporter

Exposes Prometheus statistics from Couchbase's N1QL monitoring APIs. Enter as many servers as you want as shown in the example:

To start a container:
```sh
docker run --rm --name test -p 8380:8380 -e CLUSTERS="{\"myserver\":\"my_ip\"}" -e HTTPUSER=couchbase_user -e="HTTPPASSWORD=couchbase_password" n1qlexporter:1.0.0
```

The configuration is done via environment variable or by mapping a JSON file to /app/settings.json in the following format:

```json
{
	"httpuser": "",
	"httppassword": "",
	"usehttps": "false",
	"clusters": {
		"myClusterName": "localhost"
	}
}
```

You just need to provide a single host name and the exporter will discover all query nodes.

The following metrics are exposed:

| Metric name | Metric type | Description |
|----------|------|------|
| n1ql_active_time_execution| Histogram | Active queries current time execution per cluster/node/query type |
| n1ql_active_accumulated_queries| Histogram | Active queries running per node/cluster |
| n1ql_active_time_waiting| Histogram | Active queries waiting time for execution per node/cluster/query type |
| n1ql_active_consistency| Counter | Active queries count per cluster/scan consistency |
| n1ql_completed_result_count| Histogram | Completed (usually slow) queries count per cluster/query type |
| n1ql_completed_result_size| Histogram | Completed (usually slow) queries results size (in bytes) per cluster/query type |
| n1ql_completed_time_execution| Histogram | Completed (usually slow) queries response time per cluster/node/query type/status |
| n1ql_completed_time_waiting| Histogram | Completed (usually slow) queries waiting time for execution per cluster/node/query type |
| n1ql_completed_primaryindex| Counter | Completed (usually slow) queries using primary index scan per cluster/query type |
| n1ql_vitals_completed_queries| Gauge | Executed queries by cluster/node |
| n1ql_vitals_cpu_usage| Gauge | current CPU required by cluster/node/space (space: system or user) |



Todo:

- Handle configuration errors
- Obtain version per node

