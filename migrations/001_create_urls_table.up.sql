-- Create the urls table for storing shortened URLs
CREATE TABLE IF NOT EXISTS urls (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    click_count BIGINT DEFAULT 0
);

-- Create an index on short_code for fast lookups
CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);