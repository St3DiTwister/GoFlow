package handler

import (
	"GoFlow/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type KafkaProducer interface {
	Publish(ctx context.Context, key, value []byte) error
}

type EventHandler struct {
	Producer KafkaProducer
}

func (h *EventHandler) Track(c *gin.Context) {
	var event model.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	event.EventID = uuid.NewString()
	event.ReceivedAt = time.Now()

	fmt.Printf("Получено событие: %s от сайта %s\n", event.Type, event.SiteID)

	payload, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("Ошибка маршалинга: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process event"})
		return
	}
	err = h.Producer.Publish(c, []byte(event.SiteID), payload)
	if err != nil {
		fmt.Printf("Ошибка отправки в Kafka: %v\n", err)
	}

	c.Status(http.StatusAccepted)
}
