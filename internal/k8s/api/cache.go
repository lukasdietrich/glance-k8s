package api

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const (
	envCacheTTL     = "GLANCE_KUBE_CACHE_TTL"
	defaultCacheTTL = 5 * time.Second
)

// listCache is a process-local, read-through cache for cluster-wide
// List() results. Entries expire after a fixed TTL and concurrent
// callers for the same key are deduplicated onto a single in-flight
// fetch via singleflight. Errors are not cached so a transient apiserver
// failure does not lock the cache for the full TTL.
//
// The cached value is the returned slice; callers must not mutate it.
type listCache struct {
	ttl    time.Duration
	mu     sync.RWMutex
	values map[string]cacheEntry
	flight singleflight.Group
}

type cacheEntry struct {
	value     any
	expiresAt time.Time
}

func newListCache(ttl time.Duration) *listCache {
	return &listCache{
		ttl:    ttl,
		values: map[string]cacheEntry{},
	}
}

// cachedList returns items from the cache when fresh; otherwise it
// invokes fetch and caches the result. When ttl is non-positive the
// cache is bypassed and fetch runs on every call.
func cachedList[Item any](ctx context.Context, c *listCache, key string, fetch func(context.Context) ([]Item, error)) ([]Item, error) {
	if c == nil || c.ttl <= 0 {
		return fetch(ctx)
	}

	c.mu.RLock()
	entry, ok := c.values[key]
	c.mu.RUnlock()
	if ok && time.Now().Before(entry.expiresAt) {
		return entry.value.([]Item), nil
	}

	result, err, _ := c.flight.Do(key, func() (any, error) {
		items, err := fetch(ctx)
		if err != nil {
			return nil, err
		}

		c.mu.Lock()
		c.values[key] = cacheEntry{
			value:     items,
			expiresAt: time.Now().Add(c.ttl),
		}
		c.mu.Unlock()

		return items, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]Item), nil
}

func readCacheTTL() time.Duration {
	raw := os.Getenv(envCacheTTL)
	if raw == "" {
		slog.Debug("kube api response cache using default ttl",
			slog.String("env", envCacheTTL),
			slog.Duration("ttl", defaultCacheTTL),
		)
		return defaultCacheTTL
	}

	d, err := time.ParseDuration(raw)
	if err != nil {
		slog.Warn("invalid kube api response cache ttl, falling back to default",
			slog.String("env", envCacheTTL),
			slog.String("value", raw),
			slog.Duration("default", defaultCacheTTL),
			slog.Any("err", err),
		)
		return defaultCacheTTL
	}

	if d < 0 {
		slog.Warn("negative kube api response cache ttl, falling back to default",
			slog.String("env", envCacheTTL),
			slog.Duration("value", d),
			slog.Duration("default", defaultCacheTTL),
		)
		return defaultCacheTTL
	}

	if d == 0 {
		slog.Info("kube api response cache disabled", slog.String("env", envCacheTTL))
	} else {
		slog.Debug("kube api response cache enabled",
			slog.String("env", envCacheTTL),
			slog.Duration("ttl", d),
		)
	}

	return d
}
