package middleware

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type TokenValidator interface {
	ValidateToken(ctx interface{ Value(interface{}) interface{} }, tokenHash string) (hostID string, ok bool)
}

type contextKey string

const HostIDKey contextKey = "host_id"

func Auth(pg interface {
	ValidateToken(ctx interface{ Value(interface{}) interface{} }, tokenHash string) (string, bool)
}) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		hash := hashToken(token)

		hostID, ok := pg.ValidateToken(c.Request.Context(), hash)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("host_id", hostID)
		c.Next()
	}
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum)
}
