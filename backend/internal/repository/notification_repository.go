package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/rufflogix/computer-network-project/internal/entity"
)

type NotificationRepository interface {
	CreateNotification(notification *entity.Notification) error
	GetNotificationByID(id int64) (*entity.Notification, error)
	GetNotificationsByUser(userID int64) ([]*entity.Notification, error)
	GetUnreadNotificationsByUser(userID int64) ([]*entity.Notification, error)
	UpdateNotificationStatus(id int64, status entity.NotificationStatus) error
	DeleteNotification(id int64) error
}

type implNotificationRepository struct {
	notifications  map[int64]*entity.Notification
	notificationID int64
	mu             sync.RWMutex
}

func NewNotificationRepository() NotificationRepository {
	return &implNotificationRepository{
		notifications: make(map[int64]*entity.Notification),
	}
}

func (r *implNotificationRepository) CreateNotification(notification *entity.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.notificationID++
	notification.ID = r.notificationID
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()
	notification.Status = entity.NotificationUnread

	r.notifications[notification.ID] = notification
	return nil
}

func (r *implNotificationRepository) GetNotificationByID(id int64) (*entity.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	notification, ok := r.notifications[id]
	if !ok {
		return nil, fmt.Errorf("notification not found")
	}

	return notification, nil
}

func (r *implNotificationRepository) GetNotificationsByUser(userID int64) ([]*entity.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var notifications []*entity.Notification
	for _, notif := range r.notifications {
		if notif.RecipientID == userID {
			notifications = append(notifications, notif)
		}
	}

	return notifications, nil
}

func (r *implNotificationRepository) GetUnreadNotificationsByUser(userID int64) ([]*entity.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var notifications []*entity.Notification
	for _, notif := range r.notifications {
		if notif.RecipientID == userID && notif.Status == entity.NotificationUnread {
			notifications = append(notifications, notif)
		}
	}

	return notifications, nil
}

func (r *implNotificationRepository) UpdateNotificationStatus(id int64, status entity.NotificationStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	notification, ok := r.notifications[id]
	if !ok {
		return fmt.Errorf("notification not found")
	}

	notification.Status = status
	notification.UpdatedAt = time.Now()
	return nil
}

func (r *implNotificationRepository) DeleteNotification(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.notifications, id)
	return nil
}
