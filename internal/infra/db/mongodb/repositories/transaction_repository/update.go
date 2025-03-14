package transaction_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UpdateTransactionRepository struct {
	Db *mongo.Database
}

func NewUpdateTransactionRepository(db *mongo.Database) *UpdateTransactionRepository {
	return &UpdateTransactionRepository{
		Db: db,
	}
}

func (r *UpdateTransactionRepository) Update(transactionId primitive.ObjectID, transaction *models.Transaction) (*models.Transaction, error) {
	collection := r.Db.Collection("transaction")

	transaction.UpdatedAt = time.Now().UTC()

	update := bson.M{
		"$set": transaction,
	}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err := collection.UpdateOne(ctx, bson.M{"_id": transactionId, "workspace_id": transaction.WorkspaceId}, update)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}
