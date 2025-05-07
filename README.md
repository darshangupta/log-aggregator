# Log Aggregator

A simple log aggregation system built with Go and Kafka.

## Architecture

```
┌──────────────┐      ┌───────────────┐      ┌────────────────────┐
│ Your Service │───▶ │ Kafka Producer │───▶ │ Kafka Topic: logs  │
└──────────────┘      └───────────────┘      └────────┬───────────┘
                                                     ▼
                                          ┌────────────────────┐
                                          │ Kafka Consumer (Go)│
                                          └────────┬───────────┘
                                                   ▼
                                          ┌────────────────────┐
                                          │ Local DB / JSON FS │
                                          └────────────────────┘
```

## Project Structure

```
log-aggregator/
├── producer/              # Sends logs to Kafka
│   └── main.go
├── consumer/              # Reads logs from Kafka, saves to file/db
│   └── main.go
├── common/                # Shared code and models
│   └── log.go
├── docker/                # Docker-related files
├── docker-compose.yml     # For Kafka + Zookeeper
├── go.mod
└── README.md
```

## Setup and Usage

### Prerequisites

- Go 1.20+ installed
- Docker and Docker Compose installed

### Getting Started

1. Start Kafka and Zookeeper:

```
docker-compose up -d
```

2. Build the producer and consumer:

```
go build -o bin/producer ./producer
go build -o bin/consumer ./consumer
```

3. Run the consumer to listen for logs:

```
./bin/consumer
```

4. In a separate terminal, run the producer to generate logs:

```
./bin/producer
```

### Configuration

Both the producer and consumer support various command-line flags:

- Producer:
  - `-rate`: Number of logs per second (default: 1)
  - `-topic`: Kafka topic name (default: "logs")

- Consumer:
  - `-topic`: Kafka topic name (default: "logs")
  - `-output`: Output file path (default: "logs.json")

## Log Format

```json
{
  "timestamp": "2023-05-06T12:00:00Z",
  "level": "INFO",
  "service": "auth-service",
  "message": "User login successful",
  "metadata": {
    "user_id": "abc123",
    "ip": "192.168.1.1"
  }
}
``` 