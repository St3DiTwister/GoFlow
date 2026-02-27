package main

import (
	"GoFlow/internal/kafka"
	"GoFlow/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"log"
)

func main() {
	brokers := []string{"localhost:9092"}

	topic := "telemetry"
	groupID := "event-processor-v1"

	consumer := kafka.NewConsumer(brokers, topic, groupID)
	defer consumer.Close()

	fmt.Println("Worker started. Waiting for messages")

	ctx := context.Background()
	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		var event model.Event
		err = json.Unmarshal(msg.Value, &event)
		if err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}

		fmt.Printf("Worker [Partition %d]: Received event %s from site %s\n",
			msg.Partition, event.EventID, event.SiteID)
	}
}
