package provider

import (
	"context"
	"fmt"
	"time"
)

func waitForCondition(ctx context.Context, timeout time.Duration, interval time.Duration, check func(context.Context) (bool, error)) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		ok, err := check(ctx)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for async operation")
		case <-ticker.C:
		}
	}
}
