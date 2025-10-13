package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader {
	CheckOrigin: func (r *http.Request) bool {
		return true
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
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

func main() {
	http.HandleFunc("/ws", wsHandler)
	err := http.ListenAndServe(":8080", nil)
	
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}