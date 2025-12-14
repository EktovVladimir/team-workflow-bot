# Multi-stage build for Go app
FROM golang:1.25-alpine AS builder

# Enable static build
ENV CGO_ENABLED=0
WORKDIR /app

# Install git (if needed for go modules)
RUN apk add --no-cache git

# Copy go mod/sum and download deps
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the wfbot command
RUN go build -o /out/wfbot ./cmd/wfbot

# Final minimal image
FROM alpine:3.20

# Create non-root user
RUN adduser -D -u 10001 appuser

# Workdir and copy binary
WORKDIR /app
COPY --from=builder /out/wfbot /app/wfbot

# Run as non-root
USER appuser

# Default command
CMD ["/app/wfbot"]

