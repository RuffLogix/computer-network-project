//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/rufflogix/computer-network-project/internal/controller"
	"github.com/rufflogix/computer-network-project/internal/repository"
	"github.com/rufflogix/computer-network-project/internal/service"
)

func InitializeWSHandler() controller.WSHandler {
	wire.Build(
		controller.NewWSHandler,
		service.NewChatService,
		service.NewRoomService,
		repository.NewChatRepository,
	)

	return nil
}

func InitializeSSEHandler() controller.SSEHandler {
	wire.Build(
		controller.NewSSEHandler,
	)

	return nil
}

func InitializeHTTPHandler() controller.HTTPHandler {
	wire.Build(
		controller.NewHTTPHander,
	)

	return nil
}
