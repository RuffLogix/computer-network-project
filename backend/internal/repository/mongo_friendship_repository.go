package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/rufflogix/computer-network-project/internal/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoFriendshipRepository struct {
	db                *mongo.Database
	collection        *mongo.Collection
	friendshipCounter *mongo.Collection
}

func NewMongoFriendshipRepository(db *mongo.Database) FriendshipRepository {
	repo := &MongoFriendshipRepository{
		db:                db,
		collection:        db.Collection("friendships"),
		friendshipCounter: db.Collection("counters"),
	}

	// Initialize counter for friendship IDs if not exists
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repo.friendshipCounter.FindOneAndUpdate(
		ctx,
		bson.M{"_id": "friendship_id"},
		bson.M{"$setOnInsert": bson.M{"seq": int64(0)}},
		options.FindOneAndUpdate().SetUpsert(true),
	)

	return repo
}

func (r *MongoFriendshipRepository) getNextID() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result struct {
		Seq int64 `bson:"seq"`
	}
	err := r.friendshipCounter.FindOneAndUpdate(
		ctx,
		bson.M{"_id": "friendship_id"},
		bson.M{"$inc": bson.M{"seq": 1}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&result)
	if err != nil {
		return 0, err
	}
	return result.Seq, nil
}

func (r *MongoFriendshipRepository) CreateFriendship(userID, friendID int64) (*entity.Friendship, error) {
	id, err := r.getNextID()
	if err != nil {
		return nil, err
	}

	friendship := &entity.Friendship{
		ID:        id,
		UserID:    userID,
		FriendID:  friendID,
		Status:    entity.Pending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = r.collection.InsertOne(ctx, friendship)
	if err != nil {
		return nil, err
	}

	return friendship, nil
}

func (r *MongoFriendshipRepository) GetFriendship(userID, friendID int64) (*entity.Friendship, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"user_id": userID, "friend_id": friendID},
			{"user_id": friendID, "friend_id": userID},
		},
	}

	var friendship entity.Friendship
	err := r.collection.FindOne(ctx, filter).Decode(&friendship)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("friendship not found")
		}
		return nil, err
	}

	return &friendship, nil
}

func (r *MongoFriendshipRepository) GetFriendshipsByUser(userID int64) ([]*entity.Friendship, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"user_id": userID},
			{"friend_id": userID},
		},
		"status": entity.Accepted,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var friendships []*entity.Friendship
	for cursor.Next(ctx) {
		var friendship entity.Friendship
		if err := cursor.Decode(&friendship); err != nil {
			return nil, err
		}
		friendships = append(friendships, &friendship)
	}

	return friendships, nil
}

func (r *MongoFriendshipRepository) GetPendingFriendships(userID int64) ([]*entity.Friendship, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"friend_id": userID,
		"status":    entity.Pending,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var friendships []*entity.Friendship
	for cursor.Next(ctx) {
		var friendship entity.Friendship
		if err := cursor.Decode(&friendship); err != nil {
			return nil, err
		}
		friendships = append(friendships, &friendship)
	}

	return friendships, nil
}

func (r *MongoFriendshipRepository) UpdateFriendshipStatus(userID, friendID int64, status entity.FriendshipStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"user_id": userID, "friend_id": friendID},
			{"user_id": friendID, "friend_id": userID},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *MongoFriendshipRepository) DeleteFriendship(userID, friendID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"user_id": userID, "friend_id": friendID},
			{"user_id": friendID, "friend_id": userID},
		},
	}

	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}
