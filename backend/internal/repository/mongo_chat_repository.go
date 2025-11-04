package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rufflogix/computer-network-project/internal/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoChatRepository struct {
	db             *mongo.Database
	chatsCol       *mongo.Collection
	messagesCol    *mongo.Collection
	reactionsCol   *mongo.Collection
	chatMembersCol *mongo.Collection
	chatIDCounter  *mongo.Collection
}

func NewMongoChatRepository(db *mongo.Database) ChatRepository {
	repo := &MongoChatRepository{
		db:             db,
		chatsCol:       db.Collection("chats"),
		messagesCol:    db.Collection("messages"),
		reactionsCol:   db.Collection("reactions"),
		chatMembersCol: db.Collection("chat_members"),
		chatIDCounter:  db.Collection("counters"),
	}

	// Initialize counter for chat IDs if not exists
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repo.chatIDCounter.FindOneAndUpdate(
		ctx,
		bson.M{"_id": "chat_id"},
		bson.M{"$setOnInsert": bson.M{"seq": int64(0)}},
		options.FindOneAndUpdate().SetUpsert(true),
	)

	repo.chatIDCounter.FindOneAndUpdate(
		ctx,
		bson.M{"_id": "message_id"},
		bson.M{"$setOnInsert": bson.M{"seq": int64(0)}},
		options.FindOneAndUpdate().SetUpsert(true),
	)

	repo.chatIDCounter.FindOneAndUpdate(
		ctx,
		bson.M{"_id": "reaction_id"},
		bson.M{"$setOnInsert": bson.M{"seq": int64(0)}},
		options.FindOneAndUpdate().SetUpsert(true),
	)

	if err := repo.migrateLegacyFields(); err != nil {
		log.Printf("warning: failed to migrate legacy chat fields: %v", err)
	}

	return repo
}

func (r *MongoChatRepository) migrateLegacyFields() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	renames := []struct {
		col    *mongo.Collection
		fromTo map[string]string
	}{
		{
			col: r.chatsCol,
			fromTo: map[string]string{
				"ispublic":  "is_public",
				"createdby": "created_by",
				"createdat": "created_at",
				"updatedat": "updated_at",
			},
		},
		{
			col: r.messagesCol,
			fromTo: map[string]string{
				"chatid":    "chat_id",
				"mediaurl":  "media_url",
				"replytoid": "reply_to_id",
				"createdby": "created_by",
				"createdat": "created_at",
				"updatedat": "updated_at",
			},
		},
		{
			col: r.reactionsCol,
			fromTo: map[string]string{
				"messageid": "message_id",
				"createdby": "created_by",
				"createdat": "created_at",
			},
		},
		{
			col: r.chatMembersCol,
			fromTo: map[string]string{
				"chatid":   "chat_id",
				"userid":   "user_id",
				"joinedat": "joined_at",
			},
		},
	}

	for _, item := range renames {
		for from, to := range item.fromTo {
			if err := r.renameField(ctx, item.col, from, to); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *MongoChatRepository) renameField(ctx context.Context, col *mongo.Collection, from, to string) error {
	filter := bson.M{from: bson.M{"$exists": true}}
	update := bson.M{"$rename": bson.M{from: to}}

	if _, err := col.UpdateMany(ctx, filter, update); err != nil {
		return err
	}

	return nil
}

func (r *MongoChatRepository) getNextSequence(name string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result struct {
		Seq int64 `bson:"seq"`
	}

	err := r.chatIDCounter.FindOneAndUpdate(
		ctx,
		bson.M{"_id": name},
		bson.M{"$inc": bson.M{"seq": int64(1)}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&result)

	if err != nil {
		return 0, err
	}

	return result.Seq, nil
}

// Chat operations
func (r *MongoChatRepository) CreateChat(chat *entity.Chat) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Generate new ID
	id, err := r.getNextSequence("chat_id")
	if err != nil {
		return err
	}

	chat.ID = id
	chat.CreatedAt = time.Now()
	chat.UpdatedAt = time.Now()

	_, err = r.chatsCol.InsertOne(ctx, chat)
	return err
}

func (r *MongoChatRepository) GetChatByID(id int64) (*entity.Chat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var chat entity.Chat
	err := r.chatsCol.FindOne(ctx, bson.M{"id": id}).Decode(&chat)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("chat not found")
		}
		return nil, err
	}

	return &chat, nil
}

