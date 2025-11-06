package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rufflogix/computer-network-project/internal/service"
)

type AuthHandler interface {
	RegisterRoutes(*gin.RouterGroup)
}

type implAuthHandler struct {
	authService  service.AuthService
	chatService  service.ChatService
	globalChatID int64
}

func NewAuthHandler(authService service.AuthService, chatService service.ChatService, globalChatID int64) AuthHandler {
	return &implAuthHandler{
		authService:  authService,
		chatService:  chatService,
		globalChatID: globalChatID,
	}
}

func (h *implAuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.register)
		auth.POST("/login", h.login)
		auth.POST("/guest", h.createGuest)
	}
}

func (h *implAuthHandler) register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=50"`
		Password string `json:"password" binding:"required,min=6"`
		Name     string `json:"name" binding:"required,min=1,max=100"`
		Email    string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.authService.Register(req.Username, req.Password, req.Name, req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add user to global chat
	if h.globalChatID != 0 {
		if err := h.chatService.AddMember(h.globalChatID, user.NumericID, "member"); err != nil {
			// Log error but don't fail registration
			// TODO: Add proper logging
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": token,
	})
}

func (h *implAuthHandler) login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Add user to global chat if not already a member
	if h.globalChatID != 0 {
		if err := h.chatService.AddMember(h.globalChatID, user.NumericID, "member"); err != nil {
			// Log error but don't fail login
			// TODO: Add proper logging
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

func (h *implAuthHandler) createGuest(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required,min=1,max=100"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.authService.CreateGuestUser(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add guest user to global chat
	if h.globalChatID != 0 {
		if err := h.chatService.AddMember(h.globalChatID, user.NumericID, "member"); err != nil {
			// Log error but don't fail guest creation
			// TODO: Add proper logging
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": token,
	})
}
