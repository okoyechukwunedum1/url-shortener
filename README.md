#  URL Shortener

A simple, fast, and beginner-friendly URL shortener built with **Go**, **Gin**, **PostgreSQL**, and **Redis**.

##  Features

- ✅ **Shorten URLs** — Convert long URLs into short Base62 codes
- ✅ **Redirect** — Instant redirection to original URLs
- ✅ **Redis Caching** — Lightning-fast lookups for popular links
- ✅ **PostgreSQL Persistence** — Reliable, durable storage
- ✅ **Rate Limiting** — Protects against abuse (10 req/sec per IP)
- ✅ **Docker Support** — One-command setup with Docker Compose
- ✅ **Health Check** — `GET /health` endpoint for monitoring
- ✅ **Unit Tests** — Comprehensive test coverage

##  Architecture

```mermaid
graph LR
    A[Client] --&gt;|POST /api/shorten| B[Gin API]
    A --&gt;|GET /:code| B
    B --&gt;|Cache Hit| C[Redis]
    B --&gt;|Cache Miss| D[PostgreSQL]
    C --&gt;|Return URL| B
    D --&gt;|Return URL| B
    B --&gt;|302 Redirect| A
