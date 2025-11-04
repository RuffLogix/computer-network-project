package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/rufflogix/computer-network-project/internal/entity"
	"github.com/rufflogix/computer-network-project/internal/repository"
)

type InvitationService interface {
	// Chat invitations
	CreateChatInvitation(chatID, createdBy int64, expiresIn *time.Duration, maxUses *int) (*entity.ChatInvitation, error)
	ValidateChatInvitation(code string) (*entity.ChatInvitation, error)
	JoinChatViaInvitation(code string, userID int64) error
	GetChatInvitations(chatID int64) ([]*entity.ChatInvitation, error)

	// Friend invitations
	CreateFriendInvitation(userID int64, expiresIn *time.Duration, maxUses *int) (*entity.FriendInvitation, error)
	ValidateFriendInvitation(code string) (*entity.FriendInvitation, error)
	AcceptFriendInvitation(code string, userID int64) error
	GetFriendInvitations(userID int64) ([]*entity.FriendInvitation, error)
	SendFriendRequest(senderID int64, targetIdentifier string) error
	GetFriendships(userID int64) ([]*entity.Friendship, error)
}

type implInvitationService struct {
	invitationRepo  repository.InvitationRepository
	chatRepo        repository.ChatRepository
	friendshipRepo  repository.FriendshipRepository
	notificationSvc NotificationService
	userRepo        repository.UserRepository
}

func NewInvitationService(
	invitationRepo repository.InvitationRepository,
	chatRepo repository.ChatRepository,
	friendshipRepo repository.FriendshipRepository,
	notificationSvc NotificationService,
	userRepo repository.UserRepository,
) InvitationService {
	return &implInvitationService{
		invitationRepo:  invitationRepo,
		chatRepo:        chatRepo,
		friendshipRepo:  friendshipRepo,
		notificationSvc: notificationSvc,
		userRepo:        userRepo,
	}
}

// Chat invitations
func (s *implInvitationService) CreateChatInvitation(chatID, createdBy int64, expiresIn *time.Duration, maxUses *int) (*entity.ChatInvitation, error) {
	var expiresAt *time.Time
	if expiresIn != nil {
		expTime := time.Now().Add(*expiresIn)
		expiresAt = &expTime
	}

	return s.invitationRepo.CreateChatInvitation(chatID, createdBy, expiresAt, maxUses)
}

func (s *implInvitationService) ValidateChatInvitation(code string) (*entity.ChatInvitation, error) {
	return s.invitationRepo.GetChatInvitationByCode(code)
}

func (s *implInvitationService) JoinChatViaInvitation(code string, userID int64) error {
	invitation, err := s.invitationRepo.GetChatInvitationByCode(code)
	if err != nil {
		return err
	}

	members, err := s.chatRepo.GetChatMembers(invitation.ChatID)
	if err != nil {
		return err
	}

	// Add user to chat
	member := &entity.ChatMember{
		ChatID: invitation.ChatID,
		UserID: userID,
		Role:   "member",
	}

	if err := s.chatRepo.AddChatMember(member); err != nil {
		return err
	}

	// Increment usage count
	if err := s.invitationRepo.UseChatInvitation(code); err != nil {
		return err
	}

	// Notify existing members (excluding the new member)
	for _, existing := range members {
		if existing.UserID == userID {
			continue
		}

		notification := &entity.Notification{
			RecipientID: existing.UserID,
			SenderID:    userID,
			Type:        entity.GroupMemberJoined,
			Title:       "New member joined",
			Message:     "A new member has joined your private group",
			ReferenceID: &invitation.ChatID,
		}
		s.notificationSvc.SendNotification(notification)
	}

	return nil
}

// Friend invitations
func (s *implInvitationService) CreateFriendInvitation(userID int64, expiresIn *time.Duration, maxUses *int) (*entity.FriendInvitation, error) {
	var expiresAt *time.Time
	if expiresIn != nil {
		expTime := time.Now().Add(*expiresIn)
		expiresAt = &expTime
	}

	return s.invitationRepo.CreateFriendInvitation(userID, expiresAt, maxUses)
}

func (s *implInvitationService) ValidateFriendInvitation(code string) (*entity.FriendInvitation, error) {
	return s.invitationRepo.GetFriendInvitationByCode(code)
}

