package entity

import "time"

type NotificationType string

const (
	FriendRequest     NotificationType = "friend_request"
	FriendAccepted    NotificationType = "friend_accepted"
	GroupInvitation   NotificationType = "group_invitation"
	MessageReaction   NotificationType = "message_reaction"
	MessageReply      NotificationType = "message_reply"
	GroupMemberJoined NotificationType = "group_member_joined"
)

type NotificationStatus string

const (
	NotificationUnread   NotificationStatus = "unread"
	NotificationRead     NotificationStatus = "read"
	NotificationAccepted NotificationStatus = "accepted"
	NotificationRejected NotificationStatus = "rejected"
)

type Notification struct {
	ID          int64              `json:"id"`
	RecipientID int64              `json:"recipient_id"`
	SenderID    int64              `json:"sender_id"`
	Type        NotificationType   `json:"type"`
	Status      NotificationStatus `json:"status"`
	Title       string             `json:"title"`
	Message     string             `json:"message"`
	ReferenceID *int64             `json:"reference_id,omitempty"` // Can reference chat, message, friendship, etc.
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	Sender      *User              `json:"sender,omitempty"`
}
