package api

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCachedList_HitWithinTTL(t *testing.T) {
	cache := newListCache(time.Minute)

	var calls int32
	fetch := func(context.Context) ([]int, error) {
		atomic.AddInt32(&calls, 1)
		return []int{1, 2, 3}, nil
	}

	for i := 0; i < 5; i++ {
		got, err := cachedList(context.Background(), cache, "k", fetch)
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

func TestCachedList_RefetchAfterTTL(t *testing.T) {
	cache := newListCache(20 * time.Millisecond)

	var calls int32
	fetch := func(context.Context) ([]int, error) {
		atomic.AddInt32(&calls, 1)
		return []int{1}, nil
	}

	if _, err := cachedList(context.Background(), cache, "k", fetch); err != nil {
		t.Fatalf("first call: %v", err)
	}

	time.Sleep(40 * time.Millisecond)

	if _, err := cachedList(context.Background(), cache, "k", fetch); err != nil {
		t.Fatalf("second call: %v", err)
	}

	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("fetch called %d times, want 2", got)
	}
}

func TestCachedList_ConcurrentCallsDeduped(t *testing.T) {
	cache := newListCache(time.Minute)

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
			got, err := cachedList(context.Background(), cache, "k", fetch)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if len(got) != 1 || got[0] != 42 {
				t.Errorf("unexpected payload: %v", got)
			}
		}()
	}

	// Give all goroutines a chance to enter the singleflight wait.
	time.Sleep(20 * time.Millisecond)
	close(gate)
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("fetch called %d times, want 1", got)
	}
}

func TestCachedList_ErrorNotCached(t *testing.T) {
	cache := newListCache(time.Minute)

	var calls int32
	wantErr := errors.New("boom")
	fetch := func(context.Context) ([]int, error) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			return nil, wantErr
		}
		return []int{7}, nil
	}

	if _, err := cachedList(context.Background(), cache, "k", fetch); !errors.Is(err, wantErr) {
		t.Fatalf("first call err = %v, want %v", err, wantErr)
	}

	got, err := cachedList(context.Background(), cache, "k", fetch)
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

func TestCachedList_TTLZeroBypasses(t *testing.T) {
	cache := newListCache(0)

	var calls int32
	fetch := func(context.Context) ([]int, error) {
		atomic.AddInt32(&calls, 1)
		return []int{1}, nil
	}

	for i := 0; i < 3; i++ {
		if _, err := cachedList(context.Background(), cache, "k", fetch); err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
	}

	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("fetch called %d times, want 3", got)
	}
}

func TestCachedList_NilCacheBypasses(t *testing.T) {
	var calls int32
	fetch := func(context.Context) ([]int, error) {
		atomic.AddInt32(&calls, 1)
		return []int{1}, nil
	}

	for i := 0; i < 3; i++ {
		if _, err := cachedList[int](context.Background(), nil, "k", fetch); err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
	}

	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("fetch called %d times, want 3", got)
	}
}

func TestCachedList_KeysAreIndependent(t *testing.T) {
	cache := newListCache(time.Minute)

	var aCalls, bCalls int32
	aFetch := func(context.Context) ([]int, error) {
		atomic.AddInt32(&aCalls, 1)
		return []int{1}, nil
	}
	bFetch := func(context.Context) ([]string, error) {
		atomic.AddInt32(&bCalls, 1)
		return []string{"x"}, nil
	}

	for i := 0; i < 3; i++ {
		if _, err := cachedList(context.Background(), cache, "a", aFetch); err != nil {
			t.Fatalf("a call %d: %v", i, err)
		}
		if _, err := cachedList(context.Background(), cache, "b", bFetch); err != nil {
			t.Fatalf("b call %d: %v", i, err)
		}
	}

	if got := atomic.LoadInt32(&aCalls); got != 1 {
		t.Fatalf("a fetch called %d times, want 1", got)
	}
	if got := atomic.LoadInt32(&bCalls); got != 1 {
		t.Fatalf("b fetch called %d times, want 1", got)
	}
}

func TestReadCacheTTL(t *testing.T) {
	cases := []struct {
		name string
		env  string
		want time.Duration
	}{
		{name: "unset", env: "", want: defaultCacheTTL},
		{name: "valid_5s", env: "5s", want: 5 * time.Second},
		{name: "valid_zero", env: "0s", want: 0},
		{name: "valid_500ms", env: "500ms", want: 500 * time.Millisecond},
		{name: "invalid", env: "not-a-duration", want: defaultCacheTTL},
		{name: "negative", env: "-1s", want: defaultCacheTTL},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// readCacheTTL treats an empty env var the same as unset
			// (os.Getenv returns "" for both), so t.Setenv("", ...)
			// covers the unset case.
			t.Setenv(envCacheTTL, tc.env)

			got := readCacheTTL()
			if got != tc.want {
				t.Fatalf("readCacheTTL() = %v, want %v", got, tc.want)
			}
		})
	}
}
