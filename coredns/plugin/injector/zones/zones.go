package zones

import (
	"context"
	"sync"
	"time"
)

type Provider interface {
	Zones(ctx context.Context) ([]string, error)
}

type ProviderFunc func(ctx context.Context) ([]string, error)

func (f ProviderFunc) Zones(ctx context.Context) ([]string, error) {
	return f(ctx)
}

const failureBackoff = 5 * time.Second

type Cache struct {
	provider Provider
	ttl      time.Duration

	mu      sync.Mutex
	zones   []string
	err     error
	expires time.Time
}

func NewCache(p Provider, ttl time.Duration) *Cache {
	return &Cache{provider: p, ttl: ttl}
}

// Zones loads zones on first use and reloads them once the cached value is older than ttl.
// If a reload fails, the last loaded zones (nil if there are none) are returned together with the error.
// The returned slice must not be modified.
func (c *Cache) Zones(ctx context.Context) ([]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if time.Now().Before(c.expires) {
		return c.zones, c.err
	}
	zones, err := c.provider.Zones(ctx)
	c.err = err
	if err != nil {
		// Avoid calling a failing provider on every request.
		c.expires = time.Now().Add(min(c.ttl, failureBackoff))
		return c.zones, err
	}
	c.zones = zones
	c.expires = time.Now().Add(c.ttl)
	return zones, nil
}
