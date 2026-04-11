# ---- Build Stage ----
FROM golang:1.26-alpine AS builder

# CGO is required for go-sqlite3
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with CGO enabled
RUN CGO_ENABLED=1 go build -o /app/gitpigeon ./cmd/api

# ---- Runtime Stage ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary and migrations from builder
COPY --from=builder /app/gitpigeon .
COPY --from=builder /app/migrations ./migrations

# Create directory for SQLite database
RUN mkdir -p /app/data

EXPOSE 8080

CMD ["./gitpigeon"]
