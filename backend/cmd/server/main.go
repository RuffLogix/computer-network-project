package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rufflogix/computer-network-project/internal/config"
	"github.com/rufflogix/computer-network-project/internal/controller"
	"github.com/rufflogix/computer-network-project/internal/entity"
	"github.com/rufflogix/computer-network-project/internal/middleware"
	"github.com/rufflogix/computer-network-project/internal/repository"
	"github.com/rufflogix/computer-network-project/internal/service"
)

func initializeGlobalChat(chatService service.ChatService) {
	// Check if "Global Chat" already exists in public chats
	chats, err := chatService.GetPublicChats()
	if err != nil {
		log.Printf("Warning: Could not check for existing global chat: %v", err)
		return
	}

	for _, chat := range chats {
		if chat.Name == "Global Chat" && chat.Type == "public_group" {
			log.Println("Global Chat already exists")
			return
		}
	}

	// Create Global Chat
	globalChat := &entity.Chat{
		Name:        "Global Chat",
		Description: "A public chat room for everyone",
		Type:        entity.PublicGroup,
		IsPublic:    true,
		CreatedBy:   0, // System created
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = chatService.CreateChat(globalChat)
	if err != nil {
		log.Printf("Warning: Could not create global chat: %v", err)
		return
	}

	log.Println("Global Chat created successfully")
}

func main() {
	// Connect to MongoDB
	db := config.ConnectDB()

	// Create database indexes
	if err := repository.CreateChatIndexes(db); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewMongoChatRepository(db)
	invitationRepo := repository.NewInvitationRepository()
	friendshipRepo := repository.NewFriendshipRepository()
	notificationRepo := repository.NewNotificationRepository()

	// Initialize services
	roomService := service.NewRoomService()
	chatService := service.NewChatService(chatRepo, userRepo)
	notificationService := service.NewNotificationService(notificationRepo, friendshipRepo, chatRepo, userRepo, roomService)
	invitationService := service.NewInvitationService(invitationRepo, chatRepo, friendshipRepo, notificationService, userRepo)
	authService := service.NewAuthService(userRepo)

	// Initialize global chat if it doesn't exist
	initializeGlobalChat(chatService)

	// Initialize handlers
	httpHandler := controller.NewHTTPHandler(chatService, invitationService, notificationService, authService, roomService, userRepo)
	wsHandler := controller.NewWSHandler(chatService, roomService, notificationService, invitationService)
	authHandler := controller.NewAuthHandler(authService)

	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-User-ID", "x-user-id"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	os.MkdirAll("./uploads", 0o755)
	r.Static("/uploads", "./uploads")

	api := r.Group("/api")
	{
		authHandler.RegisterRoutes(api)
	}

	httpHandler.RegisterRoutes(r)

	r.GET("/ws", middleware.OptionalAuthMiddleware(authService), func(c *gin.Context) {
		wsHandler.HandleWS(c.Writer, c.Request)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...\n", port)
	r.Run(":" + port)
}
