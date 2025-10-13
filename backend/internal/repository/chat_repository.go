package repository

type ChatRepository interface {
}

type implChatRepository struct {
}

func NewChatRepository() ChatRepository {
	return &implChatRepository{}
}
