package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/user/log-aggregator/common"
)

// Configuration options
var (
	topicFlag  = flag.String("topic", "logs", "Kafka topic to consume logs from")
	outputFlag = flag.String("output", "logs.json", "File to write logs to")
	bufferFlag = flag.Int("buffer", 100, "Number of logs to buffer before writing to file")
)

func main() {
	flag.Parse()

	// Create Kafka consumer
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        "localhost:9092",
		"group.id":                 "log-consumer-group",
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  5000,
		"session.timeout.ms":       30000,
	})

	if err != nil {
		log.Fatalf("Failed to create consumer: %s", err)
	}
	defer c.Close()

	// Subscribe to topic
	err = c.SubscribeTopics([]string{*topicFlag}, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %s", err)
	}

	fmt.Printf("Subscribed to topic: %s\n", *topicFlag)

	// Set up signal handling for graceful shutdown
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Set up buffer channel and file writer
	logBuffer := make(chan common.LogEntry, *bufferFlag)
	var wg sync.WaitGroup

	// Start file writer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		processLogs(logBuffer, *outputFlag)
	}()

	// Process messages until termination signal
	run := true
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				fmt.Printf("Received message from topic %s [%d] at offset %v\n",
					*e.TopicPartition.Topic, e.TopicPartition.Partition, e.TopicPartition.Offset)

				// Parse log entry
				logEntry, err := common.FromJSON(e.Value)
				if err != nil {
					log.Printf("Error parsing log entry: %v", err)
					continue
				}

				// Send to buffer
				logBuffer <- logEntry

			case kafka.Error:
				fmt.Fprintf(os.Stderr, "Error: %v\n", e)
			}
		}
	}

	// Close channel and wait for writer to finish
	close(logBuffer)
	wg.Wait()
	fmt.Println("Consumer shutdown complete")
}

// processLogs writes log entries to a file
func processLogs(logBuffer chan common.LogEntry, outputFile string) {
	var logs []common.LogEntry
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Create or open output file
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	// Check if file is empty, write opening bracket if needed
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("Failed to stat output file: %v", err)
	}

	// If the file is empty, initialize it as a JSON array
	if fileInfo.Size() == 0 {
		file.WriteString("[\n")
	} else {
		// If file exists and has content, seek to remove the closing bracket
		file.Seek(-2, 2) // Seek to position before "]"
	}

	// Buffer logs and write periodically
	for {
		select {
		case log, ok := <-logBuffer:
			if !ok {
				// Channel closed, write remaining logs and exit
				writeLogs(file, logs, true)
				return
			}
			logs = append(logs, log)
			
			// Write immediately if buffer is full
			if len(logs) >= *bufferFlag {
				writeLogs(file, logs, false)
				logs = logs[:0]
			}
			
		case <-ticker.C:
			// Write periodically even if buffer isn't full
			if len(logs) > 0 {
				writeLogs(file, logs, false)
				logs = logs[:0]
			}
		}
	}
}

// writeLogs writes buffered logs to file
func writeLogs(file *os.File, logs []common.LogEntry, isFinal bool) {
	if len(logs) == 0 {
		return
	}
	
	for i, entry := range logs {
		// Convert to JSON
		data, err := json.MarshalIndent(entry, "  ", "  ")
		if err != nil {
			log.Printf("Error marshaling log entry: %v", err)
			continue
		}
		
		// Write comma separator if needed
		if i > 0 || file.Stat().Size() > 2 { // 2 is the size of "[\n"
			file.WriteString(",\n")
		}
		
		// Write log entry
		file.Write(data)
	}
	
	// If this is the final write, close the JSON array
	if isFinal {
		file.WriteString("\n]")
	}
	
	file.Sync()
	fmt.Printf("Wrote %d logs to %s\n", len(logs), file.Name())
} 