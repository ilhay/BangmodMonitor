package handler

import (
	"net/http"
	"time"

	"github.com/bangmodmonitor/api/storage"
	"github.com/gin-gonic/gin"
)

type IngestHandler struct {
	ch *storage.CH
}

func NewIngest(ch *storage.CH) *IngestHandler {
	return &IngestHandler{ch: ch}
}

type ingestPayload struct {
	Metrics agentMetrics `json:"metrics"`
}

type agentMetrics struct {
	Timestamp int64          `json:"timestamp"`
	Hostname  string         `json:"hostname"`
	CPU       cpuMetric      `json:"cpu"`
	Memory    memMetric      `json:"memory"`
	Disks     []diskMetric   `json:"disks"`
	Networks  []netMetric    `json:"networks"`
}

type cpuMetric struct {
	UsagePercent float64 `json:"usage_percent"`
	Cores        int     `json:"cores"`
}
type memMetric struct {
	TotalBytes   uint64  `json:"total_bytes"`
	UsedBytes    uint64  `json:"used_bytes"`
	UsagePercent float64 `json:"usage_percent"`
}
type diskMetric struct {
	Path         string  `json:"path"`
	TotalBytes   uint64  `json:"total_bytes"`
	UsedBytes    uint64  `json:"used_bytes"`
	UsagePercent float64 `json:"usage_percent"`
}
type netMetric struct {
	Interface   string `json:"interface"`
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

func (h *IngestHandler) Handle(c *gin.Context) {
	hostID := c.GetString("host_id")

	var payload ingestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m := payload.Metrics
	ts := time.Unix(m.Timestamp, 0)

	if err := h.ch.InsertAgentMetric(c.Request.Context(), storage.AgentMetricRow{
		Timestamp:  ts,
		HostID:     hostID,
		Hostname:   m.Hostname,
		Region:     "default",
		CPUPercent: float32(m.CPU.UsagePercent),
		CPUCores:   uint8(m.CPU.Cores),
		MemTotal:   m.Memory.TotalBytes,
		MemUsed:    m.Memory.UsedBytes,
		MemPercent: float32(m.Memory.UsagePercent),
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store failed"})
		return
	}

	var diskRows []storage.DiskRow
	for _, d := range m.Disks {
		diskRows = append(diskRows, storage.DiskRow{
			Timestamp: ts, HostID: hostID,
			Path: d.Path, TotalBytes: d.TotalBytes,
			UsedBytes: d.UsedBytes, UsePct: float32(d.UsagePercent),
		})
	}
	if len(diskRows) > 0 {
		_ = h.ch.InsertDiskMetrics(c.Request.Context(), diskRows)
	}

	var netRows []storage.NetRow
	for _, n := range m.Networks {
		netRows = append(netRows, storage.NetRow{
			Timestamp: ts, HostID: hostID,
			Interface: n.Interface, BytesSent: n.BytesSent,
			BytesRecv: n.BytesRecv, PacketsSent: n.PacketsSent,
			PacketsRecv: n.PacketsRecv,
		})
	}
	if len(netRows) > 0 {
		_ = h.ch.InsertNetMetrics(c.Request.Context(), netRows)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
