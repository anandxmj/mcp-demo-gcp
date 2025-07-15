# Multi-stage Dockerfile for flight-ticket-service
# Stage 1: Build stage
FROM golang:1.24.5-alpine AS builder

# Install git and ca-certificates (needed for fetching dependencies)
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install swag CLI tool for generating OpenAPI docs
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy source code
COPY . .

# Generate OpenAPI documentation
RUN $(go env GOPATH)/bin/swag init -g src/cmd/server/server.go -o docs

# Build the application with optimizations
# CGO_ENABLED=0 for static binary
# -ldflags="-w -s" to strip debug info and reduce binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o flight-ticket-service \
    ./src/cmd/server

# Stage 2: Final stage using Google's distroless image
FROM gcr.io/distroless/static-debian12:nonroot

# Copy ca-certificates from builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary from builder stage
COPY --from=builder /app/flight-ticket-service /flight-ticket-service

# Copy generated documentation
COPY --from=builder /app/docs /docs

# Set environment variables
ENV PORT=8080
ENV GIN_MODE=release

# Expose port 8080 for Google Cloud Run
EXPOSE 8080

# Use nonroot user (already set by distroless:nonroot)
USER nonroot:nonroot

# Set the entrypoint
ENTRYPOINT ["/flight-ticket-service"]
