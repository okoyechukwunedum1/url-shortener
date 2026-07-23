package service

import (
	"testing"

	"url-shortener/internal/cache"
	"url-shortener/internal/repository"
)

func TestShortenAndResolve(t *testing.T) {
	// Use in-memory implementations for testing
	repo := repository.NewMemoryRepository()
	c := cache.NewMemoryCache()
	svc := NewService(repo, c, "http://localhost:8080")

	// Test shortening a URL
	req := ShortenRequest{URL: "https://www.google.com/search?q=golang"}
	resp, err := svc.Shorten(req)
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	if resp.ShortCode == "" {
		t.Error("Expected short code to be generated")
	}
	if resp.ShortURL == "" {
		t.Error("Expected short URL to be generated")
	}
	if resp.OriginalURL != "https://www.google.com/search?q=golang" {
		t.Errorf("Expected original URL to be preserved, got %s", resp.OriginalURL)
	}

	// Test resolving the short code
	originalURL, err := svc.Resolve(resp.ShortCode)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if originalURL != "https://www.google.com/search?q=golang" {
		t.Errorf("Expected %s, got %s", "https://www.google.com/search?q=golang", originalURL)
	}

	// Test resolving again (should hit cache)
	originalURL2, err := svc.Resolve(resp.ShortCode)
	if err != nil {
		t.Fatalf("Resolve second time failed: %v", err)
	}
	if originalURL2 != originalURL {
		t.Error("Cache returned different URL")
	}
}

func TestShortenInvalidURL(t *testing.T) {
	repo := repository.NewMemoryRepository()
	c := cache.NewMemoryCache()
	svc := NewService(repo, c, "http://localhost:8080")

	// Empty URL
	_, err := svc.Shorten(ShortenRequest{URL: ""})
	if err == nil {
		t.Error("Expected error for empty URL")
	}

	// Invalid URL format
	_, err = svc.Shorten(ShortenRequest{URL: "not-a-url"})
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestResolveNotFound(t *testing.T) {
	repo := repository.NewMemoryRepository()
	c := cache.NewMemoryCache()
	svc := NewService(repo, c, "http://localhost:8080")

	_, err := svc.Resolve("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent short code")
	}
}

func TestShortenMultipleURLs(t *testing.T) {
	repo := repository.NewMemoryRepository()
	c := cache.NewMemoryCache()
	svc := NewService(repo, c, "http://localhost:8080")

	urls := []string{
		"https://example.com/page1",
		"https://example.com/page2",
		"https://example.com/page3",
	}

	codes := make(map[string]bool)

	for _, u := range urls {
		resp, err := svc.Shorten(ShortenRequest{URL: u})
		if err != nil {
			t.Fatalf("Shorten failed for %s: %v", u, err)
		}

		// Ensure each code is unique
		if codes[resp.ShortCode] {
			t.Errorf("Duplicate short code generated: %s", resp.ShortCode)
		}
		codes[resp.ShortCode] = true
	}
}
