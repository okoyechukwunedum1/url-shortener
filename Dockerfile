# Build stage: compile the Go binary
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Install git and ca-certificates (needed for fetching dependencies)
RUN apk add --no-cache git ca-certificates

# Copy go.mod and go.sum first (for better Docker layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code
COPY . .

# Build the binary
# CGO_ENABLED=0 creates a static binary with no C dependencies
# -ldflags="-w -s" strips debug info to make the binary smaller
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# Final stage: minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/server .

# Expose port 8080
EXPOSE 8080

# Run the binary
CMD ["./server"]