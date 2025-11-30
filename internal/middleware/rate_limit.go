package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Wrapper untuk menyimpan waktu akses terakhir
type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	ips map[string]*client
	mu  sync.Mutex
	r   rate.Limit
	b   int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	rl := &RateLimiter{
		ips: make(map[string]*client),
		r:   r,
		b:   b,
	}
	// Jalankan Garbage Collector di background
	go rl.cleanupLoop()
	return rl
}

func (i *RateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = &client{
			limiter:  rate.NewLimiter(i.r, i.b),
			lastSeen: time.Now(),
		}
		i.ips[ip] = limiter
		return limiter.limiter
	}

	limiter.lastSeen = time.Now()
	return limiter.limiter
}

// Hapus IP yang tidak aktif lebih dari 3 menit untuk membebaskan memori
func (i *RateLimiter) cleanupLoop() {
	for {
		time.Sleep(1 * time.Minute)
		i.mu.Lock()
		for ip, client := range i.ips {
			if time.Since(client.lastSeen) > 3*time.Minute {
				delete(i.ips, ip)
			}
		}
		i.mu.Unlock()
	}
}

func (i *RateLimiter) GetLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := i.AddIP(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Terlalu banyak percobaan. Silakan tunggu sebentar.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

var LoginRateLimiter = NewRateLimiter(1, 5)