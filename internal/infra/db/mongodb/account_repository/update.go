package account_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UpdateAccountMongoRepository struct {
	Db *mongo.Database
}

func NewUpdateAccountMongoRepository(db *mongo.Database) *UpdateAccountMongoRepository {
	return &UpdateAccountMongoRepository{
		Db: db,
	}
}

func (u *UpdateAccountMongoRepository) Update(id string, account *models.AccountInput) (*models.Account, error) {
	collection := u.Db.Collection("account")

	update := bson.M{
		"$set": bson.M{
			"name":       account.Name,
			"bank_id":    account.BankId,
			"balance":    account.Balance,
			"updated_at": time.Now(),
		},
	}

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	result := collection.FindOneAndUpdate(context.Background(), bson.M{"_id": objectId}, update)
	if result.Err() != nil {
		return nil, result.Err()
	}

	var updatedAccount models.Account
	if err := result.Decode(&updatedAccount); err != nil {
		return nil, err
	}

	return &updatedAccount, nil
}
