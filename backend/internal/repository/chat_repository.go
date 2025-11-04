package repository

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/rufflogix/computer-network-project/internal/entity"
)

type ChatRepository interface {
	// Chat operations
	CreateChat(chat *entity.Chat) error
	GetChatByID(id int64) (*entity.Chat, error)
	GetChatsByUser(userID int64) ([]*entity.Chat, error)
	GetPublicChats() ([]*entity.Chat, error)
	GetAllChats() ([]*entity.Chat, error)
	UpdateChat(chat *entity.Chat) error
	DeleteChat(id int64) error

	// Message operations
	CreateMessage(message *entity.Message) error
	GetMessageByID(id int64) (*entity.Message, error)
	GetMessagesByChat(chatID int64, limit, offset int) ([]*entity.Message, error)
	UpdateMessage(message *entity.Message) error
	DeleteMessage(id int64) error

	// Reaction operations
	CreateReaction(reaction *entity.Reaction, userID int64) error
	GetReactionsByMessage(messageID int64) ([]*entity.Reaction, error)
	DeleteReaction(id int64) error

	// Chat member operations
	AddChatMember(member *entity.ChatMember) error
	GetChatMembers(chatID int64) ([]*entity.ChatMember, error)
	RemoveChatMember(chatID, userID int64) error
	IsChatMember(chatID, userID int64) (bool, error)
}

type implChatRepository struct {
	chats        map[int64]*entity.Chat
	messages     map[int64]*entity.Message
	reactions    map[int64]*entity.Reaction
	chatMembers  map[int64][]*entity.ChatMember
	chatID       int64
	messageID    int64
	reactionID   int64
	chatMemberID int64
	mu           sync.RWMutex
}

func NewChatRepository() ChatRepository {
	return &implChatRepository{
		chats:       make(map[int64]*entity.Chat),
		messages:    make(map[int64]*entity.Message),
		reactions:   make(map[int64]*entity.Reaction),
		chatMembers: make(map[int64][]*entity.ChatMember),
	}
}

// Chat operations
func (r *implChatRepository) CreateChat(chat *entity.Chat) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.chatID++
	chat.ID = r.chatID
	chat.CreatedAt = time.Now()
	chat.UpdatedAt = time.Now()

	r.chats[chat.ID] = chat
	return nil
}

func (r *implChatRepository) GetChatByID(id int64) (*entity.Chat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chat, ok := r.chats[id]
	if !ok {
		return nil, fmt.Errorf("chat not found")
	}

	return chat, nil
}

func (r *implChatRepository) GetChatsByUser(userID int64) ([]*entity.Chat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var chats []*entity.Chat
	for _, members := range r.chatMembers {
		for _, member := range members {
			if member.UserID == userID {
				if chat, ok := r.chats[member.ChatID]; ok {
					chats = append(chats, chat)
				}
				break
			}
		}
	}

	return chats, nil
}

func (r *implChatRepository) GetPublicChats() ([]*entity.Chat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var chats []*entity.Chat
	for _, chat := range r.chats {
		if chat.IsPublic {
			chats = append(chats, chat)
		}
	}

	return chats, nil
}

func (r *implChatRepository) GetAllChats() ([]*entity.Chat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var chats []*entity.Chat
	for _, chat := range r.chats {
		chats = append(chats, chat)
	}

	return chats, nil
}

func (r *implChatRepository) UpdateChat(chat *entity.Chat) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.chats[chat.ID]; !ok {
		return fmt.Errorf("chat not found")
	}

	chat.UpdatedAt = time.Now()
	r.chats[chat.ID] = chat
	return nil
}

func (r *implChatRepository) DeleteChat(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.chats, id)
	delete(r.chatMembers, id)
	return nil
}

// Message operations
func (r *implChatRepository) CreateMessage(message *entity.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.messageID++
	message.ID = r.messageID
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()
	if message.Reactions == nil {
		message.Reactions = []*entity.Reaction{}
	}

	if message.ReplyToID != nil {
		if reply, ok := r.messages[*message.ReplyToID]; ok {
			message.ReplyTo = reply
		}
	}

	r.messages[message.ID] = message
	return nil
}

func (r *implChatRepository) GetMessageByID(id int64) (*entity.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	message, ok := r.messages[id]
	if !ok {
		return nil, fmt.Errorf("message not found")
	}

	return message, nil
}

