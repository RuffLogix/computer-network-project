package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rufflogix/computer-network-project/internal/entity"
	"github.com/rufflogix/computer-network-project/internal/service"
)

type WSHandler interface {
	HandleWS(http.ResponseWriter, *http.Request)
}

type implWSHandler struct {
	chatService         service.ChatService
	roomService         service.RoomService
	notificationService service.NotificationService
	invitationService   service.InvitationService
}

func NewWSHandler(
	chatService service.ChatService,
	roomService service.RoomService,
	notificationService service.NotificationService,
	invitationService service.InvitationService,
) WSHandler {
	return &implWSHandler{
		chatService:         chatService,
		roomService:         roomService,
		notificationService: notificationService,
		invitationService:   invitationService,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (h *implWSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	var userID int64
	log.Printf("WebSocket connection established from %s", r.RemoteAddr)

	// Set read deadline to prevent hanging connections
	conn.SetReadDeadline(time.Time{})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var event entity.Event
		if err := json.Unmarshal(message, &event); err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			continue
		}

		// Set userID from the first event and register client
		if userID == 0 && event.CreatedBy != 0 {
			userID = event.CreatedBy
			h.roomService.AddClient(conn, userID)
			log.Printf("User %d connected via WebSocket", userID)

			// Send initial online status of all friends to the newly connected user
			h.sendInitialFriendsOnlineStatus(userID)

			// Broadcast online status to friends
			h.broadcastUserStatus(userID, true)

			// Send the complete online users list to the newly connected user
			h.sendOnlineUsersList(userID)
		}

		// Update userID if changed
		if event.CreatedBy != 0 {
			userID = event.CreatedBy
		}

		switch event.Type {
		case "connect":
			// Connection event - client is now registered
			log.Printf("User %d sent connect event - connection acknowledged", userID)
			continue

		case entity.JOIN:
			h.handleJoin(conn, &event)

		case entity.LEAVE:
			h.handleLeave(&event)

		case entity.SEND_MESSAGE:
			h.handleSendMessage(&event)

		case entity.EDIT_MESSAGE:
			h.handleEditMessage(&event)

		case entity.DELETE_MESSAGE:
			h.handleDeleteMessage(&event)

		case entity.ADD_REACTION:
			h.handleAddReaction(&event)

		case entity.REMOVE_REACTION:
			h.handleRemoveReaction(&event)

		case entity.TYPING:
			h.handleTyping(&event)

		case entity.NOTIFICATION:
			h.handleNotification(&event)

		default:
			log.Printf("Unknown event type: %s", event.Type)
		}
	}

	// Clean up on disconnect
	if userID != 0 {
		// Broadcast offline status to friends
		h.broadcastUserStatus(userID, false)
		h.roomService.RemoveClient(userID)
	}
}

func (h *implWSHandler) handleJoin(conn *websocket.Conn, event *entity.Event) {
	chatID, ok := event.Data["chat_id"].(float64)
	if !ok {
		log.Printf("Invalid chat_id in join event")
		return
	}

	h.roomService.AddClient(conn, event.CreatedBy)
	isNewJoin := h.roomService.JoinRoom(event.CreatedBy, int64(chatID))

	if isNewJoin {
		log.Printf("User %d joined chat room %d", event.CreatedBy, int64(chatID))
	}
}

func (h *implWSHandler) handleLeave(event *entity.Event) {
	chatID, ok := event.Data["chat_id"].(float64)
	if !ok {
		return
	}

	h.roomService.LeaveRoom(event.CreatedBy, int64(chatID))
}

func (h *implWSHandler) handleSendMessage(event *entity.Event) {
	chatID, ok := event.Data["chat_id"].(float64)
	if !ok {
		return
	}

	content, _ := event.Data["content"].(string)
	msgType, _ := event.Data["type"].(string)
	mediaURL, _ := event.Data["media_url"].(string)

	var replyToID *int64
	if replyTo, ok := event.Data["reply_to_id"].(float64); ok {
		replyID := int64(replyTo)
		replyToID = &replyID
	}

	message := &entity.Message{
		ChatID:    int64(chatID),
		Content:   content,
		Type:      entity.MessageType(msgType),
		MediaURL:  mediaURL,
		ReplyToID: replyToID,
		CreatedBy: event.CreatedBy,
	}

	if err := h.chatService.SendMessage(message); err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	// Broadcast message to all chat members
	h.broadcastEvent(int64(chatID), entity.Event{
		Type:      entity.SEND_MESSAGE,
		Data:      map[string]interface{}{"message": message},
		CreatedBy: event.CreatedBy,
	}, 0)
}

