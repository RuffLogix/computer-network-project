package repository

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateChatIndexes creates indexes for better query performance
func CreateChatIndexes(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Chats collection indexes
	chatsCol := db.Collection("chats")
	_, err := chatsCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "is_public", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "type", Value: 1}},
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to create chats indexes: %v", err)
		return err
	}

	// Messages collection indexes
	messagesCol := db.Collection("messages")
	_, err = messagesCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "chat_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{{Key: "created_by", Value: 1}},
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to create messages indexes: %v", err)
		return err
	}

	// Reactions collection indexes
	reactionsCol := db.Collection("reactions")
	_, err = reactionsCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "message_id", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "message_id", Value: 1},
				{Key: "user_id", Value: 1},
				{Key: "type", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to create reactions indexes: %v", err)
		return err
	}

	// Chat members collection indexes
	chatMembersCol := db.Collection("chat_members")
	_, err = chatMembersCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "chat_id", Value: 1},
				{Key: "user_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "chat_id", Value: 1}},
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to create chat_members indexes: %v", err)
		return err
	}

	log.Println("Database indexes created successfully")
	return nil
}
