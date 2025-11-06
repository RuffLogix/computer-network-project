//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/rufflogix/computer-network-project/internal/controller"
	"github.com/rufflogix/computer-network-project/internal/repository"
	"github.com/rufflogix/computer-network-project/internal/service"
	"go.mongodb.org/mongo-driver/mongo"
)

type ServerHandlers struct {
	HTTP controller.HTTPHandler
}

func provideServerHandlers(
	httpHandler controller.HTTPHandler,
) ServerHandlers {
	return ServerHandlers{
		HTTP: httpHandler,
	}
}

func InitializeHandlers(db *mongo.Database) ServerHandlers {
	wire.Build(
		repository.NewMongoChatRepository,
		repository.NewInvitationRepository,
		repository.NewMongoFriendshipRepository,
		repository.NewMongoNotificationRepository,
		repository.NewUserRepository,
		service.NewRoomService,
		service.NewChatService,
		service.NewNotificationService,
		service.NewInvitationService,
		service.NewAuthService,
		controller.NewHTTPHandler,
		provideServerHandlers,
	)

	return ServerHandlers{}
}
