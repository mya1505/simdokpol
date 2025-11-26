package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter struct untuk menyimpan limiter per IP
type RateLimiter struct {
	ips map[string]*rate.Limiter
	mu  sync.Mutex
	r   rate.Limit
	b   int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		ips: make(map[string]*rate.Limiter),
		r:   r,
		b:   b,
	}
}

func (i *RateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

// GetLimiterMiddleware mengembalikan handler Gin
func (i *RateLimiter) GetLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := i.AddIP(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Terlalu banyak percobaan login. Silakan tunggu beberapa saat.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// LoginRateLimiter: Izinkan 1 request per detik, burst 5 (max 5 percobaan cepat)
// Setelah itu harus nunggu.
var LoginRateLimiter = NewRateLimiter(1, 5)