package cache

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache defines the interface for URL caching
type Cache interface {
	Get(shortCode string) (string, error)
	Set(shortCode, originalURL string, ttl time.Duration) error
	Delete(shortCode string) error
	Close() error
}

// RedisCache implements Cache using Redis
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache creates a new Redis cache client
func NewRedisCache(addr, password string, db int) (Cache, error) {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

// Get retrieves a URL from cache
// Returns empty string if the key is not found (cache miss)
func (c *RedisCache) Get(shortCode string) (string, error) {
	val, err := c.client.Get(c.ctx, "url:"+shortCode).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

// Set stores a URL in cache with a time-to-live (TTL)
func (c *RedisCache) Set(shortCode, originalURL string, ttl time.Duration) error {
	return c.client.Set(c.ctx, "url:"+shortCode, originalURL, ttl).Err()
}

// Delete removes a URL from cache
func (c *RedisCache) Delete(shortCode string) error {
	return c.client.Del(c.ctx, "url:"+shortCode).Err()
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// MemoryCache is a simple in-memory cache for testing
type MemoryCache struct {
	mu   sync.RWMutex
	data map[string]cacheEntry
}

type cacheEntry struct {
	value     string
	expiresAt time.Time
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
