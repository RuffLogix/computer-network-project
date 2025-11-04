package service

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rufflogix/computer-network-project/internal/entity"
)

type RoomService interface {
	AddClient(*websocket.Conn, int64)
	RemoveClient(int64)
	Broadcast([]byte, int64)
	JoinRoom(userID, chatID int64) bool
	LeaveRoom(userID, chatID int64)
	BroadcastToRoom(chatID int64, message []byte)
	BroadcastToRoomExcept(chatID int64, message []byte, excludeUserID int64)
	SendToUser(userID int64, event entity.Event)
	GetOnlineUsers() []int64
}

type implRoomService struct {
	clients   map[int64]*websocket.Conn // userID -> connection
	rooms     map[int64]map[int64]bool  // chatID -> set of userIDs
	userRooms map[int64]map[int64]bool  // userID -> set of chatIDs
	mutex     sync.RWMutex
}

func NewRoomService() RoomService {
	return &implRoomService{
		clients:   make(map[int64]*websocket.Conn),
		rooms:     make(map[int64]map[int64]bool),
		userRooms: make(map[int64]map[int64]bool),
	}
}

func (s *implRoomService) AddClient(conn *websocket.Conn, id int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.clients[id] = conn

	if s.userRooms[id] == nil {
		s.userRooms[id] = make(map[int64]bool)
	}
}

func (s *implRoomService) RemoveClient(id int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Remove from all rooms
	if rooms, ok := s.userRooms[id]; ok {
		for chatID := range rooms {
			if s.rooms[chatID] != nil {
				delete(s.rooms[chatID], id)
			}
		}
		delete(s.userRooms, id)
	}

	delete(s.clients, id)
}

func (s *implRoomService) Broadcast(message []byte, senderID int64) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for id, conn := range s.clients {
		if senderID == id {
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Error broadcasting to user %d: %v", id, err)
		}
	}
}

func (s *implRoomService) JoinRoom(userID, chatID int64) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if user is already in the room
	if s.rooms[chatID] != nil && s.rooms[chatID][userID] {
		return false // User already in room, don't broadcast
	}

	if s.rooms[chatID] == nil {
		s.rooms[chatID] = make(map[int64]bool)
	}
	s.rooms[chatID][userID] = true

	if s.userRooms[userID] == nil {
		s.userRooms[userID] = make(map[int64]bool)
	}
	s.userRooms[userID][chatID] = true

	return true // User newly joined, should broadcast
}

func (s *implRoomService) LeaveRoom(userID, chatID int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.rooms[chatID] != nil {
		delete(s.rooms[chatID], userID)
	}

	if s.userRooms[userID] != nil {
		delete(s.userRooms[userID], chatID)
	}
}

func (s *implRoomService) BroadcastToRoom(chatID int64, message []byte) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	roomMembers, ok := s.rooms[chatID]
	if !ok {
		return
	}

	for userID := range roomMembers {
		if conn, ok := s.clients[userID]; ok {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error sending to user %d in room %d: %v", userID, chatID, err)
			}
		}
	}
}

func (s *implRoomService) BroadcastToRoomExcept(chatID int64, message []byte, excludeUserID int64) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	roomMembers, ok := s.rooms[chatID]
	if !ok {
		return
	}

	for userID := range roomMembers {
		if userID == excludeUserID {
			continue
		}

		if conn, ok := s.clients[userID]; ok {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error sending to user %d in room %d: %v", userID, chatID, err)
			}
		}
	}
}

func (s *implRoomService) SendToUser(userID int64, event entity.Event) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	conn, ok := s.clients[userID]
	if !ok {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("Error sending to user %d: %v", userID, err)
	}
}

func (s *implRoomService) GetOnlineUsers() []int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	userIDs := make([]int64, 0, len(s.clients))
	for userID := range s.clients {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}
