package handler

import (
	"GoFlow/internal/model"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type EventHandler struct {
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

	c.Status(http.StatusAccepted)
}
