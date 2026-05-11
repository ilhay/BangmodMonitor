package handler

import (
	"net/http"
	"strconv"

	"github.com/bangmodmonitor/api/storage"
	"github.com/gin-gonic/gin"
)

type MetricsHandler struct {
	ch *storage.CH
}

func NewMetrics(ch *storage.CH) *MetricsHandler {
	return &MetricsHandler{ch: ch}
}

func (h *MetricsHandler) Recent(c *gin.Context) {
	hostID := c.Param("hostId")
	limit := 100
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 1000 {
			limit = n
		}
	}

	points, err := h.ch.QueryRecentMetrics(c.Request.Context(), hostID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"host_id": hostID, "points": points})
}
