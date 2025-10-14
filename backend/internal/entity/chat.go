package entity

import "time"

type ChatType string

const (
	Individual   ChatType = "individual"
	PrivateGroup ChatType = "private_group"
	PublicGroup  ChatType = "public_group"
)

type Chat struct {
	ID        int64     `json:"id"`
	Type      ChatType  `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MessageType string

const (
	Text MessageType = "text"
)

type Message struct {
	ID        int64       `json:"id"`
	Content   string      `json:"content"`
	Type      MessageType `json:"type"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	CreatedBy int64       `json:"created_by"`
}

type ChatInvitation struct {
	ID        int64  `json:""`
	Code      string `json:"code"`
	CreatedBy int64  `json:"created_by"`
}
