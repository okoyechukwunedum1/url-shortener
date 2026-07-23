package model

import "time"

// URL represents a shortened URL in the system
type URL struct {
	ID          int64      `json:"-" db:"id"`                         // Internal ID, hidden from JSON
	ShortCode   string     `json:"short_code" db:"short_code"`        // The short code (e.g., "a3fK9z")
	OriginalURL string     `json:"original_url" db:"original_url"`    // The original long URL
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`        // When the URL was created
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"` // Optional expiration time
	ClickCount  int64      `json:"click_count" db:"click_count"`      // How many times the link was clicked
}