func (r *MongoChatRepository) GetChatsByUser(userID int64) ([]*entity.Chat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find all chat memberships for this user
	cursor, err := r.chatMembersCol.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []*entity.ChatMember
	if err = cursor.All(ctx, &members); err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return []*entity.Chat{}, nil
	}

	// Extract chat IDs
	chatIDs := make([]int64, len(members))
	for i, member := range members {
		chatIDs[i] = member.ChatID
	}

	// Find all chats
	cursor, err = r.chatsCol.Find(ctx, bson.M{"id": bson.M{"$in": chatIDs}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var chats []*entity.Chat
	if err = cursor.All(ctx, &chats); err != nil {
		return nil, err
	}

	return chats, nil
}

func (r *MongoChatRepository) GetPublicChats() ([]*entity.Chat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.chatsCol.Find(ctx, bson.M{"is_public": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var chats []*entity.Chat
	if err = cursor.All(ctx, &chats); err != nil {
		return nil, err
	}

	return chats, nil
}

func (r *MongoChatRepository) UpdateChat(chat *entity.Chat) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chat.UpdatedAt = time.Now()

	_, err := r.chatsCol.UpdateOne(
		ctx,
		bson.M{"id": chat.ID},
		bson.M{"$set": chat},
	)
	return err
}

func (r *MongoChatRepository) DeleteChat(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.chatsCol.DeleteOne(ctx, bson.M{"id": id})
	return err
}

// Message operations
func (r *MongoChatRepository) CreateMessage(message *entity.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Generate new ID
	id, err := r.getNextSequence("message_id")
	if err != nil {
		return err
	}

	message.ID = id
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()

	// Populate reply_to if reply_to_id is set
	if message.ReplyToID != nil {
		replyMsg, err := r.GetMessageByID(*message.ReplyToID)
		if err == nil {
			message.ReplyTo = replyMsg
		}
	}

	_, err = r.messagesCol.InsertOne(ctx, message)
	return err
}

func (r *MongoChatRepository) GetMessageByID(id int64) (*entity.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var message entity.Message
	err := r.messagesCol.FindOne(ctx, bson.M{"id": id}).Decode(&message)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("message not found")
		}
		return nil, err
	}

	return &message, nil
}

