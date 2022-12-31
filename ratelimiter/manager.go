package ratelimiter

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
)

type HTTPManager struct {
	// mu protects the lims map.
	mu sync.Mutex
	// lims is a map of IP addresses to rate limiters.
	lims map[string]Limiter
	// ctx is the context for the rate limiter.
	ctx context.Context
	// rate is events per second.
	rate float64
	// burst is the burst size (if supported by Limiter).
	burst int
	// factory is the LimiterFactory to use.
	factory LimiterFactory
}

type LimiterFactory func(rate float64, burst int) Limiter

func NewHTTPManager(ctx context.Context, rate float64, b int, factory LimiterFactory) *HTTPManager {
	return &HTTPManager{
		lims: make(map[string]Limiter),
		ctx:  context.Background(),
	}
}

func (man *HTTPManager) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		limiter := man.getLimiter(ip)

		if err := limiter.Wait(man.ctx); err != nil {
			fmt.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (man *HTTPManager) HandlerFunc(next http.HandlerFunc) http.Handler {
	return man.Handler(http.HandlerFunc(next))
}

func (man *HTTPManager) getLimiter(ip string) Limiter {
	man.mu.Lock()
	defer man.mu.Unlock()
	limiter, ok := man.lims[ip]
	if !ok {
		limiter = man.factory(man.rate, man.burst)
		man.lims[ip] = limiter
	}
	return limiter
}