func (s *implInvitationService) AcceptFriendInvitation(code string, userID int64) error {
	invitation, err := s.invitationRepo.GetFriendInvitationByCode(code)
	if err != nil {
		return err
	}

	// Check if user is trying to accept their own invitation
	if invitation.UserID == userID {
		return fmt.Errorf("cannot accept your own friend invitation")
	}

	// Check if users are already friends or have a pending request (both directions)
	existingFriendship, err := s.friendshipRepo.GetFriendship(invitation.UserID, userID)
	if err == nil {
		if existingFriendship.Status == entity.Accepted {
			return fmt.Errorf("you are already friends with this user")
		}
		if existingFriendship.Status == entity.Pending {
			// Check who initiated the request
			if existingFriendship.UserID == userID {
				return fmt.Errorf("you already sent a friend request to this user")
			} else {
				return fmt.Errorf("this user already sent you a friend request, check your notifications")
			}
		}
	}

	// Create friendship request
	friendship, err := s.friendshipRepo.CreateFriendship(invitation.UserID, userID)
	if err != nil {
		return err
	}

	// Create notification for the invitation creator
	notification := &entity.Notification{
		RecipientID: invitation.UserID,
		SenderID:    userID,
		Type:        entity.FriendRequest,
		Title:       "New Friend Request",
		Message:     "Someone accepted your friend invitation",
		ReferenceID: &friendship.ID,
	}

	s.notificationSvc.SendNotification(notification)

	// Increment usage count
	return s.invitationRepo.UseFriendInvitation(code)
}

func (s *implInvitationService) GetChatInvitations(chatID int64) ([]*entity.ChatInvitation, error) {
	return s.invitationRepo.GetChatInvitationsByChat(chatID)
}

func (s *implInvitationService) GetFriendInvitations(userID int64) ([]*entity.FriendInvitation, error) {
	return s.invitationRepo.GetFriendInvitationsByUser(userID)
}

func (s *implInvitationService) SendFriendRequest(senderID int64, targetIdentifier string) error {
	// Find target user by ID or username
	var targetUser *entity.User
	var err error

	// Try to parse as numeric ID first
	if targetID, parseErr := strconv.ParseInt(targetIdentifier, 10, 64); parseErr == nil {
		targetUser, err = s.userRepo.GetUserByNumericID(targetID)
	} else {
		// Try as username
		targetUser, err = s.userRepo.GetUserByUsername(targetIdentifier)
	}

	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Check if sender is trying to add themselves
	if targetUser.NumericID == senderID {
		return fmt.Errorf("cannot add yourself as a friend")
	}

	// Check if users are already friends or have a pending request (both directions)
	existingFriendship, err := s.friendshipRepo.GetFriendship(senderID, targetUser.NumericID)
	if err == nil {
		if existingFriendship.Status == entity.Accepted {
			return fmt.Errorf("you are already friends with this user")
		}
		if existingFriendship.Status == entity.Pending {
			// Check who initiated the request
			if existingFriendship.UserID == senderID {
				return fmt.Errorf("you already sent a friend request to this user")
			} else {
				return fmt.Errorf("this user already sent you a friend request, check your notifications")
			}
		}
	}

	// Create friendship request
	friendship, err := s.friendshipRepo.CreateFriendship(senderID, targetUser.NumericID)
	if err != nil {
		return err
	}

	// Create notification for the target user
	notification := &entity.Notification{
		RecipientID: targetUser.NumericID,
		SenderID:    senderID,
		Type:        entity.FriendRequest,
		Title:       "New Friend Request",
		Message:     "Someone sent you a friend request",
		ReferenceID: &friendship.ID,
	}

	s.notificationSvc.SendNotification(notification)

	return nil
}

func (s *implInvitationService) GetFriendships(userID int64) ([]*entity.Friendship, error) {
	friendships, err := s.friendshipRepo.GetFriendshipsByUser(userID)
	if err != nil {
		return nil, err
	}

	// Populate user data for each friendship
	for _, friendship := range friendships {
		if friendship.UserID == userID {
			// Get friend's data
			friend, err := s.userRepo.GetUserByNumericID(friendship.FriendID)
			if err == nil {
				friendship.Friend = friend
			}
		} else {
			// Get user's data
			user, err := s.userRepo.GetUserByNumericID(friendship.UserID)
			if err == nil {
				friendship.User = user
			}
		}
	}

	return friendships, nil
}