func (r *MongoChatRepository) GetMessagesByChat(chatID int64, limit, offset int) ([]*entity.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.messagesCol.Find(ctx, bson.M{"chat_id": chatID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*entity.Message
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	// Populate user information and reactions for each message
	usersCol := r.db.Collection("users")
	for _, msg := range messages {
		var user entity.User
		err := usersCol.FindOne(ctx, bson.M{"numeric_id": msg.CreatedBy}).Decode(&user)
		if err == nil {
			msg.CreatedByUser = &user
		}

		// If message has a reply, fetch the replied message
		if msg.ReplyToID != nil {
			replyMsg, err := r.GetMessageByID(*msg.ReplyToID)
			if err == nil {
				msg.ReplyTo = replyMsg
			}
		}

		// Load reactions for this message
		reactions, err := r.GetReactionsByMessage(msg.ID)
		if err == nil {
			msg.Reactions = reactions
		}
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *MongoChatRepository) UpdateMessage(message *entity.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message.UpdatedAt = time.Now()

	_, err := r.messagesCol.UpdateOne(
		ctx,
		bson.M{"id": message.ID},
		bson.M{"$set": message},
	)
	return err
}

func (r *MongoChatRepository) DeleteMessage(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.messagesCol.DeleteOne(ctx, bson.M{"id": id})
	return err
}

// Reaction operations
func (r *MongoChatRepository) CreateReaction(reaction *entity.Reaction, userID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if reaction record already exists for this message and type
	filter := bson.M{
		"message_id": reaction.MessageID,
		"type":       reaction.Type,
	}

	var existingReaction entity.Reaction
	err := r.reactionsCol.FindOne(ctx, filter).Decode(&existingReaction)
	if err == nil {
		// Ensure user_ids is initialized for backward compatibility
		if existingReaction.UserIDs == nil {
			existingReaction.UserIDs = []int64{}
		}

		// Check if user already reacted
		userIndex := -1
		for i, uid := range existingReaction.UserIDs {
			if uid == userID {
				userIndex = i
				break
			}
		}

		if userIndex >= 0 {
			// User already reacted, remove them (toggle off)
			existingReaction.UserIDs = append(existingReaction.UserIDs[:userIndex], existingReaction.UserIDs[userIndex+1:]...)
			existingReaction.Count = len(existingReaction.UserIDs)
			existingReaction.UpdatedAt = time.Now()

			if existingReaction.Count == 0 {
				// No users left, delete the reaction
				_, err = r.reactionsCol.DeleteOne(ctx, filter)
				return err
			} else {
				// Update the reaction
				_, err = r.reactionsCol.UpdateOne(ctx, filter, bson.M{
					"$set": bson.M{
						"user_ids":   existingReaction.UserIDs,
						"count":      existingReaction.Count,
						"updated_at": existingReaction.UpdatedAt,
					},
				})
				return err
			}
		} else {
			// User hasn't reacted, add them (toggle on)
			existingReaction.UserIDs = append(existingReaction.UserIDs, userID)
			existingReaction.Count = len(existingReaction.UserIDs)
			existingReaction.UpdatedAt = time.Now()

			_, err = r.reactionsCol.UpdateOne(ctx, filter, bson.M{
				"$set": bson.M{
					"user_ids":   existingReaction.UserIDs,
					"count":      existingReaction.Count,
					"updated_at": existingReaction.UpdatedAt,
				},
			})
			return err
		}
	} else if err == mongo.ErrNoDocuments {
		// Reaction doesn't exist, create new one
		id, err := r.getNextSequence("reaction_id")
		if err != nil {
			return err
		}

		reaction.ID = id
		reaction.UserIDs = []int64{userID}
		reaction.Count = 1
		reaction.CreatedAt = time.Now()
		reaction.UpdatedAt = time.Now()

		_, err = r.reactionsCol.InsertOne(ctx, reaction)
		return err
	} else {
		return err
	}
}

func (r *MongoChatRepository) GetReactionsByMessage(messageID int64) ([]*entity.Reaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.reactionsCol.Find(ctx, bson.M{"message_id": messageID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reactions []*entity.Reaction
	if err = cursor.All(ctx, &reactions); err != nil {
		return nil, err
	}

	// Ensure user_ids is initialized for backward compatibility
	for _, reaction := range reactions {
		if reaction.UserIDs == nil {
			reaction.UserIDs = []int64{}
		}
	}

	return reactions, nil
}

func (r *MongoChatRepository) DeleteReaction(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.reactionsCol.DeleteOne(ctx, bson.M{"id": id})
	return err
}

// Chat member operations
func (r *MongoChatRepository) AddChatMember(member *entity.ChatMember) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if member already exists
	exists, err := r.IsChatMember(member.ChatID, member.UserID)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already a member
	}

	member.JoinedAt = time.Now()

	_, err = r.chatMembersCol.InsertOne(ctx, member)
	return err
}

func (r *MongoChatRepository) GetChatMembers(chatID int64) ([]*entity.ChatMember, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.chatMembersCol.Find(ctx, bson.M{"chat_id": chatID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []*entity.ChatMember
	if err = cursor.All(ctx, &members); err != nil {
		return nil, err
	}

	return members, nil
}

func (r *MongoChatRepository) RemoveChatMember(chatID, userID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.chatMembersCol.DeleteOne(ctx, bson.M{
		"chat_id": chatID,
		"user_id": userID,
	})
	return err
}

func (r *MongoChatRepository) IsChatMember(chatID, userID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := r.chatMembersCol.CountDocuments(ctx, bson.M{
		"chat_id": chatID,
		"user_id": userID,
	})
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
