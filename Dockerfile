# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/server

# Final stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/internal/config/config.yaml ./internal/config/
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./server"]
