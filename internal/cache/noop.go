package cache

import (
	"time"

	"github.com/chubin/wttr.in/internal/domain"
	"github.com/chubin/wttr.in/internal/weather"
)

// NoOpCacher is a no-operation implementation of Cacher.
// Used when a cache layer is disabled in configuration.
type NoOpCacher struct{}

// NewNoOpCacher returns a new no-op cache.
func NewNoOpCacher() weather.Cacher {
	return &NoOpCacher{}
}

func (n *NoOpCacher) Get(key string) *domain.CacheEntry {
	return nil // never a cache hit
}

func (n *NoOpCacher) Set(key string, entry domain.CacheEntry) {
	// do nothing
}

func (n *NoOpCacher) SetInProgressIfNotExists(key string) bool {
	return true // always allow the caller to proceed with fetching
}

func (n *NoOpCacher) IsInProgress(key string) bool {
	return false
}

func (n *NoOpCacher) WaitForCompletion(key string, maxWait time.Duration) (*domain.CacheEntry, error) {
	// Since we never set "in progress" for others, we can return immediately with no entry
	return nil, nil
}

func (n *NoOpCacher) Remove(key string) {
	// do nothing
}

func (n *NoOpCacher) Close() error {
	return nil
}
