# --------------------------
# 1️⃣ Build stage (Alpine-based Go)
# --------------------------
FROM golang:1.22-alpine AS builder

# Install build tools
RUN apk add --no-cache git

WORKDIR /app

# Copy only go.mod & go.sum first for caching
COPY go.mod go.sum ./
RUN go mod download

# Now copy the rest of the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o exchange-rate-service ./cmd/server

# --------------------------
# 2️⃣ Runtime stage (small Alpine image)
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
