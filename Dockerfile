# Multi-stage build for AFTERDARK SSH daemon
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/afterdark ./cmd/afterdark

# Minimal runtime image
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S afterdark && adduser -S afterdark -G afterdark \
    && mkdir -p /app/data && chown -R afterdark:afterdark /app

WORKDIR /app
COPY --from=builder /bin/afterdark /app/afterdark

USER afterdark

ENV PORT=2222
ENV HOST=0.0.0.0
ENV DATA_DIR=/app/data
ENV DB_PATH=/app/data/afterdark.db
ENV HOST_KEY_PATH=/app/data/afterdark_ed25519

EXPOSE 2222

VOLUME ["/app/data"]

ENTRYPOINT ["/app/afterdark"]
