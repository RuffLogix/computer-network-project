package entity

type EventType string

const (
	JOIN         EventType = "join"
	SEND_MESSAGE EventType = "send_message"
)

type Event struct {
	Type      EventType `json:"type"`
	Content   string    `json:"content"`
	CreatedBy int64     `json:"created_by"`
}