func (r *implChatRepository) GetMessagesByChat(chatID int64, limit, offset int) ([]*entity.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var messages []*entity.Message
	for _, msg := range r.messages {
		if msg.ChatID == chatID {
			messages = append(messages, msg)
		}
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt.Before(messages[j].CreatedAt)
	})

	// Simple pagination
	start := offset
	end := offset + limit
	if start > len(messages) {
		return []*entity.Message{}, nil
	}
	if end > len(messages) {
		end = len(messages)
	}

	// Attach reactions and reply references for each message before returning slice copy
	result := make([]*entity.Message, 0, end-start)
	for _, msg := range messages[start:end] {
		reactions := make([]*entity.Reaction, 0)
		for _, reaction := range r.reactions {
			if reaction.MessageID == msg.ID {
				reactions = append(reactions, reaction)
			}
		}
		msg.Reactions = reactions

		if msg.ReplyToID != nil {
			if reply, ok := r.messages[*msg.ReplyToID]; ok {
				msg.ReplyTo = reply
			}
		}

		result = append(result, msg)
	}

	return result, nil
}

func (r *implChatRepository) UpdateMessage(message *entity.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.messages[message.ID]; !ok {
		return fmt.Errorf("message not found")
	}

	message.UpdatedAt = time.Now()
	r.messages[message.ID] = message
	return nil
}

func (r *implChatRepository) DeleteMessage(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.messages, id)
	return nil
}

// Reaction operations
func (r *implChatRepository) CreateReaction(reaction *entity.Reaction, userID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if reaction record already exists for this message and type
	for _, existing := range r.reactions {
		if existing.MessageID == reaction.MessageID && existing.Type == reaction.Type {
			// Check if user already reacted
			for _, uid := range existing.UserIDs {
				if uid == userID {
					return nil // Already reacted
				}
			}
			// Add user to existing reaction
			existing.UserIDs = append(existing.UserIDs, userID)
			existing.Count = len(existing.UserIDs)
			existing.UpdatedAt = time.Now()
			return nil
		}
	}

	// Create new reaction
	r.reactionID++
	reaction.ID = r.reactionID
	reaction.UserIDs = []int64{userID}
	reaction.Count = 1
	reaction.CreatedAt = time.Now()
	reaction.UpdatedAt = time.Now()

	r.reactions[reaction.ID] = reaction

	if message, ok := r.messages[reaction.MessageID]; ok {
		message.Reactions = append(message.Reactions, reaction)
	}
	return nil
}

func (r *implChatRepository) GetReactionsByMessage(messageID int64) ([]*entity.Reaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var reactions []*entity.Reaction
	for _, reaction := range r.reactions {
		if reaction.MessageID == messageID {
			reactions = append(reactions, reaction)
		}
	}

	return reactions, nil
}

func (r *implChatRepository) DeleteReaction(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	reaction, ok := r.reactions[id]
	if ok {
		if message, exists := r.messages[reaction.MessageID]; exists {
			filtered := make([]*entity.Reaction, 0, len(message.Reactions))
			for _, rct := range message.Reactions {
				if rct.ID != id {
					filtered = append(filtered, rct)
				}
			}
			message.Reactions = filtered
		}
	}

	delete(r.reactions, id)
	return nil
}

// Chat member operations
func (r *implChatRepository) AddChatMember(member *entity.ChatMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.chatMembers[member.ChatID] {
		if existing.UserID == member.UserID {
			return nil
		}
	}

	r.chatMemberID++
	member.ID = r.chatMemberID
	member.JoinedAt = time.Now()

	r.chatMembers[member.ChatID] = append(r.chatMembers[member.ChatID], member)
	return nil
}

func (r *implChatRepository) GetChatMembers(chatID int64) ([]*entity.ChatMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	members, ok := r.chatMembers[chatID]
	if !ok {
		return []*entity.ChatMember{}, nil
	}

	return members, nil
}

func (r *implChatRepository) RemoveChatMember(chatID, userID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	members, ok := r.chatMembers[chatID]
	if !ok {
		return fmt.Errorf("chat not found")
	}

	for i, member := range members {
		if member.UserID == userID {
			r.chatMembers[chatID] = append(members[:i], members[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("member not found")
}

func (r *implChatRepository) IsChatMember(chatID, userID int64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	members, ok := r.chatMembers[chatID]
	if !ok {
		return false, nil
	}

	for _, member := range members {
		if member.UserID == userID {
			return true, nil
		}
	}

	return false, nil
}
