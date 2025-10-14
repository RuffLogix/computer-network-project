package service

import (
	"sync"

	"github.com/gorilla/websocket"
)

type RoomService interface {
	AddClient(*websocket.Conn, int64)
	RemoveClient(int64)
	Broadcast([]byte, int64)
}

type implRoomService struct {
	clients map[int64]*websocket.Conn
	mutex   sync.RWMutex
}

func NewRoomService() RoomService {
	return &implRoomService{clients: make(map[int64]*websocket.Conn)}
}

func (s *implRoomService) AddClient(conn *websocket.Conn, id int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.clients[id] = conn
}

func (s *implRoomService) RemoveClient(id int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.clients, id)
}

func (s *implRoomService) Broadcast(message []byte, senderID int64) {
	for id, conn := range s.clients {
		if senderID == id {
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
	}
}
