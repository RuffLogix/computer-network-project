package entity

type EventType string

const (
	JOIN            EventType = "join"
	LEAVE           EventType = "leave"
	SEND_MESSAGE    EventType = "send_message"
	DELETE_MESSAGE  EventType = "delete_message"
	EDIT_MESSAGE    EventType = "edit_message"
	ADD_REACTION    EventType = "add_reaction"
	REMOVE_REACTION EventType = "remove_reaction"
	TYPING          EventType = "typing"
	NOTIFICATION    EventType = "notification"
	FRIEND_INVITE   EventType = "friend_invite"
	GROUP_INVITE    EventType = "group_invite"
)

type Event struct {
	Type      EventType              `json:"type"`
	Data      map[string]interface{} `json:"data"`
	CreatedBy int64                  `json:"created_by"`
}
