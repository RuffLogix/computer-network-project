package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rufflogix/computer-network-project/internal/entity"
	"github.com/rufflogix/computer-network-project/internal/middleware"
	"github.com/rufflogix/computer-network-project/internal/service"
)

type HTTPHandler interface {
	RegisterRoutes(*gin.Engine)
}

type implHTTPHandler struct {
	chatService         service.ChatService
	invitationService   service.InvitationService
	notificationService service.NotificationService
	authService         service.AuthService
	roomService         service.RoomService
}

func NewHTTPHandler(
	chatService service.ChatService,
	invitationService service.InvitationService,
	notificationService service.NotificationService,
	authService service.AuthService,
	roomService service.RoomService,
) HTTPHandler {
	return &implHTTPHandler{
		chatService:         chatService,
		invitationService:   invitationService,
		notificationService: notificationService,
		authService:         authService,
		roomService:         roomService,
	}
}

func (h *implHTTPHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")

	// Public routes (no auth required)
	{
		// Public chats can be viewed by anyone
		api.GET("/chats/public", h.getPublicChats)
	}

	// Protected routes (require authentication)
	authorized := api.Group("")
	authorized.Use(middleware.AuthMiddleware(h.authService))
	{
		// Chat routes
		chats := authorized.Group("/chats")
		{
			chats.POST("", h.createChat)
			chats.GET("/:id", h.getChat)
			chats.GET("", h.getUserChats)
			chats.POST("/:id/messages", h.sendMessage)
			chats.GET("/:id/messages", h.getMessages)
			chats.POST("/:id/members", h.addMember)
			chats.DELETE("/:id/members/:userId", h.removeMember)
			chats.POST("/:id/join", h.joinPublicChat)
		}

		// Message routes
		messages := authorized.Group("/messages")
		{
			messages.PUT("/:id", h.editMessage)
			messages.DELETE("/:id", h.deleteMessage)
			messages.POST("/:id/reactions", h.addReaction)
			messages.GET("/:id/reactions", h.getReactions)
			messages.DELETE("/reactions/:id", h.removeReaction)
		}

		// Invitation routes
		invitations := authorized.Group("/invitations")
		{
			invitations.POST("/chat", h.createChatInvitation)
			invitations.POST("/friend", h.createFriendInvitation)
			invitations.POST("/friend/request", h.sendFriendRequest)
			invitations.POST("/chat/:code/join", h.joinChatViaInvitation)
			invitations.POST("/friend/:code/accept", h.acceptFriendInvitation)
			invitations.GET("/chat/:id", h.listChatInvitations)
			invitations.GET("/friend", h.listFriendInvitations)
		}

		// Notification routes
		notifications := authorized.Group("/notifications")
		{
			notifications.GET("", h.getUserNotifications)
			notifications.GET("/unread", h.getUnreadNotifications)
			notifications.PUT("/:id/read", h.markNotificationAsRead)
			notifications.POST("/:id/accept", h.acceptNotification)
			notifications.POST("/:id/reject", h.rejectNotification)
		}

		// Friends routes
		authorized.GET("/friends", h.getFriends)

		// Media upload route
		authorized.POST("/upload", h.uploadMedia)
	}
}

// Chat handlers
func (h *implHTTPHandler) createChat(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		IsPublic    bool   `json:"is_public"`
		Type        string `json:"type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (assumes auth middleware sets this)
	userID := c.GetInt64("user_id")

	chat := &entity.Chat{
		Type:        entity.ChatType(req.Type),
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    req.IsPublic,
		CreatedBy:   userID,
	}

	if err := h.chatService.CreateChat(chat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add creator as admin member
	h.chatService.AddMember(chat.ID, userID, "admin")

	c.JSON(http.StatusCreated, chat)
}

func (h *implHTTPHandler) getChat(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	chat, err := h.chatService.GetChat(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chat)
}

func (h *implHTTPHandler) getUserChats(c *gin.Context) {
	userID := c.GetInt64("user_id")

	chats, err := h.chatService.GetUserChats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chats)
}

func (h *implHTTPHandler) getPublicChats(c *gin.Context) {
	chats, err := h.chatService.GetPublicChats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chats)
}

func (h *implHTTPHandler) sendMessage(c *gin.Context) {
	chatID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Content   string `json:"content"`
		Type      string `json:"type" binding:"required"`
		MediaURL  string `json:"media_url"`
		ReplyToID *int64 `json:"reply_to_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt64("user_id")

	message := &entity.Message{
		ChatID:    chatID,
		Content:   req.Content,
		Type:      entity.MessageType(req.Type),
		MediaURL:  req.MediaURL,
		ReplyToID: req.ReplyToID,
		CreatedBy: userID,
	}

	if err := h.chatService.SendMessage(message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, message)
}

