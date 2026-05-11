package mq

import (
	"encoding/json"
	"time"
)

const (
	TopicAgentMetrics = "bangmod.agent-metrics"
	TopicProbeResults = "bangmod.probe-results"
)

// AgentMetricMsg is the message envelope published to TopicAgentMetrics.
type AgentMetricMsg struct {
	Timestamp  time.Time `json:"timestamp"`
	HostID     string    `json:"host_id"`
	Hostname   string    `json:"hostname"`
	Region     string    `json:"region"`
	CPUPercent float32   `json:"cpu_percent"`
	CPUCores   uint8     `json:"cpu_cores"`
	MemTotal   uint64    `json:"mem_total"`
	MemUsed    uint64    `json:"mem_used"`
	MemPercent float32   `json:"mem_percent"`
	Disks      []DiskMsg `json:"disks,omitempty"`
	Networks   []NetMsg  `json:"networks,omitempty"`
}

type DiskMsg struct {
	Path       string  `json:"path"`
	TotalBytes uint64  `json:"total_bytes"`
	UsedBytes  uint64  `json:"used_bytes"`
	UsePct     float32 `json:"use_pct"`
}

type NetMsg struct {
	Interface   string `json:"interface"`
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

// ProbeResultMsg is the message envelope published to TopicProbeResults.
type ProbeResultMsg struct {
	Timestamp  time.Time `json:"timestamp"`
	HostID     string    `json:"host_id"`
	TargetURL  string    `json:"target_url"`
	Region     string    `json:"region"`
	StatusCode uint16    `json:"status_code"`
	ResponseMS uint32    `json:"response_ms"`
	IsUp       uint8     `json:"is_up"`
}

func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func Unmarshal[T any](data []byte) (T, error) {
	var v T
	err := json.Unmarshal(data, &v)
	return v, err
}
