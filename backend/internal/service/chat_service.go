package service

import (
	"sync"

	"github.com/rufflogix/computer-network-project/internal/entity"
	"github.com/rufflogix/computer-network-project/internal/repository"
)

type ChatService interface {
}

type implChatService struct {
	chatRepository repository.ChatRepository
	clients        map[string]*entity.Client
	mu             sync.RWMutex
}

func NewChatService(chatRepository repository.ChatRepository) ChatService {
	return &implChatService{chatRepository: chatRepository}
}

func (chatService *implChatService) AddClient(c *entity.Client) {
	chatService.mu.Lock()
	defer chatService.mu.Unlock()
	chatService.clients[c.ID] = c
}

func (chatService *implChatService) RemoveClient(id string) {
	chatService.mu.Lock()
	defer chatService.mu.Unlock()
	delete(chatService.clients, id)
}
