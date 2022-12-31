package ratelimiter

import (
	"context"

	uber "go.uber.org/ratelimit"
)

type UberLimiter struct {
	lim uber.Limiter
}

func (x *UberLimiter) Wait(ctx context.Context) error {
	taken := make(chan struct{})
	go func() {
		x.lim.Take()
		close(taken)
	}()

	select {
	case <-taken:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func NewUberLimiter(r float64, b int) *UberLimiter {
	return &UberLimiter{
		lim: uber.New(int(r)),
	}
}

func UberLimiterFactory(rate float64, burst int) Limiter {
	return NewUberLimiter(rate, burst)
}
