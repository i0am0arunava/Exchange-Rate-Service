# --------------------------
# 1️⃣ Build stage (Alpine-based Go)
# --------------------------
FROM golang:1.23-alpine AS builder


# Install build tools
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum first (cache deps)
COPY go.mod go.sum ./
RUN go mod download

# Copy only necessary source files (avoid copying tests, docs, etc.)
COPY cmd/ cmd/
COPY internal/ internal/
COPY .env .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o exchange-rate-service ./cmd/server

# --------------------------
# 2️⃣ Runtime stage
# --------------------------
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/exchange-rate-service .
COPY .env .

EXPOSE 8080

CMD ["./exchange-rate-service"]
