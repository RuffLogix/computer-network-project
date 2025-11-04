package entity

import "time"

type ChatType string

const (
	Individual   ChatType = "individual"
	PrivateGroup ChatType = "private_group"
	PublicGroup  ChatType = "public_group"
)

type Chat struct {
	ID          int64     `bson:"id" json:"id"`
	Type        ChatType  `bson:"type" json:"type"`
	Name        string    `bson:"name" json:"name"`
	Description string    `bson:"description,omitempty" json:"description,omitempty"`
	IsPublic    bool      `bson:"is_public" json:"is_public"`
	CreatedBy   int64     `bson:"created_by" json:"created_by"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}

type MessageType string

const (
	Text   MessageType = "text"
	Image  MessageType = "image"
	Video  MessageType = "video"
	System MessageType = "system"
)

type Message struct {
	ID            int64       `bson:"id" json:"id"`
	ChatID        int64       `bson:"chat_id" json:"chat_id"`
	Content       string      `bson:"content" json:"content"`
	Type          MessageType `bson:"type" json:"type"`
	MediaURL      string      `bson:"media_url,omitempty" json:"media_url,omitempty"`
	ReplyToID     *int64      `bson:"reply_to_id,omitempty" json:"reply_to_id,omitempty"`
	ReplyTo       *Message    `bson:"-" json:"reply_to,omitempty"`
	Reactions     []*Reaction `bson:"-" json:"reactions,omitempty"`
	CreatedAt     time.Time   `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time   `bson:"updated_at" json:"updated_at"`
	CreatedBy     int64       `bson:"created_by" json:"created_by"`
	CreatedByUser *User       `bson:"-" json:"created_by_user,omitempty"`
}

type ReactionType string

const (
	Like  ReactionType = "like"
	Love  ReactionType = "love"
	Laugh ReactionType = "laugh"
	Wow   ReactionType = "wow"
	Sad   ReactionType = "sad"
	Angry ReactionType = "angry"
)

type Reaction struct {
	ID        int64        `bson:"id" json:"id"`
	MessageID int64        `bson:"message_id" json:"message_id"`
	Type      ReactionType `bson:"type" json:"type"`
	Count     int          `bson:"count" json:"count"`
	UserIDs   []int64      `bson:"user_ids" json:"user_ids"`
	CreatedAt time.Time    `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time    `bson:"updated_at" json:"updated_at"`
}

type ChatInvitation struct {
	ID        int64      `bson:"id" json:"id"`
	ChatID    int64      `bson:"chat_id" json:"chat_id"`
	Code      string     `bson:"code" json:"code"`
	ExpiresAt *time.Time `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	MaxUses   *int       `bson:"max_uses,omitempty" json:"max_uses,omitempty"`
	UsedCount int        `bson:"used_count" json:"used_count"`
	CreatedBy int64      `bson:"created_by" json:"created_by"`
	CreatedAt time.Time  `bson:"created_at" json:"created_at"`
	IsActive  bool       `bson:"is_active" json:"is_active"`
}

type ChatMember struct {
	ID       int64     `bson:"id,omitempty" json:"id"`
	ChatID   int64     `bson:"chat_id" json:"chat_id"`
	UserID   int64     `bson:"user_id" json:"user_id"`
	Role     string    `bson:"role" json:"role"` // admin, member
	JoinedAt time.Time `bson:"joined_at" json:"joined_at"`
}
