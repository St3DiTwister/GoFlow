package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	EventID    string                 `json:"event_id"`
	SiteID     string                 `json:"site_id"`
	Type       string                 `json:"type"`
	UserID     string                 `json:"user_id"`
	Path       string                 `json:"path"`
	Timestamp  time.Time              `json:"timestamp"`
	Properties map[string]interface{} `json:"properties"`
}

var (
	eventTypes = []string{"page_view", "click", "scroll", "form_submit"}
	paths      = []string{"/", "/products", "/cart", "/checkout", "/pricing"}
	siteIDs    = []string{"fe1988f2-53b8-4505-90e1-29da5b8eeace", "b2744901-013f-4cf4-88da-1ea4d53d5248", "saas-app-main"}
)

func main() {
	var count int
	var duration int

	fmt.Print("Сколько запросов отправить? ")
	fmt.Scan(&count)
	fmt.Print("За сколько секунд их распределить? ")
	fmt.Scan(&duration)

	url := "http://localhost:80/track"
	delay := time.Duration(duration*1000/count) * time.Millisecond

	fmt.Printf("🚀 Начинаю отправку %d запросов с задержкой %v...\n", count, delay)

	for i := 0; i < count; i++ {
		event := generateRandomEvent()
		sendEvent(url, event)
		time.Sleep(delay)
	}

	fmt.Println("\n✅ Все запросы отправлены!")
}

func generateRandomEvent() Event {
	return Event{
		EventID:   uuid.NewString(),
		SiteID:    siteIDs[rand.Intn(len(siteIDs))],
		Type:      eventTypes[rand.Intn(len(eventTypes))],
		UserID:    fmt.Sprintf("user-%d", rand.Intn(1000)),
		Path:      paths[rand.Intn(len(paths))],
		Timestamp: time.Now(),
		Properties: map[string]interface{}{
			"browser": "Chrome",
			"os":      "Windows",
		},
	}
}

func sendEvent(url string, event Event) {
	jsonData, _ := json.Marshal(event)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Ошибка: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("[%s] Отправлен %s на %s (Status: %d)\n", time.Now().Format("15:04:05"), event.Type, event.Path, resp.StatusCode)
}
