package transaction_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateTransactionRepository struct {
	Db *mongo.Database
}

func NewCreateTransactionRepository(db *mongo.Database) *CreateTransactionRepository {
	return &CreateTransactionRepository{
		Db: db,
	}
}

func (r *CreateTransactionRepository) Create(transaction *models.Transaction) (*models.Transaction, error) {
	collection := r.Db.Collection("transaction")

	transaction.Id = primitive.NewObjectID()
	transaction.CreatedAt = time.Now()
	transaction.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()
	_, err := collection.InsertOne(ctx, transaction)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}
