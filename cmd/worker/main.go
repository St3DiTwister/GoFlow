package main

import (
	"GoFlow/internal/kafka"
	"GoFlow/internal/model"
	"GoFlow/internal/storage"
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}

	chStorage, err := storage.NewClickHouseStorage(
		fmt.Sprintf("localhost:%s", os.Getenv("CH_PORT_TCP")),
		os.Getenv("CH_USER"),
		os.Getenv("CH_PASSWORD"),
		os.Getenv("CH_DB"),
	)
	if err != nil {
		log.Fatalf("ClickHouse error: %v", err)
	}

	brokers := []string{os.Getenv("KAFKA_BROKER")}
	topic := os.Getenv("KAFKA_TOPIC")
	groupID := "event-processor-v1"
	consumer := kafka.NewConsumer(brokers, topic, groupID)

	eventsChan := make(chan model.Event, 1000)

	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	wg.Add(1)
	go func() {
		defer wg.Done()
		batch := make([]model.Event, 0, 1000)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case event, ok := <-eventsChan:
				if !ok {
					if len(batch) > 0 {
						flush(chStorage, &batch)
					}
					return
				}
				batch = append(batch, event)
				if len(batch) >= 1000 {
					flush(chStorage, &batch)
					ticker.Reset(5 * time.Second)
				}
			case <-ticker.C:
				if len(batch) > 0 {
					flush(chStorage, &batch)
				}
			}
		}
	}()

	fmt.Println("Worker started. Reading Kafka...")

Loop:
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nShutting down gracefully...")
			break Loop
		default:
			readCtx, cancel := context.WithTimeout(context.Background(), time.Second)
			msg, err := consumer.ReadMessage(readCtx)
			cancel()

			if err != nil {
				continue
			}

			var event model.Event
			if err := json.Unmarshal(msg.Value, &event); err == nil {
				eventsChan <- event
			}
		}
	}
	close(eventsChan)

	if err := consumer.Close(); err != nil {
		log.Printf("Error closing Kafka consumer: %v", err)
	}

	if err := chStorage.Close(); err != nil {
		log.Printf("ClickHouse connection close error: %v", err)
	}

	waitCtx, timeout := context.WithTimeout(context.Background(), 10*time.Second)
	defer timeout()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("Worker stopped cleanly.")
	case <-waitCtx.Done():
		fmt.Println("Shutdown timed out, some data might be lost.")
	}
}

func flush(storage *storage.ClickHouseStorage, batch *[]model.Event) {
	err := storage.InsertEvents(context.Background(), *batch)
	if err != nil {
		log.Printf("Failed to flush to ClickHouse: %v", err)
		return
	}

	fmt.Printf("Flushed %d events to ClickHouse\n", len(*batch))
	*batch = (*batch)[:0]
}
