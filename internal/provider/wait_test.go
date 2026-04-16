package provider

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWaitForCondition_SucceedsImmediately(t *testing.T) {
	calls := 0
	err := waitForCondition(context.Background(), 5*time.Second, 10*time.Millisecond, func(_ context.Context) (bool, error) {
		calls++
		return true, nil
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestWaitForCondition_SucceedsAfterRetries(t *testing.T) {
	calls := 0
	err := waitForCondition(context.Background(), 5*time.Second, 10*time.Millisecond, func(_ context.Context) (bool, error) {
		calls++
		return calls >= 3, nil
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestWaitForCondition_TimesOut(t *testing.T) {
	err := waitForCondition(context.Background(), 50*time.Millisecond, 10*time.Millisecond, func(_ context.Context) (bool, error) {
		return false, nil
	})
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
}

func TestWaitForCondition_PropagatesCheckError(t *testing.T) {
	sentinel := errors.New("check failed")
	calls := 0
	err := waitForCondition(context.Background(), 5*time.Second, 10*time.Millisecond, func(_ context.Context) (bool, error) {
		calls++
		return false, sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected error to stop retrying after 1 call, got %d", calls)
	}
}
