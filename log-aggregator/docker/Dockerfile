FROM golang:1.20-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go.mod and go.sum
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build producer and consumer
RUN go build -o bin/producer ./producer
RUN go build -o bin/consumer ./consumer

# Create final lightweight image
FROM alpine:latest

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/bin/producer /app/producer
COPY --from=builder /app/bin/consumer /app/consumer

# Create volume for output logs
VOLUME /app/logs

# Set environment variables (can be overridden)
ENV KAFKA_BOOTSTRAP_SERVERS=kafka:29092
ENV LOG_TOPIC=logs
ENV LOG_RATE=1
ENV LOG_OUTPUT=/app/logs/logs.json

# Command will be provided when running the container
ENTRYPOINT ["sh"] 