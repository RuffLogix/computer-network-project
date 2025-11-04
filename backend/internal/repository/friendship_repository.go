package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/rufflogix/computer-network-project/internal/entity"
)

type FriendshipRepository interface {
	CreateFriendship(userID, friendID int64) (*entity.Friendship, error)
	GetFriendship(userID, friendID int64) (*entity.Friendship, error)
	GetFriendshipsByUser(userID int64) ([]*entity.Friendship, error)
	GetPendingFriendships(userID int64) ([]*entity.Friendship, error)
	UpdateFriendshipStatus(userID, friendID int64, status entity.FriendshipStatus) error
	DeleteFriendship(userID, friendID int64) error
}

type implFriendshipRepository struct {
	friendships  map[int64]*entity.Friendship
	friendshipID int64
	mu           sync.RWMutex
}

func NewFriendshipRepository() FriendshipRepository {
	return &implFriendshipRepository{
		friendships: make(map[int64]*entity.Friendship),
	}
}

func (r *implFriendshipRepository) CreateFriendship(userID, friendID int64) (*entity.Friendship, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.friendshipID++
	friendship := &entity.Friendship{
		ID:        r.friendshipID,
		UserID:    userID,
		FriendID:  friendID,
		Status:    entity.Pending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	r.friendships[friendship.ID] = friendship
	return friendship, nil
}

func (r *implFriendshipRepository) GetFriendship(userID, friendID int64) (*entity.Friendship, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, friendship := range r.friendships {
		if (friendship.UserID == userID && friendship.FriendID == friendID) ||
			(friendship.UserID == friendID && friendship.FriendID == userID) {
			return friendship, nil
		}
	}

	return nil, fmt.Errorf("friendship not found")
}

func (r *implFriendshipRepository) GetFriendshipsByUser(userID int64) ([]*entity.Friendship, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var friendships []*entity.Friendship
	for _, friendship := range r.friendships {
		if (friendship.UserID == userID || friendship.FriendID == userID) && friendship.Status == entity.Accepted {
			friendships = append(friendships, friendship)
		}
	}

	return friendships, nil
}

func (r *implFriendshipRepository) GetPendingFriendships(userID int64) ([]*entity.Friendship, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var friendships []*entity.Friendship
	for _, friendship := range r.friendships {
		if friendship.FriendID == userID && friendship.Status == entity.Pending {
			friendships = append(friendships, friendship)
		}
	}

	return friendships, nil
}

func (r *implFriendshipRepository) UpdateFriendshipStatus(userID, friendID int64, status entity.FriendshipStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, friendship := range r.friendships {
		if (friendship.UserID == userID && friendship.FriendID == friendID) ||
			(friendship.UserID == friendID && friendship.FriendID == userID) {
			friendship.Status = status
			friendship.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("friendship not found")
}

func (r *implFriendshipRepository) DeleteFriendship(userID, friendID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, friendship := range r.friendships {
		if (friendship.UserID == userID && friendship.FriendID == friendID) ||
			(friendship.UserID == friendID && friendship.FriendID == userID) {
			delete(r.friendships, id)
			return nil
		}
	}

	return fmt.Errorf("friendship not found")
}
