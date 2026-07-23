package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"url-shortener/internal/service"
)

// Handler holds HTTP handlers for the URL shortener
type Handler struct {
	service *service.Service
}

// NewHandler creates a new HTTP handler
func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

// ShortenURL handles POST /api/shorten
// It accepts a JSON body with a URL and returns a short code
func (h *Handler) ShortenURL(c *gin.Context) {
	var req service.ShortenRequest

	// Bind JSON body to request struct
	// Gin automatically validates "required" fields
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	// Call the service to shorten the URL
	resp, err := h.service.Shorten(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Return the shortened URL
	c.JSON(http.StatusCreated, resp)
}

// RedirectURL handles GET /:shortCode
// It looks up the short code and redirects to the original URL
func (h *Handler) RedirectURL(c *gin.Context) {
	shortCode := c.Param("shortCode")

	// A short code should be at least 1 character and at most 10
	if len(shortCode) == 0 || len(shortCode) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid short code",
		})
		return
	}

	// Resolve the short code to original URL
	originalURL, err := h.service.Resolve(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Redirect to the original URL
	// Status 302 means temporary redirect (good for short URLs)
	c.Redirect(http.StatusFound, originalURL)
}

// HealthCheck handles GET /health
// Returns a simple status to confirm the server is running
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "url-shortener",
	})
}
