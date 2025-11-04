package repository

import (
	"context"
	"time"

	"github.com/rufflogix/computer-network-project/internal/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository interface {
	CreateUser(user *entity.User) error
	GetUserByID(id primitive.ObjectID) (*entity.User, error)
	GetUserByNumericID(id int64) (*entity.User, error)
	GetUserByUsername(username string) (*entity.User, error)
	GetUserByEmail(email string) (*entity.User, error)
	UpdateUser(user *entity.User) error
	DeleteUser(id primitive.ObjectID) error
	GetAllUsers() ([]*entity.User, error)
}

type implUserRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &implUserRepository{
		db:         db,
		collection: db.Collection("users"),
	}
}

func (r *implUserRepository) CreateUser(user *entity.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Generate numeric ID
	countersCol := r.db.Collection("counters")
	var result struct {
		Seq int64 `bson:"seq"`
	}

	err := countersCol.FindOneAndUpdate(
		ctx,
		bson.M{"_id": "user_id"},
		bson.M{
			"$inc":         bson.M{"seq": int64(1)},
			"$setOnInsert": bson.M{"_id": "user_id"},
		},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&result)

	if err != nil {
		return err
	}

	user.NumericID = result.Seq
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	insertResult, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	user.ID = insertResult.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *implUserRepository) GetUserByID(id primitive.ObjectID) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user entity.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *implUserRepository) GetUserByNumericID(id int64) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user entity.User
	err := r.collection.FindOne(ctx, bson.M{"numeric_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *implUserRepository) GetUserByUsername(username string) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user entity.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *implUserRepository) GetUserByEmail(email string) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user entity.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *implUserRepository) UpdateUser(user *entity.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user.UpdatedAt = time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": user},
	)

	return err
}

func (r *implUserRepository) DeleteUser(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *implUserRepository) GetAllUsers() ([]*entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*entity.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}
