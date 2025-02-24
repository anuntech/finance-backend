package account_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
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

func (u *UpdateAccountMongoRepository) Update(id primitive.ObjectID, account *models.AccountInput) (*models.Account, error) {
	collection := u.Db.Collection("account")

	update := bson.M{
		"$set": bson.M{
			"name":       account.Name,
			"bank_id":    account.BankId,
			"balance":    account.Balance,
			"updated_at": time.Now(),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	result := collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, update)
	if result.Err() != nil {
		return nil, result.Err()
	}

	var updatedAccount models.Account
	if err := result.Decode(&updatedAccount); err != nil {
		return nil, err
	}

	return &updatedAccount, nil
}
