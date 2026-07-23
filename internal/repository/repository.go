package repository

import (
	"database/sql"
	"sync"

	"url-shortener/internal/model"

	_ "github.com/lib/pq"
)

// Repository defines the interface for URL data storage
// This allows us to swap PostgreSQL with a memory store for testing
type Repository interface {
	Save(url *model.URL) error
	FindByShortCode(shortCode string) (*model.URL, error)
	IncrementClickCount(shortCode string) error
	Close() error
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
// It automatically creates the required table if it doesn't exist
func NewPostgresRepository(databaseURL string) (Repository, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	repo := &PostgresRepository{db: db}
	if err := repo.createTable(); err != nil {
		return nil, err
	}

	return repo, nil
}

// createTable creates the urls table and index if they don't exist
func (r *PostgresRepository) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id BIGSERIAL PRIMARY KEY,
		short_code VARCHAR(10) UNIQUE NOT NULL,
		original_url TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP WITH TIME ZONE,
		click_count BIGINT DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
	`
	_, err := r.db.Exec(query)
	return err
}

// Save inserts a new URL into the database and sets the ID
func (r *PostgresRepository) Save(url *model.URL) error {
	query := `
		INSERT INTO urls (short_code, original_url, created_at, expires_at, click_count)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	return r.db.QueryRow(
		query,
		url.ShortCode,
		url.OriginalURL,
		url.CreatedAt,
		url.ExpiresAt,
		url.ClickCount,
	).Scan(&url.ID)
}

// FindByShortCode looks up a URL by its short code
// Returns nil if not found
func (r *PostgresRepository) FindByShortCode(shortCode string) (*model.URL, error) {
	query := `SELECT id, short_code, original_url, created_at, expires_at, click_count FROM urls WHERE short_code = $1`
	row := r.db.QueryRow(query, shortCode)

	var url model.URL
	err := row.Scan(&url.ID, &url.ShortCode, &url.OriginalURL, &url.CreatedAt, &url.ExpiresAt, &url.ClickCount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &url, nil
}

// IncrementClickCount increases the click count for a URL by 1
func (r *PostgresRepository) IncrementClickCount(shortCode string) error {
	query := `UPDATE urls SET click_count = click_count + 1 WHERE short_code = $1`
	_, err := r.db.Exec(query, shortCode)
	return err
}

// Close closes the database connection
func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

// MemoryRepository is a simple in-memory implementation for testing
// It is NOT thread-safe for production use but fine for unit tests
type MemoryRepository struct {
	mu   sync.RWMutex
	data map[string]*model.URL
}

// NewMemoryRepository creates a new in-memory repository
func NewMemoryRepository() Repository {
	return &MemoryRepository{
		data: make(map[string]*model.URL),
	}
}

// Save stores a URL in memory
func (m *MemoryRepository) Save(url *model.URL) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if url.ID == 0 {
		url.ID = int64(len(m.data) + 1)
	}
	m.data[url.ShortCode] = url
	return nil
}

// FindByShortCode retrieves a URL from memory
func (m *MemoryRepository) FindByShortCode(shortCode string) (*model.URL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if url, ok := m.data[shortCode]; ok {
		return url, nil
	}
	return nil, nil
}

// IncrementClickCount increases the click count in memory
func (m *MemoryRepository) IncrementClickCount(shortCode string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if url, ok := m.data[shortCode]; ok {
		url.ClickCount++
	}
	return nil
}

// Close is a no-op for memory repository
func (m *MemoryRepository) Close() error {
	return nil
}
