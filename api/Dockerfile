FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev make git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build --tags "fts5" -o cnapi ./cmd/server

FROM alpine:latest

# Install runtime dependencies for SQLite
RUN apk add --no-cache ca-certificates sqlite

WORKDIR /app

COPY --from=builder /app/cnapi .

COPY --from=builder /app/sources ./sources

RUN mkdir -p data

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./cnapi"]