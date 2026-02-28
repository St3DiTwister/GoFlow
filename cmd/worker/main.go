package main

import (
	"GoFlow/internal/kafka"
	"GoFlow/internal/model"
	"GoFlow/internal/storage"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

func main() {
	brokers := []string{"localhost:9092"}

	topic := "telemetry"
	groupID := "event-processor-v1"

	consumer := kafka.NewConsumer(brokers, topic, groupID)
	defer consumer.Close()

	chStorage, _ := storage.NewClickHouseStorage("localhost:9000")

	eventsChan := make(chan model.Event, 1000)

	go func() {
		batch := make([]model.Event, 0, 1000)
		ticker := time.NewTicker(5 * time.Second)

		for {
			select {
			case event := <-eventsChan:
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

	for {
		msg, err := consumer.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Kafka read error: %v", err)
			continue
		}

		var event model.Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Unmarshal error: %v", err)
			continue
		}

		eventsChan <- event
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
