// Package cache provides a type-safe wrapper around go-cache for caching API responses.
package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

// Cache wraps go-cache with type-safe accessors
type Cache struct {
	c *gocache.Cache
}

// New creates a new cache with default expiration and cleanup interval
func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return &Cache{
		c: gocache.New(defaultExpiration, cleanupInterval),
	}
}

// Set adds an item to the cache with the default expiration
func (c *Cache) Set(key string, value any) {
	c.c.Set(key, value, gocache.DefaultExpiration)
}

// SetWithExpiration adds an item to the cache with a custom expiration
func (c *Cache) SetWithExpiration(key string, value any, d time.Duration) {
	c.c.Set(key, value, d)
}

// Get retrieves an item from the cache with type assertion
func Get[T any](c *Cache, key string) (T, bool) {
	var zero T

	val, found := c.c.Get(key)
	if !found {
		return zero, false
	}

	typed, ok := val.(T)
	if !ok {
		return zero, false
	}

	return typed, true
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.c.Delete(key)
}

// Flush removes all items from the cache
func (c *Cache) Flush() {
	c.c.Flush()
}
