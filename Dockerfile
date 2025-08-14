# --------------------------
# 1️⃣ Build stage
# --------------------------
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o exchange-rate-service ./cmd/server

# --------------------------
# 2️⃣ Runtime stage
# --------------------------
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/exchange-rate-service .

# ✅ Copy the .env file into container
COPY .env .

EXPOSE 8080

CMD ["./exchange-rate-service"]
