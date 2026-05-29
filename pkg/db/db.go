package db

import (
	"context"
	"fmt"
	"time"
)

func (p *Pool) HealthCheck(ctx context.Context) error {
	return p.Ping(ctx)
}

func retryWithBackoff(ctx context.Context, fn func(context.Context) error, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if err = fn(ctx); err == nil {
			return nil
		}
		if i < maxRetries-1 {
			wait := time.Duration(1<<uint(i)) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}
	}
	return fmt.Errorf("retry failed after %d attempts: %w", maxRetries, err)
}