func (h *implWSHandler) handleEditMessage(event *entity.Event) {
	messageID, ok := event.Data["message_id"].(float64)
	if !ok {
		return
	}

	content, _ := event.Data["content"].(string)

	if err := h.chatService.EditMessage(int64(messageID), content); err != nil {
		log.Printf("Error editing message: %v", err)
		return
	}

	// Get message to find chat ID
	message, err := h.chatService.GetMessages(0, 1, 0) // This is simplified
	if err != nil || len(message) == 0 {
		return
	}

	// Broadcast edit event
	h.broadcastEvent(message[0].ChatID, entity.Event{
		Type: entity.EDIT_MESSAGE,
		Data: map[string]interface{}{
			"message_id": messageID,
			"content":    content,
		},
		CreatedBy: event.CreatedBy,
	}, 0)
}

func (h *implWSHandler) handleDeleteMessage(event *entity.Event) {
	messageID, ok := event.Data["message_id"].(float64)
	if !ok {
		return
	}

	chatID, _ := event.Data["chat_id"].(float64)

	if err := h.chatService.DeleteMessage(int64(messageID)); err != nil {
		log.Printf("Error deleting message: %v", err)
		return
	}

	// Broadcast delete event
	h.broadcastEvent(int64(chatID), entity.Event{
		Type:      entity.DELETE_MESSAGE,
		Data:      map[string]interface{}{"message_id": messageID},
		CreatedBy: event.CreatedBy,
	}, 0)
}

func (h *implWSHandler) handleAddReaction(event *entity.Event) {
	messageID, ok := event.Data["message_id"].(float64)
	if !ok {
		return
	}

	reactionType, _ := event.Data["type"].(string)
	chatID, _ := event.Data["chat_id"].(float64)

	// Create reaction object for the service call
	reaction := &entity.Reaction{
		MessageID: int64(messageID),
		Type:      entity.ReactionType(reactionType),
	}

	// Add/toggle reaction (service handles the toggle logic)
	if err := h.chatService.AddReaction(reaction, event.CreatedBy); err != nil {
		log.Printf("Error adding reaction: %v", err)
		return
	}

	// Get updated reactions for the message
	updatedReactions, err := h.chatService.GetMessageReactions(int64(messageID))
	if err != nil {
		log.Printf("Error getting updated reactions: %v", err)
		return
	}

	// Find the reaction that was just updated
	var updatedReaction *entity.Reaction
	for _, r := range updatedReactions {
		if r.Type == reaction.Type {
			updatedReaction = r
			break
		}
	}

	if updatedReaction == nil {
		log.Printf("Could not find updated reaction")
		return
	}

	// Broadcast the updated reaction
	h.broadcastEvent(int64(chatID), entity.Event{
		Type:      entity.ADD_REACTION,
		Data:      map[string]interface{}{"reaction": updatedReaction},
		CreatedBy: event.CreatedBy,
	}, 0)
}

func (h *implWSHandler) handleRemoveReaction(event *entity.Event) {
	reactionID, ok := event.Data["reaction_id"].(float64)
	if !ok {
		return
	}

	chatID, _ := event.Data["chat_id"].(float64)

	if err := h.chatService.RemoveReaction(int64(reactionID), event.CreatedBy); err != nil {
		log.Printf("Error removing reaction: %v", err)
		return
	}

	// Broadcast reaction removal
	h.broadcastEvent(int64(chatID), entity.Event{
		Type:      entity.REMOVE_REACTION,
		Data:      map[string]interface{}{"reaction_id": reactionID},
		CreatedBy: event.CreatedBy,
	}, 0)
}

func (h *implWSHandler) handleTyping(event *entity.Event) {
	chatID, ok := event.Data["chat_id"].(float64)
	if !ok {
		return
	}

	isTyping, _ := event.Data["is_typing"].(bool)

	// Broadcast typing indicator (exclude sender)
	h.broadcastEvent(int64(chatID), entity.Event{
		Type: entity.TYPING,
		Data: map[string]interface{}{
			"user_id":   event.CreatedBy,
			"is_typing": isTyping,
		},
		CreatedBy: event.CreatedBy,
	}, event.CreatedBy)
}

