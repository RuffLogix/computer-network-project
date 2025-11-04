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
	WS   controller.WSHandler
	Auth controller.AuthHandler
}

func provideServerHandlers(
	httpHandler controller.HTTPHandler,
	wsHandler controller.WSHandler,
	authHandler controller.AuthHandler,
) ServerHandlers {
	return ServerHandlers{
		HTTP: httpHandler,
		WS:   wsHandler,
		Auth: authHandler,
	}
}

func InitializeHandlers(db *mongo.Database) ServerHandlers {
	wire.Build(
		repository.NewChatRepository,
		repository.NewInvitationRepository,
		repository.NewFriendshipRepository,
		repository.NewNotificationRepository,
		repository.NewUserRepository,
		service.NewRoomService,
		service.NewChatService,
		service.NewNotificationService,
		service.NewInvitationService,
		service.NewAuthService,
		controller.NewHTTPHandler,
		controller.NewWSHandler,
		controller.NewAuthHandler,
		provideServerHandlers,
	)

	return ServerHandlers{}
}
