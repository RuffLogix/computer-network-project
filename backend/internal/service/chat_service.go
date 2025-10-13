package service

import "github.com/rufflogix/computer-network-project/internal/repository"

type ChatService interface {
}

type implChatService struct {
	chatRepository repository.ChatRepository
}

func NewChatService(chatRepository repository.ChatRepository) ChatService {
	return &implChatService{chatRepository: chatRepository}
}
