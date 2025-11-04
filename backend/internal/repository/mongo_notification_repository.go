package repository

import (
	"context"
	"time"

	"github.com/rufflogix/computer-network-project/internal/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoNotificationRepository struct {
	db                  *mongo.Database
	collection          *mongo.Collection
	notificationCounter *mongo.Collection
}

func NewMongoNotificationRepository(db *mongo.Database) NotificationRepository {
	repo := &MongoNotificationRepository{
		db:                  db,
		collection:          db.Collection("notifications"),
		notificationCounter: db.Collection("counters"),
	}

	// Initialize counter for notification IDs if not exists
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repo.notificationCounter.FindOneAndUpdate(
		ctx,
		bson.M{"_id": "notification_id"},
		bson.M{"$setOnInsert": bson.M{"seq": int64(0)}},
		options.FindOneAndUpdate().SetUpsert(true),
	)

	return repo
}

func (r *MongoNotificationRepository) getNextID() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result struct {
		Seq int64 `bson:"seq"`
	}
	err := r.notificationCounter.FindOneAndUpdate(
		ctx,
		bson.M{"_id": "notification_id"},
		bson.M{"$inc": bson.M{"seq": 1}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&result)
	if err != nil {
		return 0, err
	}
	return result.Seq, nil
}

func (r *MongoNotificationRepository) CreateNotification(notification *entity.Notification) error {
	id, err := r.getNextID()
	if err != nil {
		return err
	}

	notification.ID = id
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()
	notification.Status = entity.NotificationUnread

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = r.collection.InsertOne(ctx, notification)
	return err
}

func (r *MongoNotificationRepository) GetNotificationByID(id int64) (*entity.Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var notification entity.Notification
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&notification)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &notification, nil
}

func (r *MongoNotificationRepository) GetNotificationsByUser(userID int64) ([]*entity.Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"recipient_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []*entity.Notification
	for cursor.Next(ctx) {
		var notification entity.Notification
		if err := cursor.Decode(&notification); err != nil {
			return nil, err
		}
		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

func (r *MongoNotificationRepository) GetUnreadNotificationsByUser(userID int64) ([]*entity.Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{
		"recipient_id": userID,
		"status":       entity.NotificationUnread,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []*entity.Notification
	for cursor.Next(ctx) {
		var notification entity.Notification
		if err := cursor.Decode(&notification); err != nil {
			return nil, err
		}
		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

func (r *MongoNotificationRepository) UpdateNotificationStatus(id int64, status entity.NotificationStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"id": id}, update)
	return err
}

func (r *MongoNotificationRepository) DeleteNotification(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
	return err
}
