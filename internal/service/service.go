package service

import (
	"errors"
	"fmt"
	"net/url"
	"sync/atomic"
	"time"

	"url-shortener/internal/cache"
	"url-shortener/internal/model"
	"url-shortener/internal/repository"
	"url-shortener/internal/util"
)

// Service handles the business logic for URL shortening
type Service struct {
	repo    repository.Repository
	cache   cache.Cache
	baseURL string
	counter uint64 // atomic counter for generating unique IDs
}

// NewService creates a new URL shortener service
func NewService(repo repository.Repository, cache cache.Cache, baseURL string) *Service {
	return &Service{
		repo:    repo,
		cache:   cache,
		baseURL: baseURL,
	}
}

// ShortenRequest is what the API receives from the client
type ShortenRequest struct {
	URL string `json:"url" binding:"required"`
}

// ShortenResponse is what the API sends back to the client
type ShortenResponse struct {
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// Shorten takes a long URL and returns a short code
func (s *Service) Shorten(req ShortenRequest) (*ShortenResponse, error) {
	// Validate the URL
	if req.URL == "" {
		return nil, errors.New("URL is required")
	}

	// Parse and validate URL format
	parsed, err := url.Parse(req.URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("invalid URL format")
	}

	// Normalize URL (remove trailing slash for consistency)
	normalizedURL := req.URL
	if len(normalizedURL) > 1 && normalizedURL[len(normalizedURL)-1] == '/' {
		normalizedURL = normalizedURL[:len(normalizedURL)-1]
	}

	// Generate a unique ID using atomic counter + timestamp
	// This ensures uniqueness even with high concurrency
	uniqueID := s.generateUniqueID()

	// Convert the unique ID to a Base62 short code
	shortCode := util.EncodeBase62(uniqueID)

	// Create the URL model
	now := time.Now()
	urlModel := &model.URL{
		ShortCode:   shortCode,
		OriginalURL: normalizedURL,
		CreatedAt:   now,
		ClickCount:  0,
	}

	// Save to database
	if err := s.repo.Save(urlModel); err != nil {
		return nil, fmt.Errorf("failed to save URL: %w", err)
	}

	// Cache the result for fast lookups (TTL: 24 hours)
	cacheTTL := 24 * time.Hour
	_ = s.cache.Set(shortCode, normalizedURL, cacheTTL)

	// Build the full short URL
	shortURL := fmt.Sprintf("%s/%s", s.baseURL, shortCode)

	return &ShortenResponse{
		ShortCode:   shortCode,
		ShortURL:    shortURL,
		OriginalURL: normalizedURL,
	}, nil
}

// Resolve takes a short code and returns the original URL
// It also increments the click count
func (s *Service) Resolve(shortCode string) (string, error) {
	if shortCode == "" {
		return "", errors.New("short code is required")
	}

	// 1. Try cache first (fastest)
	cachedURL, err := s.cache.Get(shortCode)
	if err != nil {
		return "", fmt.Errorf("cache error: %w", err)
	}
	if cachedURL != "" {
		// Cache hit! Increment click count in background
		go s.repo.IncrementClickCount(shortCode)
		return cachedURL, nil
	}

	// 2. Cache miss — check database
	urlModel, err := s.repo.FindByShortCode(shortCode)
	if err != nil {
		return "", fmt.Errorf("database error: %w", err)
	}
	if urlModel == nil {
		return "", errors.New("short code not found")
	}

	// Check if URL has expired
	if urlModel.ExpiresAt != nil && time.Now().After(*urlModel.ExpiresAt) {
		return "", errors.New("short URL has expired")
	}

	// 3. Store in cache for next time
	cacheTTL := 24 * time.Hour
	_ = s.cache.Set(shortCode, urlModel.OriginalURL, cacheTTL)

	// Increment click count
	go s.repo.IncrementClickCount(shortCode)

	return urlModel.OriginalURL, nil
}

// generateUniqueID creates a unique number by combining timestamp and atomic counter
func (s *Service) generateUniqueID() uint64 {
	// Use current timestamp in milliseconds as the base
	timestamp := uint64(time.Now().UnixMilli())

	// Add an atomic counter to handle multiple requests in the same millisecond
	counter := atomic.AddUint64(&s.counter, 1)

	// Combine them: timestamp in high bits, counter in low bits
	return (timestamp << 16) | (counter & 0xFFFF)
}
