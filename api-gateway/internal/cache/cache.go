package cache

import (
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
)

// Cache is a thin wrapper around a ristretto in-memory cache that stores
// raw byte slices (e.g. serialised HTTP responses).
type Cache struct {
	store      *ristretto.Cache
	defaultTTL time.Duration
}

// New creates a Cache with the given maximum memory cost (in bytes) and
// default entry TTL.
func New(maxCostBytes int64, defaultTTL time.Duration) (*Cache, error) {
	store, err := ristretto.NewCache(&ristretto.Config{
		// NumCounters should be ~10× the expected number of unique items.
		NumCounters: 1_000_000,
		MaxCost:     maxCostBytes,
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("cache: init ristretto: %w", err)
	}
	return &Cache{store: store, defaultTTL: defaultTTL}, nil
}

// Get returns the cached byte slice for key, or (nil, false) if not present.
func (c *Cache) Get(key string) ([]byte, bool) {
	val, ok := c.store.Get(key)
	if !ok {
		return nil, false
	}
	b, ok := val.([]byte)
	return b, ok
}

// Set stores val under key with the default TTL.
func (c *Cache) Set(key string, val []byte) {
	c.store.SetWithTTL(key, val, int64(len(val)), c.defaultTTL)
}

// SetTTL stores val under key with an explicit TTL.
func (c *Cache) SetTTL(key string, val []byte, ttl time.Duration) {
	c.store.SetWithTTL(key, val, int64(len(val)), ttl)
}

// Del removes key from the cache.
func (c *Cache) Del(key string) {
	c.store.Del(key)
}

// Close stops the cache's internal goroutines.
func (c *Cache) Close() {
	c.store.Close()
}
