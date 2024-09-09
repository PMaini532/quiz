# Dockerfile for service3_test
FROM golang:1.22.4-alpine AS builder
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Copy the vendor directory
COPY vendor ./vendor

# Copy the source code
COPY . .

# Build the application
RUN go build -mod=vendor -o /test-service

FROM alpine:3.18.3
WORKDIR /root/
COPY --from=builder /test-service .
EXPOSE 8024
CMD ["./test-service"]
