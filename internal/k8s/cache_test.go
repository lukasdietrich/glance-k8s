package k8s

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCache_HitWithinTTL(t *testing.T) {
	var c cache[int]

	var calls int32
	fetch := func(context.Context) ([]int, error) {
		atomic.AddInt32(&calls, 1)
		return []int{1, 2, 3}, nil
	}

	for i := 0; i < 5; i++ {
		got, err := c.get(context.Background(), fetch)
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if len(got) != 3 {
			t.Fatalf("call %d: got %d items, want 3", i, len(got))
		}
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("fetch called %d times, want 1", got)
	}
}

func TestCache_RefetchAfterTTL(t *testing.T) {
	var c cache[int]

	var calls int32
	fetch := func(context.Context) ([]int, error) {
		atomic.AddInt32(&calls, 1)
		return []int{1}, nil
	}

	if _, err := c.get(context.Background(), fetch); err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Force expiry by rewinding the expiry timestamp into the past.
	// This avoids the test-runtime cost of waiting for the real TTL
	// and keeps the test deterministic on slow CI.
	c.expiresAt = time.Now().Add(-time.Second)

	if _, err := c.get(context.Background(), fetch); err != nil {
		t.Fatalf("second call: %v", err)
	}

	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("fetch called %d times, want 2", got)
	}
}

func TestCache_ConcurrentCallsDeduped(t *testing.T) {
	var c cache[int]

	var calls int32
	gate := make(chan struct{})
	fetch := func(context.Context) ([]int, error) {
		atomic.AddInt32(&calls, 1)
		<-gate
		return []int{42}, nil
	}

	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			got, err := c.get(context.Background(), fetch)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if len(got) != 1 || got[0] != 42 {
				t.Errorf("unexpected payload: %v", got)
			}
		}()
	}

	// Let all goroutines park inside singleflight.
	time.Sleep(20 * time.Millisecond)
	close(gate)
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("fetch called %d times, want 1", got)
	}
}

func TestCache_ErrorNotCached(t *testing.T) {
	var c cache[int]

	var calls int32
	wantErr := errors.New("boom")
	fetch := func(context.Context) ([]int, error) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			return nil, wantErr
		}
		return []int{7}, nil
	}

	if _, err := c.get(context.Background(), fetch); !errors.Is(err, wantErr) {
		t.Fatalf("first call err = %v, want %v", err, wantErr)
	}

	got, err := c.get(context.Background(), fetch)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if len(got) != 1 || got[0] != 7 {
		t.Fatalf("second call payload = %v, want [7]", got)
	}

	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("fetch called %d times, want 2", got)
	}
}

func TestCache_TypedSeparately(t *testing.T) {
	// Two cache instances of different element types should not
	// interact; this is more a compile-time check than a runtime
	// one, but the test exercises both code paths.
	var ints cache[int]
	var strs cache[string]

	gotInts, err := ints.get(context.Background(), func(context.Context) ([]int, error) {
		return []int{1, 2}, nil
	})
	if err != nil || len(gotInts) != 2 {
		t.Fatalf("ints: got %v err %v", gotInts, err)
	}

	gotStrs, err := strs.get(context.Background(), func(context.Context) ([]string, error) {
		return []string{"a"}, nil
	})
	if err != nil || len(gotStrs) != 1 {
		t.Fatalf("strs: got %v err %v", gotStrs, err)
	}
}
