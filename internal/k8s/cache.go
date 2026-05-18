package k8s

import (
	"context"
	"time"

	"golang.org/x/sync/singleflight"
)

// cacheTTL is the fixed TTL for cached cluster-wide List() responses.
// Short enough that dashboard reloads still reflect cluster changes
// quickly, long enough that all widgets on one pageload share results.
const cacheTTL = 5 * time.Second

// cache stores a single typed slice with a TTL. All access goes through
// singleflight, which serializes concurrent get() calls into a single
// fetch and provides the happens-before needed to access value/expiresAt
// without an explicit mutex. Errors are not cached.
type cache[T any] struct {
	value     []T
	expiresAt time.Time
	flight    singleflight.Group
}

// get returns the cached value when fresh; otherwise invokes fetch and
// caches the result.
func (c *cache[T]) get(ctx context.Context, fetch func(context.Context) ([]T, error)) ([]T, error) {
	result, err, _ := c.flight.Do("", func() (any, error) {
		if time.Now().Before(c.expiresAt) {
			return c.value, nil
		}

		v, err := fetch(ctx)
		if err != nil {
			return nil, err
		}

		c.value = v
		c.expiresAt = time.Now().Add(cacheTTL)
		return v, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]T), nil
}
