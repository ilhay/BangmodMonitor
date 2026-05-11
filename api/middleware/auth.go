package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

	"github.com/bangmodmonitor/api/cache"
	"github.com/gin-gonic/gin"
)

type TokenValidator interface {
	ValidateToken(ctx context.Context, tokenHash string) (hostID, orgID string, ok bool)
}

// Auth validates agent Bearer tokens. When a Redis cache is provided, it is
// checked first before falling back to the DB — keeping validation under 1ms.
func Auth(pg TokenValidator, tc *cache.TokenCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		hash := hashToken(token)

		// Fast path: Redis cache
		hostID, orgID, ok := tc.Get(c.Request.Context(), hash)
		if !ok {
			// Slow path: DB lookup, then populate cache
			hostID, orgID, ok = pg.ValidateToken(c.Request.Context(), hash)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
				return
			}
			tc.Set(c.Request.Context(), hash, hostID, orgID)
		}

		c.Set("host_id", hostID)
		c.Set("org_id", orgID)
		c.Next()
	}
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum)
}
