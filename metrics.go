package main

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"
	"strings"
)

var (
	// General metrics
	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Count of all HTTP requests",
	}, []string{"code", "method"})

	// Replikator/Replication metrics
	replikatorReplicationLag = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "replikator_replication_lag",
		Help: "Replication lag from master server",
	}, []string{"state"})

	replikatorReplicationDiskUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "replikator_replication_disk_usage",
		Help: "Disk usage by the replication process",
	}, []string{"state"})

	replikatorDiskCapacity = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "replikator_disk_capacity",
		Help: "Disk capacity",
	})

	replikatorDiskFree = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "replikator_disk_free",
		Help: "Free disk",
	})

	replikatorMemoryCapacity = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "replikator_memory_capacity",
		Help: "Memory capacity",
	})

	replikatorMemoryFree = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "replikator_memory_free",
		Help: "Free memory",
	})

	// Replicas metrics
	replikatorReplicaDiskUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "replikator_replica_disk_usage",
		Help: "Disk usage by a replica",
	}, []string{"replica", "state"})

	replikatorReplicaMemoryAllocated = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "replikator_replica_memory_allocated",
		Help: "Memory allocated for a replica",
	}, []string{"replica", "state"})

	replikatorReplicaMemoryUsed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "replikator_replica_memory_used",
		Help: "Memory used by a replica",
	}, []string{"replica", "state"})
)

type replikatorData struct {
	DatabaseGlobalState databaseGlobalState `json:"DatabaseGlobalState"`
}

type databaseGlobalState struct {
	DatabaseInstanceState []databaseInstanceState `json:"DatabaseInstanceState"`
	ReplicationState      string                  `json:"eReplicationState"`
	ReplicationLag        string                  `json:"iReplicationLag"`
	ReplicationDiskUsage  string                  `json:"sAllocatedForDb"`
	DiskCapacity          string                  `json:"sTotalStorageCapacity"`
	DiskFree              string                  `json:"sFree"`
	MemoryCapacity        string                  `json:"sTotalMemCapacity"`
	MemoryFree            string                  `json:"sFreeMem"`
}

type databaseInstanceState struct {
	DatabaseProperties databaseProperties `json:"DatabaseProperties"`
	State              string             `json:"eState"`
	DiskUsage          string             `json:"sSizeTotal"`
	MemoryAllocated    string             `json:"sMemAllocated"`
	MemoryUsed         string             `json:"sMemUsed"`
}

type databaseProperties struct {
	InstanceId string `json:"sInstanceId"`
}

func registerMetrics() {
	prometheus.MustRegister(
		// General metrics
		httpRequestsTotal,

		// Replikator/Replication metrics
		replikatorReplicationLag,
		replikatorReplicationDiskUsage,
		replikatorDiskCapacity,
		replikatorDiskFree,
		replikatorMemoryCapacity,
		replikatorMemoryFree,

		// Replicas metrics
		replikatorReplicaDiskUsage,
		replikatorReplicaMemoryAllocated,
		replikatorReplicaMemoryUsed,
	)
}

func getMetrics() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data replikatorData

		output := execute("--output json --list")
		json.Unmarshal([]byte(output), &data)

		labels := prometheus.Labels{
			"state": strings.ToLower(data.DatabaseGlobalState.ReplicationState),
		}

		replicationLag, err := strconv.ParseFloat(data.DatabaseGlobalState.ReplicationLag, 64)
		if err != nil {
			replicationLag = -1
		}
		replikatorReplicationLag.Reset()
		replikatorReplicationLag.With(labels).Set(replicationLag)

		replicationDiskUsage, err := strconv.ParseFloat(data.DatabaseGlobalState.ReplicationDiskUsage, 64)
		if err != nil {
			replicationDiskUsage = 0
		}
		replikatorReplicationDiskUsage.Reset()
		replikatorReplicationDiskUsage.With(labels).Set(replicationDiskUsage)

		diskCapacity, _ := strconv.ParseFloat(data.DatabaseGlobalState.DiskCapacity, 64)
		replikatorDiskCapacity.Set(diskCapacity)

		diskFree, _ := strconv.ParseFloat(data.DatabaseGlobalState.DiskFree, 64)
		replikatorDiskFree.Set(diskFree)

		memoryCapacity, _ := strconv.ParseFloat(data.DatabaseGlobalState.MemoryCapacity, 64)
		replikatorMemoryCapacity.Set(memoryCapacity)

		memoryFree, _ := strconv.ParseFloat(data.DatabaseGlobalState.MemoryFree, 64)
		replikatorMemoryFree.Set(memoryFree)

		replikatorReplicaDiskUsage.Reset()
		replikatorReplicaMemoryAllocated.Reset()
		replikatorReplicaMemoryUsed.Reset()

		for _, replikator := range data.DatabaseGlobalState.DatabaseInstanceState {
			labels := prometheus.Labels{
				"replica": replikator.DatabaseProperties.InstanceId,
				"state":   strings.ToLower(replikator.State),
			}

			diskUsage, err := strconv.ParseFloat(replikator.DiskUsage, 64)
			if err != nil {
				diskUsage = 0
			}

			memoryAllocated, err := strconv.ParseFloat(replikator.MemoryAllocated, 64)
			if err != nil {
				memoryAllocated = 0
			}

			memoryUsed, err := strconv.ParseFloat(replikator.MemoryUsed, 64)
			if err != nil {
				memoryUsed = 0
			}

			replikatorReplicaDiskUsage.With(labels).Set(diskUsage)
			replikatorReplicaMemoryAllocated.With(labels).Set(memoryAllocated)
			replikatorReplicaMemoryUsed.With(labels).Set(memoryUsed)
		}

		promhttp.Handler().ServeHTTP(w, r)
	})
}
