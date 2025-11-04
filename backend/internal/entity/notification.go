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
	ID          int64              `bson:"id" json:"id"`
	RecipientID int64              `bson:"recipient_id" json:"recipient_id"`
	SenderID    int64              `bson:"sender_id" json:"sender_id"`
	Type        NotificationType   `bson:"type" json:"type"`
	Status      NotificationStatus `bson:"status" json:"status"`
	Title       string             `bson:"title" json:"title"`
	Message     string             `bson:"message" json:"message"`
	ReferenceID *int64             `bson:"reference_id,omitempty" json:"reference_id,omitempty"` // Can reference chat, message, friendship, etc.
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	Sender      *User              `bson:"-" json:"sender,omitempty"`
}
