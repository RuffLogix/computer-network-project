package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	NumericID int64              `bson:"numeric_id" json:"numeric_id"` // Numeric ID for chat operations
	Username  string             `bson:"username" json:"username"`
	Password  string             `bson:"password" json:"-"` // Never send password in JSON
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Avatar    string             `bson:"avatar,omitempty" json:"avatar,omitempty"`
	IsGuest   bool               `bson:"is_guest" json:"is_guest"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type FriendshipStatus string

const (
	Pending  FriendshipStatus = "pending"
	Accepted FriendshipStatus = "accepted"
	Rejected FriendshipStatus = "rejected"
	Blocked  FriendshipStatus = "blocked"
)

type Friendship struct {
	ID        int64            `json:"id"`
	UserID    int64            `json:"user_id"`
	FriendID  int64            `json:"friend_id"`
	Status    FriendshipStatus `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	User      *User            `json:"user,omitempty"`
	Friend    *User            `json:"friend,omitempty"`
}

type FriendInvitation struct {
	ID        int64      `json:"id"`
	Code      string     `json:"code"`
	UserID    int64      `json:"user_id"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	MaxUses   *int       `json:"max_uses,omitempty"`
	UsedCount int        `json:"used_count"`
	CreatedAt time.Time  `json:"created_at"`
	IsActive  bool       `json:"is_active"`
}