func (h *implWSHandler) handleNotification(event *entity.Event) {
	recipientID, ok := event.Data["recipient_id"].(float64)
	if !ok {
		return
	}

	// Send notification directly to recipient
	h.roomService.SendToUser(int64(recipientID), *event)
}

// Send initial online status of all friends to a newly connected user
func (h *implWSHandler) sendInitialFriendsOnlineStatus(userID int64) {
	// Get user's friends
	friendships, err := h.invitationService.GetFriendships(userID)
	if err != nil {
		log.Printf("Error getting friendships for user %d: %v", userID, err)
		return
	}

	// Get all online users
	onlineUserIDs := h.roomService.GetOnlineUsers()
	onlineUsersMap := make(map[int64]bool)
	for _, id := range onlineUserIDs {
		onlineUsersMap[id] = true
	}

	// Send online status for each friend who is currently online
	for _, friendship := range friendships {
		var friendID int64
		if friendship.UserID == userID {
			friendID = friendship.FriendID
		} else {
			friendID = friendship.UserID
		}

		// Only send if friend is online
		if onlineUsersMap[friendID] {
			statusEvent := entity.Event{
				Type:      "user_online",
				CreatedBy: friendID,
				Data: map[string]interface{}{
					"user_id": friendID,
				},
			}
			h.roomService.SendToUser(userID, statusEvent)
		}
	}
}

// Broadcast online/offline status to friends
func (h *implWSHandler) broadcastUserStatus(userID int64, isOnline bool) {
	// Get user's friends
	friendships, err := h.invitationService.GetFriendships(userID)
	if err != nil {
		log.Printf("Error getting friendships for user %d: %v", userID, err)
		return
	}

	// Prepare status event
	eventType := "user_online"
	if !isOnline {
		eventType = "user_offline"
	}

	statusEvent := entity.Event{
		Type:      entity.EventType(eventType),
		CreatedBy: userID,
		Data: map[string]interface{}{
			"user_id": userID,
		},
	}

	// Send status to each friend
	for _, friendship := range friendships {
		var friendID int64
		if friendship.UserID == userID {
			friendID = friendship.FriendID
		} else {
			friendID = friendship.UserID
		}
		h.roomService.SendToUser(friendID, statusEvent)
	}

	// Broadcast the updated online users list to all connected clients
	h.broadcastOnlineUsersList()
}

// Broadcast the complete list of online users to all connected clients
func (h *implWSHandler) broadcastOnlineUsersList() {
	onlineUserIDs := h.roomService.GetOnlineUsers()

	// Send personalized online user lists to each connected user
	for _, userID := range onlineUserIDs {
		h.sendOnlineUsersList(userID)
	}
}

// Send the complete list of online users to a specific user (filtered by shared chats)
func (h *implWSHandler) sendOnlineUsersList(userID int64) {
	onlineUserIDs := h.roomService.GetOnlineUsers()

	// Get user's chats to find all members they can see
	userChats, err := h.chatService.GetUserChats(userID)
	if err != nil {
		log.Printf("Error getting user chats for online list: %v", err)
		return
	}

	// Build a set of user IDs that share chats with the current user
	visibleUserIDs := make(map[int64]bool)
	for _, chat := range userChats {
		members, err := h.chatService.GetMembers(chat.ID)
		if err != nil {
			continue
		}
		for _, member := range members {
			if member.UserID != userID {
				visibleUserIDs[member.UserID] = true
			}
		}
	}

	// Filter online users to only those visible to this user
	var filteredOnlineUsers []int64
	for _, id := range onlineUserIDs {
		if visibleUserIDs[id] || id == userID {
			filteredOnlineUsers = append(filteredOnlineUsers, id)
		}
	}

	onlineUsersEvent := entity.Event{
		Type: "online_users_list",
		Data: map[string]interface{}{
			"online_users": filteredOnlineUsers,
		},
	}

	h.roomService.SendToUser(userID, onlineUsersEvent)
}

func (h *implWSHandler) broadcastEvent(chatID int64, event entity.Event, excludeUserID int64) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	if excludeUserID == 0 {
		h.roomService.BroadcastToRoom(chatID, data)
	} else {
		h.roomService.BroadcastToRoomExcept(chatID, data, excludeUserID)
	}
}
