package sources

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter_Wait(t *testing.T) {
	rl := NewRateLimiter(50 * time.Millisecond)

	start := time.Now()
	ctx := context.Background()

	// First call should be immediate
	if err := rl.Wait(ctx); err != nil {
		t.Fatalf("Wait 1: %v", err)
	}

	// Second call should wait ~50ms
	if err := rl.Wait(ctx); err != nil {
		t.Fatalf("Wait 2: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed < 40*time.Millisecond {
		t.Errorf("elapsed = %v, want >= 40ms", elapsed)
	}
}

func TestRateLimiter_ContextCancel(t *testing.T) {
	rl := NewRateLimiter(5 * time.Second)

	ctx := context.Background()
	_ = rl.Wait(ctx) // first call

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := rl.Wait(ctx)
	if err == nil {
		t.Error("expected context error")
	}
}

func TestRateLimiterPerSecond(t *testing.T) {
	rl := NewRateLimiterPerSecond(10) // 10 rps = 100ms interval
	if rl.Interval() != 100*time.Millisecond {
		t.Errorf("Interval = %v, want 100ms", rl.Interval())
	}
}

func TestRateLimiterPerSecond_Zero(t *testing.T) {
	rl := NewRateLimiterPerSecond(0) // should default to 1 rps
	if rl.Interval() != time.Second {
		t.Errorf("Interval = %v, want 1s", rl.Interval())
	}
}

func TestRateLimiter_IsolatedInstances(t *testing.T) {
	// Verify two rate limiters don't interfere with each other
	rl1 := NewRateLimiter(100 * time.Millisecond)
	rl2 := NewRateLimiter(100 * time.Millisecond)

	ctx := context.Background()

	// Consume rl1's token
	_ = rl1.Wait(ctx)

	// rl2 should be immediate (not blocked by rl1)
	start := time.Now()
	_ = rl2.Wait(ctx)
	elapsed := time.Since(start)

	if elapsed > 20*time.Millisecond {
		t.Errorf("rl2 waited %v — should be independent of rl1", elapsed)
	}
}
