package model

import "time"

type Event struct {
	EventID    string                 `json:"event_id"`    // UUID
	SiteID     string                 `json:"site_id"`     // ID проекта
	Type       string                 `json:"type"`        // click, view, scroll
	UserID     string                 `json:"user_id"`     // ID юзера или сессии
	Path       string                 `json:"path"`        // /home, /cart
	Timestamp  time.Time              `json:"timestamp"`   // Время на клиенте
	ReceivedAt time.Time              `json:"received_at"` // Время прихода на сервер
	Properties map[string]interface{} `json:"properties"`  // Доп. данные
}
