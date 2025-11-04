package repository

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/rufflogix/computer-network-project/internal/entity"
)

type InvitationRepository interface {
	// Chat invitations
	CreateChatInvitation(chatID int64, createdBy int64, expiresAt *time.Time, maxUses *int) (*entity.ChatInvitation, error)
	GetChatInvitationByCode(code string) (*entity.ChatInvitation, error)
	UseChatInvitation(code string) error
	DeactivateChatInvitation(code string) error
	GetChatInvitationsByChat(chatID int64) ([]*entity.ChatInvitation, error)

	// Friend invitations
	CreateFriendInvitation(userID int64, expiresAt *time.Time, maxUses *int) (*entity.FriendInvitation, error)
	GetFriendInvitationByCode(code string) (*entity.FriendInvitation, error)
	UseFriendInvitation(code string) error
	DeactivateFriendInvitation(code string) error
	GetFriendInvitationsByUser(userID int64) ([]*entity.FriendInvitation, error)
}

type implInvitationRepository struct {
	chatInvitations    map[string]*entity.ChatInvitation
	friendInvitations  map[string]*entity.FriendInvitation
	mu                 sync.RWMutex
	chatInvitationID   int64
	friendInvitationID int64
}

func NewInvitationRepository() InvitationRepository {
	return &implInvitationRepository{
		chatInvitations:   make(map[string]*entity.ChatInvitation),
		friendInvitations: make(map[string]*entity.FriendInvitation),
	}
}

func generateInviteCode() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// Chat Invitations
func (r *implInvitationRepository) CreateChatInvitation(chatID int64, createdBy int64, expiresAt *time.Time, maxUses *int) (*entity.ChatInvitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.chatInvitationID++
	code := generateInviteCode()

	invitation := &entity.ChatInvitation{
		ID:        r.chatInvitationID,
		ChatID:    chatID,
		Code:      code,
		ExpiresAt: expiresAt,
		MaxUses:   maxUses,
		UsedCount: 0,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	r.chatInvitations[code] = invitation
	return invitation, nil
}

func (r *implInvitationRepository) GetChatInvitationByCode(code string) (*entity.ChatInvitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	invitation, ok := r.chatInvitations[code]
	if !ok {
		return nil, fmt.Errorf("invitation not found")
	}

	if !invitation.IsActive {
		return nil, fmt.Errorf("invitation is not active")
	}

	if invitation.ExpiresAt != nil && time.Now().After(*invitation.ExpiresAt) {
		return nil, fmt.Errorf("invitation has expired")
	}

	if invitation.MaxUses != nil && invitation.UsedCount >= *invitation.MaxUses {
		return nil, fmt.Errorf("invitation has reached maximum uses")
	}

	return invitation, nil
}

func (r *implInvitationRepository) UseChatInvitation(code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	invitation, ok := r.chatInvitations[code]
	if !ok {
		return fmt.Errorf("invitation not found")
	}

	invitation.UsedCount++
	return nil
}

func (r *implInvitationRepository) DeactivateChatInvitation(code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	invitation, ok := r.chatInvitations[code]
	if !ok {
		return fmt.Errorf("invitation not found")
	}

	invitation.IsActive = false
	return nil
}

func (r *implInvitationRepository) GetChatInvitationsByChat(chatID int64) ([]*entity.ChatInvitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var invitations []*entity.ChatInvitation
	for _, inv := range r.chatInvitations {
		if inv.ChatID == chatID {
			invitations = append(invitations, inv)
		}
	}

	return invitations, nil
}

// Friend Invitations
func (r *implInvitationRepository) CreateFriendInvitation(userID int64, expiresAt *time.Time, maxUses *int) (*entity.FriendInvitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.friendInvitationID++
	code := generateInviteCode()

	invitation := &entity.FriendInvitation{
		ID:        r.friendInvitationID,
		Code:      code,
		UserID:    userID,
		ExpiresAt: expiresAt,
		MaxUses:   maxUses,
		UsedCount: 0,
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	r.friendInvitations[code] = invitation
	return invitation, nil
}

func (r *implInvitationRepository) GetFriendInvitationByCode(code string) (*entity.FriendInvitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	invitation, ok := r.friendInvitations[code]
	if !ok {
		return nil, fmt.Errorf("invitation not found")
	}

	if !invitation.IsActive {
		return nil, fmt.Errorf("invitation is not active")
	}

	if invitation.ExpiresAt != nil && time.Now().After(*invitation.ExpiresAt) {
		return nil, fmt.Errorf("invitation has expired")
	}

	if invitation.MaxUses != nil && invitation.UsedCount >= *invitation.MaxUses {
		return nil, fmt.Errorf("invitation has reached maximum uses")
	}

	return invitation, nil
}

func (r *implInvitationRepository) UseFriendInvitation(code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	invitation, ok := r.friendInvitations[code]
	if !ok {
		return fmt.Errorf("invitation not found")
	}

	invitation.UsedCount++
	return nil
}

func (r *implInvitationRepository) DeactivateFriendInvitation(code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	invitation, ok := r.friendInvitations[code]
	if !ok {
		return fmt.Errorf("invitation not found")
	}

	invitation.IsActive = false
	return nil
}

func (r *implInvitationRepository) GetFriendInvitationsByUser(userID int64) ([]*entity.FriendInvitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var invitations []*entity.FriendInvitation
	for _, inv := range r.friendInvitations {
		if inv.UserID == userID {
			invitations = append(invitations, inv)
		}
	}

	return invitations, nil
}
