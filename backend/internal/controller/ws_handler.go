package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rufflogix/computer-network-project/internal/entity"
	"github.com/rufflogix/computer-network-project/internal/service"
)

type WSHandler interface {
	HandleWS(http.ResponseWriter, *http.Request)
}

type implWSHandler struct {
	chatService service.ChatService
	roomService service.RoomService
}

func NewWSHandler(chatService service.ChatService, roomService service.RoomService) WSHandler {
	return &implWSHandler{chatService: chatService, roomService: roomService}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *implWSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var messageJSON entity.Message
		if err := json.Unmarshal(message, &messageJSON); err != nil {
			break
		}

		h.roomService.AddClient(conn, messageJSON.CreatedBy)
		h.roomService.Broadcast([]byte(messageJSON.Content), messageJSON.CreatedBy)
	}
}
