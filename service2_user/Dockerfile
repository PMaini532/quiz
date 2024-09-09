# Dockerfile for service2_user
FROM golang:1.22.4-alpine AS builder
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Copy the vendor directory
COPY vendor ./vendor

# Copy the source code
COPY . .

# Build the application
RUN go build -mod=vendor -o /user-service

FROM alpine:3.18.3
WORKDIR /root/
COPY --from=builder /user-service .
EXPOSE 8023
CMD ["./user-service"]
