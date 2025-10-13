package controller

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rufflogix/computer-network-project/internal/service"
)

type WSHandler interface {
	HandleWS(http.ResponseWriter, *http.Request)
}

type implWSHandler struct {
	chatService service.ChatService
}

func NewWSHandler(chatService service.ChatService) WSHandler {
	return &implWSHandler{chatService: chatService}
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

		fmt.Println("Received:", string(message))

		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
	}
}
