package cache

import (
	"sync"
	"time"
)

// Cache defines the interface for URL caching
type Cache interface {
	Get(shortCode string) (string, error)
	Set(shortCode, originalURL string, ttl time.Duration) error
	Delete(shortCode string) error
	Close() error
}

// MemoryCache is a simple in-memory cache for deployment
type MemoryCache struct {
	mu   sync.RWMutex
	data map[string]cacheEntry
}

type cacheEntry struct {
	value     string
	expiresAt time.Time
}

// NewRedisCache creates a new cache (uses MemoryCache for deployment)
func NewRedisCache(addr, password string, db int) (Cache, error) {
	return NewMemoryCache(), nil
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache() Cache {
	return &MemoryCache{
		data: make(map[string]cacheEntry),
	}
}

// Get retrieves a URL from memory cache
func (m *MemoryCache) Get(shortCode string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, ok := m.data["url:"+shortCode]
	if !ok || time.Now().After(entry.expiresAt) {
		return "", nil
	}
	return entry.value, nil
}

// Set stores a URL in memory cache with TTL
func (m *MemoryCache) Set(shortCode, originalURL string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data["url:"+shortCode] = cacheEntry{
		value:     originalURL,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

// Delete removes a URL from memory cache
func (m *MemoryCache) Delete(shortCode string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, "url:"+shortCode)
	return nil
}

// Close is a no-op for memory cache
func (m *MemoryCache) Close() error {
	return nil
}
