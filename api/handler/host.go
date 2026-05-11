package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/bangmodmonitor/api/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type HostHandler struct {
	maria *storage.Maria
	ch    *storage.CH
}

func NewHost(maria *storage.Maria, ch *storage.CH) *HostHandler {
	return &HostHandler{maria: maria, ch: ch}
}

func (h *HostHandler) List(c *gin.Context) {
	orgID := c.GetString("org_id")
	hosts, err := h.maria.ListHosts(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if hosts == nil {
		hosts = []storage.Host{}
	}
	c.JSON(http.StatusOK, gin.H{"hosts": hosts})
}

type createHostReq struct {
	Name   string `json:"name" binding:"required"`
	Region string `json:"region" binding:"required"`
}

func (h *HostHandler) Create(c *gin.Context) {
	orgID := c.GetString("org_id")
	var req createHostReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enforce plan host limit
	sub, _ := h.maria.GetSubscription(c.Request.Context(), orgID)
	if sub != nil {
		plan, _ := h.maria.GetPlan(c.Request.Context(), sub.PlanID)
		if plan != nil {
			hostCount, _ := h.maria.CountActiveHosts(c.Request.Context(), orgID)
			if hostCount >= plan.HostLimit {
				c.JSON(http.StatusPaymentRequired, gin.H{
					"error":      "host limit reached for your plan",
					"limit":      plan.HostLimit,
					"current":    hostCount,
					"upgrade_to": "starter",
				})
				return
			}
		}
	}

	plainToken, tokenHash, err := generateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	hostID := uuid.New().String()
	if err := h.maria.CreateHost(c.Request.Context(), hostID, orgID, req.Name, tokenHash, req.Region); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"host": storage.Host{
			ID:     hostID,
			OrgID:  orgID,
			Name:   req.Name,
			Region: req.Region,
		},
		// Plain token shown exactly ONCE — customer must save it now
		"token": plainToken,
		"install_linux":   fmt.Sprintf("curl -fsSL https://your-domain.com/install.sh | bash -s -- --token=%s --region=%s", plainToken, req.Region),
		"install_windows": fmt.Sprintf("irm https://your-domain.com/install.ps1 | iex -Token %s -Region %s", plainToken, req.Region),
	})
}

func (h *HostHandler) Delete(c *gin.Context) {
	orgID := c.GetString("org_id")
	hostID := c.Param("id")

	if err := h.maria.DeleteHost(c.Request.Context(), hostID, orgID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *HostHandler) RotateToken(c *gin.Context) {
	orgID := c.GetString("org_id")
	hostID := c.Param("id")

	plainToken, tokenHash, err := generateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	if err := h.maria.RotateToken(c.Request.Context(), hostID, orgID, tokenHash); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": plainToken,
		"note":  "Old token is immediately revoked. Save this token — it will not be shown again.",
	})
}

func (h *HostHandler) Metrics(c *gin.Context) {
	orgID := c.GetString("org_id")
	hostID := c.Param("id")

	// Verify ownership
	host, err := h.maria.GetHost(c.Request.Context(), hostID, orgID)
	if err != nil || host == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
		return
	}

	points, err := h.ch.QueryRecentMetrics(c.Request.Context(), hostID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"host_id": hostID, "points": points})
}

func generateToken() (plain, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	plain = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(plain))
	hash = fmt.Sprintf("%x", sum)
	return
}
