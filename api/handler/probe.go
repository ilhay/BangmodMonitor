package handler

import (
	"net/http"
	"time"

	"github.com/bangmodmonitor/api/storage"
	"github.com/gin-gonic/gin"
)

type ProbeHandler struct {
	ch         *storage.CH
	nodeSecret string
}

func NewProbe(ch *storage.CH, nodeSecret string) *ProbeHandler {
	return &ProbeHandler{ch: ch, nodeSecret: nodeSecret}
}

type probePayload struct {
	Region  string        `json:"region" binding:"required"`
	Results []probeResult `json:"results" binding:"required"`
}

type probeResult struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	ResponseMS int64  `json:"response_ms"`
	IsUp       bool   `json:"is_up"`
	Error      string `json:"error"`
}

func (h *ProbeHandler) Ingest(c *gin.Context) {
	if h.nodeSecret != "" && c.GetHeader("X-Node-Secret") != h.nodeSecret {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid node secret"})
		return
	}

	var payload probePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ts := time.Now()
	for _, r := range payload.Results {
		isUp := uint8(0)
		if r.IsUp {
			isUp = 1
		}
		_ = h.ch.InsertProbeResult(c.Request.Context(), storage.ProbeRow{
			Timestamp:  ts,
			HostID:     "probe",
			TargetURL:  r.URL,
			Region:     payload.Region,
			StatusCode: uint16(r.StatusCode),
			ResponseMS: uint32(r.ResponseMS),
			IsUp:       isUp,
		})
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "stored": len(payload.Results)})
}

type ProbePoint struct {
	Timestamp  time.Time `json:"timestamp"`
	TargetURL  string    `json:"url"`
	Region     string    `json:"region"`
	StatusCode uint16    `json:"status_code"`
	ResponseMS uint32    `json:"response_ms"`
	IsUp       bool      `json:"is_up"`
}

func (h *ProbeHandler) Recent(c *gin.Context) {
	url := c.Query("url")
	region := c.Query("region")

	results, err := h.ch.QueryProbeResults(c.Request.Context(), url, region, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}
