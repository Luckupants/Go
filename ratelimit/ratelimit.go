//go:build !solution

package ratelimit

import (
	"context"
	"errors"
	"time"
)

// Limiter is precise rate limiter with context support.
type Limiter struct {
	stop      bool
	stopped   chan struct{}
	free      chan struct{}
	sleepTime time.Duration
}

var ErrStopped = errors.New("limiter stopped")

// NewLimiter returns limiter that throttles rate of successful Acquire() calls
// to maxSize events at any given interval.
func NewLimiter(maxCount int, interval time.Duration) *Limiter {
	answer := Limiter{stop: false, stopped: make(chan struct{}), free: make(chan struct{}, maxCount), sleepTime: interval}
	for i := 0; i < maxCount; i++ {
		answer.free <- struct{}{}
	}
	return &answer
}

func (l *Limiter) Acquire(ctx context.Context) error {
	if l.stop {
		return ErrStopped
	}
	select {
	case <-l.free:
		go func() {
			if l.sleepTime != 0 {
				ticker := time.NewTicker(l.sleepTime)
				defer ticker.Stop()
				select {
				case <-ticker.C:
				case <-l.stopped:
				}
			}
			l.free <- struct{}{}
		}()
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func (l *Limiter) Stop() {
	l.stop = true
	close(l.stopped)
}