func (h *implHTTPHandler) getMessages(c *gin.Context) {
	chatID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	messages, err := h.chatService.GetMessages(chatID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h *implHTTPHandler) editMessage(c *gin.Context) {
	messageID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.chatService.EditMessage(messageID, req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message updated"})
}

func (h *implHTTPHandler) deleteMessage(c *gin.Context) {
	messageID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.chatService.DeleteMessage(messageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message deleted"})
}

func (h *implHTTPHandler) addMember(c *gin.Context) {
	chatID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		UserID int64  `json:"user_id" binding:"required"`
		Role   string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := req.Role
	if role == "" {
		role = "member"
	}

	if err := h.chatService.AddMember(chatID, req.UserID, role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member added"})
}

func (h *implHTTPHandler) removeMember(c *gin.Context) {
	chatID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := strconv.ParseInt(c.Param("userId"), 10, 64)

	// Create a system message for user leaving before they leave
	systemMessage := &entity.Message{
		ChatID:    chatID,
		Content:   "left the chat",
		Type:      entity.System,
		CreatedBy: userID,
		CreatedAt: time.Now(),
	}

	if err := h.chatService.SendMessage(systemMessage); err != nil {
		log.Printf("Error creating leave system message: %v", err)
	}

	// Broadcast the system message to all chat members
	h.broadcastEvent(chatID, entity.Event{
		Type:      entity.SEND_MESSAGE,
		Data:      map[string]interface{}{"message": systemMessage},
		CreatedBy: userID,
	}, 0)

	if err := h.chatService.RemoveMember(chatID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed"})
}

func (h *implHTTPHandler) joinPublicChat(c *gin.Context) {
	chatID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := c.GetInt64("user_id")

	chat, err := h.chatService.GetChat(chatID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if !chat.IsPublic {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Chat is not public"})
		return
	}

	if err := h.chatService.AddMember(chatID, userID, "member"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create a system message for user joining
	systemMessage := &entity.Message{
		ChatID:    chatID,
		Content:   "joined the chat",
		Type:      entity.System,
		CreatedBy: userID,
		CreatedAt: time.Now(),
	}

	if err := h.chatService.SendMessage(systemMessage); err != nil {
		log.Printf("Error creating join system message: %v", err)
	}

	// Broadcast the system message to all chat members
	h.broadcastEvent(chatID, entity.Event{
		Type:      entity.SEND_MESSAGE,
		Data:      map[string]interface{}{"message": systemMessage},
		CreatedBy: userID,
	}, 0)

	c.JSON(http.StatusOK, gin.H{"message": "Joined chat"})
}

// Reaction handlers
func (h *implHTTPHandler) addReaction(c *gin.Context) {
	messageID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Type string `json:"type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt64("user_id")

	reaction := &entity.Reaction{
		MessageID: messageID,
		Type:      entity.ReactionType(req.Type),
	}

	if err := h.chatService.AddReaction(reaction, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, reaction)
}

func (h *implHTTPHandler) getReactions(c *gin.Context) {
	messageID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	reactions, err := h.chatService.GetMessageReactions(messageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reactions)
}

func (h *implHTTPHandler) removeReaction(c *gin.Context) {
	reactionID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := c.GetInt64("user_id")

	if err := h.chatService.RemoveReaction(reactionID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reaction removed"})
}

// Invitation handlers
func (h *implHTTPHandler) createChatInvitation(c *gin.Context) {
	var req struct {
		ChatID    int64  `json:"chat_id" binding:"required"`
		ExpiresIn *int64 `json:"expires_in"` // seconds
		MaxUses   *int   `json:"max_uses"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt64("user_id")

	chat, err := h.chatService.GetChat(req.ChatID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if chat.IsPublic {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Public chats do not require invitations"})
		return
	}

	var expiresIn *time.Duration
	if req.ExpiresIn != nil {
		duration := time.Duration(*req.ExpiresIn) * time.Second
		expiresIn = &duration
	}

	invitation, err := h.invitationService.CreateChatInvitation(req.ChatID, userID, expiresIn, req.MaxUses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, invitation)
}

func (h *implHTTPHandler) createFriendInvitation(c *gin.Context) {
	var req struct {
		ExpiresIn *int64 `json:"expires_in"` // seconds
		MaxUses   *int   `json:"max_uses"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt64("user_id")

	var expiresIn *time.Duration
	if req.ExpiresIn != nil {
		duration := time.Duration(*req.ExpiresIn) * time.Second
		expiresIn = &duration
	}

	invitation, err := h.invitationService.CreateFriendInvitation(userID, expiresIn, req.MaxUses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, invitation)
}

func (h *implHTTPHandler) joinChatViaInvitation(c *gin.Context) {
	code := c.Param("code")
	userID := c.GetInt64("user_id")

	if err := h.invitationService.JoinChatViaInvitation(code, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully joined chat"})
}

func (h *implHTTPHandler) acceptFriendInvitation(c *gin.Context) {
	code := c.Param("code")
	userID := c.GetInt64("user_id")

	if err := h.invitationService.AcceptFriendInvitation(code, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend request sent"})
}

func (h *implHTTPHandler) sendFriendRequest(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		TargetIdentifier string `json:"target_identifier" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.invitationService.SendFriendRequest(userID, req.TargetIdentifier); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend request sent"})
}

func (h *implHTTPHandler) listChatInvitations(c *gin.Context) {
	chatID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	invitations, err := h.invitationService.GetChatInvitations(chatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invitations)
}

func (h *implHTTPHandler) listFriendInvitations(c *gin.Context) {
	userID := c.GetInt64("user_id")

	invitations, err := h.invitationService.GetFriendInvitations(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invitations)
}

// Notification handlers
func (h *implHTTPHandler) getUserNotifications(c *gin.Context) {
	userID := c.GetInt64("user_id")

	notifications, err := h.notificationService.GetUserNotifications(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

func (h *implHTTPHandler) getUnreadNotifications(c *gin.Context) {
	userID := c.GetInt64("user_id")

	notifications, err := h.notificationService.GetUnreadNotifications(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

func (h *implHTTPHandler) markNotificationAsRead(c *gin.Context) {
	notificationID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.notificationService.MarkAsRead(notificationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

func (h *implHTTPHandler) acceptNotification(c *gin.Context) {
	notificationID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := c.GetInt64("user_id")

	if err := h.notificationService.AcceptNotification(notificationID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification accepted"})
}

func (h *implHTTPHandler) rejectNotification(c *gin.Context) {
	notificationID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := c.GetInt64("user_id")

	if err := h.notificationService.RejectNotification(notificationID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification rejected"})
}

// Get friends list
func (h *implHTTPHandler) getFriends(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	friendships, err := h.invitationService.GetFriendships(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get online users from room service
	onlineUserIDs := h.roomService.GetOnlineUsers()

	// Build friend list with online status
	type FriendResponse struct {
		entity.User
		IsOnline bool `json:"is_online"`
	}

	friends := make([]FriendResponse, 0)
	for _, friendship := range friendships {
		var friend *entity.User
		if friendship.UserID == userID.(int64) {
			friend = friendship.Friend
		} else {
			friend = friendship.User
		}

		if friend != nil {
			isOnline := false
			for _, onlineID := range onlineUserIDs {
				if onlineID == friend.NumericID {
					isOnline = true
					break
				}
			}

			friends = append(friends, FriendResponse{
				User:     *friend,
				IsOnline: isOnline,
			})
		}
	}

	c.JSON(http.StatusOK, friends)
}

// Media upload handler
func (h *implHTTPHandler) uploadMedia(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Save file to uploads directory
	filename := strconv.FormatInt(time.Now().Unix(), 10) + "_" + file.Filename
	filepath := "./uploads/" + filename

	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Return the file URL
	c.JSON(http.StatusOK, gin.H{
		"url":      "/uploads/" + filename,
		"filename": file.Filename,
	})
}

// Helper method to broadcast events to chat members
func (h *implHTTPHandler) broadcastEvent(chatID int64, event entity.Event, excludeUserID int64) {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	if excludeUserID > 0 {
		h.roomService.BroadcastToRoomExcept(chatID, eventJSON, excludeUserID)
	} else {
		h.roomService.BroadcastToRoom(chatID, eventJSON)
	}
}
