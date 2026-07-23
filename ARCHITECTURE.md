# Architecture

## System Overview

The URL Shortener is a simple 3-tier application:

┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│   Gin API   │────▶│   Redis     │
│  (Browser)  │     │   (Go)      │     │   (Cache)   │
└─────────────┘     └──────┬──────┘     └─────────────┘
│
▼
┌─────────────┐
│  PostgreSQL │
│ (Database)  │
└─────────────┘


## Request Flow

### Shorten URL (POST /api/shorten)

1. Client sends JSON `{"url": "https://example.com"}`
2. Rate limiter checks if client is within limits
3. Handler validates the request
4. Service generates a unique Base62 short code
5. Service saves the URL to PostgreSQL
6. Service caches the mapping in Redis (24h TTL)
7. Handler returns `{"short_code": "abc123", "short_url": "http://localhost:8080/abc123"}`

### Redirect (GET /:shortCode)

1. Client requests `GET /abc123`
2. Rate limiter checks limits
3. Service checks Redis cache first (fast path)
4. If cache miss, Service queries PostgreSQL
5. Service increments click count
6. Client receives HTTP 302 redirect to original URL

## Design Decisions

| Decision | Reason |
|----------|--------|
| **Base62 encoding** | Short, URL-safe codes using only alphanumeric characters |
| **Redis caching** | Sub-millisecond lookups for popular URLs |
| **Atomic counter + timestamp** | Unique IDs without database sequence contention |
| **Rate limiting per IP** | Prevents abuse without requiring user accounts |
| **Docker Compose** | One command to run the entire stack |

## Technology Stack

- **Go 1.22** — Fast, simple, compiled language
- **Gin** — Minimal web framework with good middleware support
- **PostgreSQL 16** — Reliable, ACID-compliant persistence
- **Redis 7** — In-memory cache with TTL support
- **Docker** — Containerization for consistent environments