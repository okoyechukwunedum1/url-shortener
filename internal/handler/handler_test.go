package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"url-shortener/internal/cache"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"
)

func setupTestRouter() (*gin.Engine, *Handler) {
	// Use test mode to suppress Gin's default logging
	gin.SetMode(gin.TestMode)

	repo := repository.NewMemoryRepository()
	c := cache.NewMemoryCache()
	svc := service.NewService(repo, c, "http://localhost:8080")
	handler := NewHandler(svc)

	router := gin.New()
	return router, handler
}

func TestShortenURL(t *testing.T) {
	router, handler := setupTestRouter()

	router.POST("/api/shorten", handler.ShortenURL)

	body := `{"url":"https://www.example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var resp service.ShortenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.ShortCode == "" {
		t.Error("Expected short_code in response")
	}
	if resp.ShortURL == "" {
		t.Error("Expected short_url in response")
	}
}

func TestShortenURLInvalid(t *testing.T) {
	router, handler := setupTestRouter()
	router.POST("/api/shorten", handler.ShortenURL)

	body := `{"url":"not-a-valid-url"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRedirectURL(t *testing.T) {
	router, handler := setupTestRouter()

	router.POST("/api/shorten", handler.ShortenURL)
	router.GET("/:shortCode", handler.RedirectURL)

	// First, create a short URL
	body := `{"url":"https://www.google.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp service.ShortenResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Now, try to redirect
	req2 := httptest.NewRequest(http.MethodGet, "/"+resp.ShortCode, nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Should redirect (302 Found)
	if w2.Code != http.StatusFound {
		t.Errorf("Expected redirect status %d, got %d", http.StatusFound, w2.Code)
	}

	// Check the Location header
	location := w2.Header().Get("Location")
	if location != "https://www.google.com" {
		t.Errorf("Expected redirect to https://www.google.com, got %s", location)
	}
}

func TestRedirectURLNotFound(t *testing.T) {
	router, handler := setupTestRouter()
	router.GET("/:shortCode", handler.RedirectURL)

	req := httptest.NewRequest(http.MethodGet, "/doesnotexist", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHealthCheck(t *testing.T) {
	router, handler := setupTestRouter()
	router.GET("/health", handler.HealthCheck)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("ok")) {
		t.Error("Expected health check to return 'ok'")
	}
}
