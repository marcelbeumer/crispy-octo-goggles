package ratelimiter

import (
	"context"

	"golang.org/x/time/rate"
)

type XtimeLimiter struct {
	lim *rate.Limiter
}

func (x *XtimeLimiter) Wait(ctx context.Context) error {
	return x.lim.Wait(ctx)
}

func NewXtimeLimiter(r float64, b int) *XtimeLimiter {
	return &XtimeLimiter{
		lim: rate.NewLimiter(rate.Limit(r), b),
	}
}

func XtimeLimiterFactory(rate float64, burst int) Limiter {
	return NewXtimeLimiter(rate, burst)
}
