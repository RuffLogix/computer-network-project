package service

import (
	"github.com/rufflogix/computer-network-project/internal/entity"
	"github.com/rufflogix/computer-network-project/internal/repository"
)

type ChatService interface {
	// Chat operations
	CreateChat(chat *entity.Chat) error
	GetChat(id int64) (*entity.Chat, error)
	GetUserChats(userID int64) ([]*entity.Chat, error)
	GetPublicChats() ([]*entity.Chat, error)
	GetAllChats() ([]*entity.Chat, error)

	// Message operations
	SendMessage(message *entity.Message) error
	GetMessages(chatID int64, limit, offset int) ([]*entity.Message, error)
	EditMessage(messageID int64, content string) error
	DeleteMessage(messageID int64) error

	// Reaction operations
	AddReaction(reaction *entity.Reaction, userID int64) error
	RemoveReaction(reactionID int64, userID int64) error
	GetMessageReactions(messageID int64) ([]*entity.Reaction, error)

	// Member operations
	AddMember(chatID, userID int64, role string) error
	RemoveMember(chatID, userID int64) error
	GetMembers(chatID int64) ([]*entity.ChatMember, error)
}

type implChatService struct {
	chatRepository repository.ChatRepository
	userRepository repository.UserRepository
}

func NewChatService(chatRepository repository.ChatRepository, userRepository repository.UserRepository) ChatService {
	return &implChatService{
		chatRepository: chatRepository,
		userRepository: userRepository,
	}
}

// Chat operations
func (s *implChatService) CreateChat(chat *entity.Chat) error {
	return s.chatRepository.CreateChat(chat)
}

func (s *implChatService) GetChat(id int64) (*entity.Chat, error) {
	return s.chatRepository.GetChatByID(id)
}

func (s *implChatService) GetUserChats(userID int64) ([]*entity.Chat, error) {
	chats, err := s.chatRepository.GetChatsByUser(userID)
	if err != nil {
		return nil, err
	}

	// For individual chats, set the name to the other participant's name
	for _, chat := range chats {
		if chat.Type == entity.Individual {
			members, err := s.GetMembers(chat.ID)
			if err != nil {
				continue // Keep the existing name if we can't get members
			}

			// Find the other member
			for _, member := range members {
				if member.UserID != userID {
					// Get the other user's information
					otherUser, err := s.userRepository.GetUserByNumericID(member.UserID)
					if err != nil {
						continue
					}
					// Use username for display, fallback to name
					displayName := otherUser.Username
					if displayName == "" {
						displayName = otherUser.Name
					}
					chat.Name = displayName
					break
				}
			}
		}
	}

	return chats, nil
}

func (s *implChatService) GetPublicChats() ([]*entity.Chat, error) {
	return s.chatRepository.GetPublicChats()
}

func (s *implChatService) GetAllChats() ([]*entity.Chat, error) {
	return s.chatRepository.GetAllChats()
}

// Message operations
func (s *implChatService) SendMessage(message *entity.Message) error {
	if err := s.chatRepository.CreateMessage(message); err != nil {
		return err
	}

	if message.CreatedBy != 0 {
		if user, err := s.userRepository.GetUserByNumericID(message.CreatedBy); err == nil {
			message.CreatedByUser = user
		}
	}

	return nil
}

func (s *implChatService) GetMessages(chatID int64, limit, offset int) ([]*entity.Message, error) {
	messages, err := s.chatRepository.GetMessagesByChat(chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Populate user information for each message
	for _, message := range messages {
		if message.CreatedBy != 0 {
			user, err := s.userRepository.GetUserByNumericID(message.CreatedBy)
			if err == nil {
				message.CreatedByUser = user
			}
		}
	}

	return messages, nil
}

func (s *implChatService) EditMessage(messageID int64, content string) error {
	message, err := s.chatRepository.GetMessageByID(messageID)
	if err != nil {
		return err
	}

	message.Content = content
	return s.chatRepository.UpdateMessage(message)
}

func (s *implChatService) DeleteMessage(messageID int64) error {
	return s.chatRepository.DeleteMessage(messageID)
}

// Reaction operations
func (s *implChatService) AddReaction(reaction *entity.Reaction, userID int64) error {
	// For the new count-based model, we don't need toggle logic here
	// The repository handles adding users to reaction counts
	return s.chatRepository.CreateReaction(reaction, userID)
}

func (s *implChatService) RemoveReaction(reactionID int64, userID int64) error {
	return s.chatRepository.DeleteReaction(reactionID)
}

func (s *implChatService) GetMessageReactions(messageID int64) ([]*entity.Reaction, error) {
	return s.chatRepository.GetReactionsByMessage(messageID)
}

// Member operations
func (s *implChatService) AddMember(chatID, userID int64, role string) error {
	member := &entity.ChatMember{
		ChatID: chatID,
		UserID: userID,
		Role:   role,
	}
	return s.chatRepository.AddChatMember(member)
}

func (s *implChatService) RemoveMember(chatID, userID int64) error {
	return s.chatRepository.RemoveChatMember(chatID, userID)
}

func (s *implChatService) GetMembers(chatID int64) ([]*entity.ChatMember, error) {
	return s.chatRepository.GetChatMembers(chatID)
}
