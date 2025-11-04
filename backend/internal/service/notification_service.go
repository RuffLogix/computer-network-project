package service

import (
	"fmt"

	"github.com/rufflogix/computer-network-project/internal/entity"
	"github.com/rufflogix/computer-network-project/internal/repository"
)

type NotificationService interface {
	SendNotification(notification *entity.Notification) error
	GetUserNotifications(userID int64) ([]*entity.Notification, error)
	GetUnreadNotifications(userID int64) ([]*entity.Notification, error)
	MarkAsRead(notificationID int64) error
	AcceptNotification(notificationID, userID int64) error
	RejectNotification(notificationID, userID int64) error
}

type implNotificationService struct {
	notificationRepo repository.NotificationRepository
	friendshipRepo   repository.FriendshipRepository
	chatRepo         repository.ChatRepository
	userRepo         repository.UserRepository
	roomService      RoomService
}

func NewNotificationService(
	notificationRepo repository.NotificationRepository,
	friendshipRepo repository.FriendshipRepository,
	chatRepo repository.ChatRepository,
	userRepo repository.UserRepository,
	roomService RoomService,
) NotificationService {
	return &implNotificationService{
		notificationRepo: notificationRepo,
		friendshipRepo:   friendshipRepo,
		chatRepo:         chatRepo,
		userRepo:         userRepo,
		roomService:      roomService,
	}
}

func (s *implNotificationService) SendNotification(notification *entity.Notification) error {
	if err := s.notificationRepo.CreateNotification(notification); err != nil {
		return err
	}

	s.emitNotification(notification)

	return nil
}

func (s *implNotificationService) GetUserNotifications(userID int64) ([]*entity.Notification, error) {
	return s.notificationRepo.GetNotificationsByUser(userID)
}

func (s *implNotificationService) GetUnreadNotifications(userID int64) ([]*entity.Notification, error) {
	return s.notificationRepo.GetUnreadNotificationsByUser(userID)
}

func (s *implNotificationService) MarkAsRead(notificationID int64) error {
	return s.notificationRepo.UpdateNotificationStatus(notificationID, entity.NotificationRead)
}

func (s *implNotificationService) AcceptNotification(notificationID, userID int64) error {
	notification, err := s.notificationRepo.GetNotificationByID(notificationID)
	if err != nil {
		return err
	}

	switch notification.Type {
	case entity.FriendRequest:
		// Check if notification is already accepted
		if notification.Status == entity.NotificationAccepted {
			return fmt.Errorf("friend request already accepted")
		}

		// Check if users are already friends
		existingFriendship, err := s.friendshipRepo.GetFriendship(notification.SenderID, userID)
		if err == nil && existingFriendship.Status == entity.Accepted {
			// Update notification status to accepted
			s.notificationRepo.UpdateNotificationStatus(notificationID, entity.NotificationAccepted)
			return fmt.Errorf("you are already friends")
		}

		// Accept friend request
		err = s.friendshipRepo.UpdateFriendshipStatus(notification.SenderID, userID, entity.Accepted)
		if err != nil {
			return err
		}

		// Check if an individual chat already exists between these users
		chatExists := false
		userChats, err := s.chatRepo.GetChatsByUser(userID)
		if err == nil {
			for _, chat := range userChats {
				if chat.Type == entity.Individual {
					members, err := s.chatRepo.GetChatMembers(chat.ID)
					if err == nil {
						// Check if both users are members
						hasSender := false
						hasRecipient := false
						for _, member := range members {
							if member.UserID == notification.SenderID {
								hasSender = true
							}
							if member.UserID == userID {
								hasRecipient = true
							}
						}
						if hasSender && hasRecipient {
							chatExists = true
							break
						}
					}
				}
			}
		}

		// Create private chat between the two users only if one doesn't exist
		if !chatExists {
			chat := &entity.Chat{
				Type:      entity.Individual,
				Name:      "", // Name will be set dynamically when fetching chats
				IsPublic:  false,
				CreatedBy: userID,
			}
			if err := s.chatRepo.CreateChat(chat); err != nil {
				// Log error but don't fail the friend request acceptance
				// TODO: Add proper logging
			} else {
				// Add both users as members
				member1 := &entity.ChatMember{
					ChatID: chat.ID,
					UserID: userID,
					Role:   "member",
				}
				member2 := &entity.ChatMember{
					ChatID: chat.ID,
					UserID: notification.SenderID,
					Role:   "member",
				}
				s.chatRepo.AddChatMember(member1)
				s.chatRepo.AddChatMember(member2)
			}
		}

		// Send acceptance notification back
		acceptNotif := &entity.Notification{
			RecipientID: notification.SenderID,
			SenderID:    userID,
			Type:        entity.FriendAccepted,
			Title:       "Friend Request Accepted",
			Message:     "Your friend request was accepted",
		}
		s.SendNotification(acceptNotif)

		// Update notification status to accepted
		if err := s.notificationRepo.UpdateNotificationStatus(notificationID, entity.NotificationAccepted); err != nil {
			return err
		}

	case entity.GroupInvitation:
		// Add user to group
		if notification.ReferenceID != nil {
			member := &entity.ChatMember{
				ChatID: *notification.ReferenceID,
				UserID: userID,
				Role:   "member",
			}
			err = s.chatRepo.AddChatMember(member)
			if err != nil {
				return err
			}
		}
	}

	if err := s.notificationRepo.UpdateNotificationStatus(notificationID, entity.NotificationAccepted); err != nil {
		return err
	}

	notification.Status = entity.NotificationAccepted
	s.emitNotification(notification)

	return nil
}

func (s *implNotificationService) RejectNotification(notificationID, userID int64) error {
	notification, err := s.notificationRepo.GetNotificationByID(notificationID)
	if err != nil {
		return err
	}

	switch notification.Type {
	case entity.FriendRequest:
		// Reject friend request
		err = s.friendshipRepo.UpdateFriendshipStatus(notification.SenderID, userID, entity.Rejected)
		if err != nil {
			return err
		}
	}

	if err := s.notificationRepo.UpdateNotificationStatus(notificationID, entity.NotificationRejected); err != nil {
		return err
	}

	notification.Status = entity.NotificationRejected
	s.emitNotification(notification)

	return nil
}

func (s *implNotificationService) emitNotification(notification *entity.Notification) {
	if s.roomService == nil {
		return
	}

	event := entity.Event{
		Type: entity.NOTIFICATION,
		Data: map[string]interface{}{
			"notification": notification,
		},
		CreatedBy: notification.SenderID,
	}

	s.roomService.SendToUser(notification.RecipientID, event)
}
