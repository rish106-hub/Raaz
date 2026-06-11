# ── Stage 1: build ───────────────────────────────────────────────────────────
FROM golang:1.21-alpine AS builder
WORKDIR /src

# Cache dep download layer separately from source copy.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -trimpath -o /raaz-server .

# ── Stage 2: run ─────────────────────────────────────────────────────────────
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -S raaz && adduser -S raaz -G raaz

WORKDIR /app
COPY --from=builder /raaz-server .
USER raaz
EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["./raaz-server"]
