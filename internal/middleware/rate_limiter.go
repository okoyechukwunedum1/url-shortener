package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter limits how many requests a client can make
// It uses a token bucket algorithm (from golang.org/x/time/rate)
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     rate.Limit    // requests per second
	burst    int           // maximum burst size
	cleanup  time.Duration // how often to clean up old visitors
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
// Default: 10 requests per second with a burst of 20
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     10, // 10 requests per second
		burst:    20, // allow bursts up to 20
		cleanup:  1 * time.Minute,
	}

	// Start a background goroutine to clean up old visitors
	go rl.cleanupVisitors()

	return rl
}

// Limit returns a Gin middleware function that enforces rate limiting
// It identifies clients by their IP address
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		limiter := rl.getVisitor(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, please slow down",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getVisitor returns the rate limiter for a specific IP
// creating a new one if it doesn't exist
func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.RLock()
	v, exists := rl.visitors[ip]
	rl.mu.RUnlock()

	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)

		rl.mu.Lock()
		// Double-check after acquiring write lock
		v, exists = rl.visitors[ip]
		if !exists {
			v = &visitor{limiter: limiter, lastSeen: time.Now()}
			rl.visitors[ip] = v
		}
		rl.mu.Unlock()
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupVisitors removes visitors who haven't been seen recently
// This prevents the map from growing forever
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(rl.cleanup)

		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 5*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}
