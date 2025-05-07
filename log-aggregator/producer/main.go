package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"github.com/darshangupta/log-aggregator/common"
)

// Configuration options
var (
	rateFlag  = flag.Int("rate", 1, "Number of logs per second to produce")
	topicFlag = flag.String("topic", "logs", "Kafka topic to send logs to")
)

// Sample data for log generation
var (
	services = []string{"auth-service", "api-gateway", "user-service", "payment-service", "order-service"}
	
	logMessages = map[common.LogLevel][]string{
		common.DEBUG: {
			"Parsing request payload",
			"Cache hit for user profile",
			"Database connection established",
			"Starting background job",
			"API request received",
		},
		common.INFO: {
			"User logged in successfully",
			"Payment processed",
			"Order created",
			"Email sent to customer",
			"API request completed successfully",
		},
		common.WARNING: {
			"High API latency detected",
			"Database connection pool nearing capacity",
			"Rate limit threshold approaching",
			"Cache miss for frequently accessed resource",
			"Request retry attempted",
		},
		common.ERROR: {
			"Failed to connect to database",
			"API timeout error",
			"Payment processing failed",
			"Invalid authentication token",
			"Service dependency unavailable",
		},
		common.FATAL: {
			"Database connection pool exhausted",
			"Out of memory error",
			"Critical security breach detected",
			"Unrecoverable system error",
			"Service initialization failed",
		},
	}
	
	userIDs = []string{"usr_123456", "usr_789012", "usr_345678", "usr_901234", "usr_567890"}
	ipAddrs = []string{"192.168.1.1", "10.0.0.2", "172.16.0.5", "192.168.0.10", "10.1.1.15"}
)

func main() {
	flag.Parse()
	
	// Validate rate
	if *rateFlag <= 0 {
		log.Fatal("Rate must be greater than 0")
	}
	
	// Create Kafka producer
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"client.id":         "log-producer",
		"acks":              "all",
	})
	
	if err != nil {
		log.Fatalf("Failed to create producer: %s", err)
	}
	defer p.Close()
	
	// Handle delivery reports
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition.Error)
				} else {
					fmt.Printf("Log delivered to %v at offset %v\n",
						ev.TopicPartition.Topic, ev.TopicPartition.Offset)
				}
			}
		}
	}()
	
	// Set up signal handling for graceful shutdown
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	
	// Calculate delay between log entries based on rate
	delay := time.Second / time.Duration(*rateFlag)
	
	// Start log generation loop
	fmt.Printf("Starting log producer at %d logs/second (every %v)\n", *rateFlag, delay)
	
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	
	logCount := 0
	
	for {
		select {
		case <-ticker.C:
			// Generate and send a log entry
			logEntry := generateRandomLog()
			logData, err := logEntry.ToJSON()
			if err != nil {
				log.Printf("Error serializing log: %v", err)
				continue
			}
			
			// Send to Kafka
			err = p.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: topicFlag, Partition: kafka.PartitionAny},
				Value:          logData,
				Key:            []byte(strconv.Itoa(logCount)),
			}, nil)
			
			if err != nil {
				log.Printf("Error sending log to Kafka: %v", err)
			}
			
			logCount++
			
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			// Wait for messages to be delivered
			p.Flush(15 * 1000)
			return
		}
	}
}

// generateRandomLog creates a random log entry for testing
func generateRandomLog() common.LogEntry {
	// Randomize level with weighted distribution
	levelDist := []common.LogLevel{
		common.DEBUG, common.DEBUG,  // 20%
		common.INFO, common.INFO, common.INFO, common.INFO, common.INFO, // 50% 
		common.WARNING, common.WARNING, // 20%
		common.ERROR,  // 10%
		// FATAL is rare, so not included in random distribution
	}
	
	level := levelDist[rand.Intn(len(levelDist))]
	service := services[rand.Intn(len(services))]
	message := logMessages[level][rand.Intn(len(logMessages[level]))]
	
	metadata := map[string]string{
		"user_id":     userIDs[rand.Intn(len(userIDs))],
		"ip":          ipAddrs[rand.Intn(len(ipAddrs))],
		"request_id":  uuid.New().String(),
		"environment": "production",
	}
	
	return common.NewLogEntry(level, service, message, metadata)
} 