package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitByIP allows at most `limit` requests per client IP inside a
// sliding window. It protects the anonymous upload endpoint from abuse
// without requiring accounts. State is in-memory, which is fine for a
// single dashboard instance.
func RateLimitByIP(limit int, window time.Duration) gin.HandlerFunc {

	type bucket struct {
		hits []time.Time
	}

	var mu sync.Mutex
	buckets := map[string]*bucket{}

	return func(ctx *gin.Context) {
		now := time.Now()
		ip := ctx.ClientIP()

		mu.Lock()

		b, ok := buckets[ip]
		if !ok {
			b = &bucket{}
			buckets[ip] = b
		}

		cutoff := now.Add(-window)
		kept := b.hits[:0]
		for _, t := range b.hits {
			if t.After(cutoff) {
				kept = append(kept, t)
			}
		}
		b.hits = kept

		if len(b.hits) >= limit {
			mu.Unlock()
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many uploads from this address, try again later",
			})
			return
		}

		b.hits = append(b.hits, now)

		// Opportunistic cleanup so idle IPs don't accumulate forever.
		if len(buckets) > 10000 {
			for k, v := range buckets {
				if len(v.hits) == 0 || v.hits[len(v.hits)-1].Before(cutoff) {
					delete(buckets, k)
				}
			}
		}

		mu.Unlock()
		ctx.Next()
	}
}
