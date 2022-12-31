package ratelimiter

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// BasicLimiter is a rate limiter based on the x/time/rate package.
type BasicLimiter struct {
	// mu protects the lims map.
	mu sync.RWMutex
	// lims is a map of IP addresses to rate limiters.
	lims map[string]*rate.Limiter
	// ctx is the context for the rate limiter.
	ctx context.Context
	// r is the rate at which requests are allowed.
	r rate.Limit
	// b is the burst size.
	b int
}

func (lim *BasicLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		limiter := lim.getLimiter(ip)

		if err := limiter.Wait(lim.ctx); err != nil {
			fmt.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (lim *BasicLimiter) getLimiter(ip string) *rate.Limiter {
	lim.mu.RLock() // <<------ how was it again with the mutexes?
	limiter, ok := lim.lims[ip]
	if !ok {
		limiter = rate.NewLimiter(lim.r, lim.b)
		lim.lims[ip] = limiter
	}
	return limiter
}

func NewRateLimiterXtime(interval time.Duration, b int, ctx context.Context) *BasicLimiter {
	if ctx == nil {
		ctx = context.Background()
	}
	return &BasicLimiter{
		lims: make(map[string]*rate.Limiter),
		ctx:  ctx,
		r:    rate.Every(interval),
		b:    b,
	}
}
