package handler

import (
	"net/http"

	"github.com/bangmodmonitor/api/storage"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	maria *storage.Maria
}

func NewAdmin(maria *storage.Maria) *AdminHandler {
	return &AdminHandler{maria: maria}
}

func (h *AdminHandler) ListOrgs(c *gin.Context) {
	if c.GetString("role") != "superadmin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "superadmin only"})
		return
	}
	orgs, err := h.maria.ListAllOrgs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if orgs == nil {
		orgs = []storage.OrgSummary{}
	}
	c.JSON(http.StatusOK, gin.H{"orgs": orgs})
}

func (h *AdminHandler) Stats(c *gin.Context) {
	if c.GetString("role") != "superadmin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "superadmin only"})
		return
	}
	totalCents, paidOrgs, err := h.maria.GetRevenueStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total_revenue_cents": totalCents,
		"paid_orgs":           paidOrgs,
	})
}
