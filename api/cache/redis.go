// Package cache provides a Redis-backed token validation cache.
// On a cache hit, token validation takes <1ms instead of a DB round-trip.
// On token revoke or rotate, the cache entry is deleted immediately so the
// old token stops working within milliseconds.
package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const tokenTTL = 5 * time.Minute

type TokenCache struct {
	rdb     *redis.Client
	enabled bool
}

func NewTokenCache(addr string) *TokenCache {
	if addr == "" {
		return &TokenCache{enabled: false}
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	})
	return &TokenCache{rdb: rdb, enabled: true}
}

func (c *TokenCache) Enabled() bool { return c.enabled }

func (c *TokenCache) Ping(ctx context.Context) error {
	if !c.enabled {
		return nil
	}
	return c.rdb.Ping(ctx).Err()
}

// Get returns (hostID, orgID, true) if the token hash is cached.
func (c *TokenCache) Get(ctx context.Context, tokenHash string) (hostID, orgID string, ok bool) {
	if !c.enabled {
		return "", "", false
	}
	val, err := c.rdb.Get(ctx, tokenKey(tokenHash)).Result()
	if err != nil {
		return "", "", false
	}
	parts := strings.SplitN(val, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// Set stores a token → (hostID, orgID) mapping with TTL.
func (c *TokenCache) Set(ctx context.Context, tokenHash, hostID, orgID string) {
	if !c.enabled {
		return
	}
	_ = c.rdb.SetEx(ctx, tokenKey(tokenHash), fmt.Sprintf("%s:%s", hostID, orgID), tokenTTL).Err()
}

// Delete removes a token from the cache immediately (call on revoke/rotate).
func (c *TokenCache) Delete(ctx context.Context, tokenHash string) {
	if !c.enabled {
		return
	}
	_ = c.rdb.Del(ctx, tokenKey(tokenHash)).Err()
}

func tokenKey(hash string) string {
	return "token:" + hash
}
