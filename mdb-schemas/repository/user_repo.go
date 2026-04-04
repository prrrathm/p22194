package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"p22194.prrrathm.com/mdb/models"
)

// UserRepo performs CRUD operations on the `users` collection.
type UserRepo struct {
	col *mongo.Collection
}

// NewUserRepo creates a UserRepo backed by the named database.
func NewUserRepo(db *mongo.Database) *UserRepo {
	return &UserRepo{col: db.Collection("users")}
}

// EnsureIndexes creates unique indexes on email and username.
// Safe to call on every startup — MongoDB is idempotent for identical index specs.
func (r *UserRepo) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true)},
	}
	if _, err := r.col.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("user_repo: ensure indexes: %w", err)
	}
	return nil
}

// Create inserts a new user document. Populates u.ID with the generated ObjectID on success.
// Propagates mongo.WriteException unwrapped so the service layer can detect code 11000 (duplicate key).
func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
	result, err := r.col.InsertOne(ctx, u)
	if err != nil {
		return fmt.Errorf("user_repo: insert: %w", err)
	}
	u.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

// FindByEmail returns the user with the given email.
// Returns mongo.ErrNoDocuments when not found.
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err != nil {
		return nil, fmt.Errorf("user_repo: find by email: %w", err)
	}
	return &u, nil
}

// FindByID returns the user with the given ObjectID.
// Returns mongo.ErrNoDocuments when not found.
func (r *UserRepo) FindByID(ctx context.Context, id bson.ObjectID) (*models.User, error) {
	var u models.User
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if err != nil {
		return nil, fmt.Errorf("user_repo: find by id: %w", err)
	}
	return &u, nil
}
