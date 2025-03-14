package edit_transaction_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateEditTransactionRepository struct {
	Db *mongo.Database
}

func NewCreateEditTransactionRepository(db *mongo.Database) *CreateEditTransactionRepository {
	return &CreateEditTransactionRepository{
		Db: db,
	}
}

func (r *CreateEditTransactionRepository) Create(transaction *models.Transaction) (*models.Transaction, error) {
	collection := r.Db.Collection("edit_transaction")

	transaction.Id = primitive.NewObjectID()
	transaction.CreatedAt = time.Now().UTC()
	transaction.UpdatedAt = time.Now().UTC()

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()
	_, err := collection.InsertOne(ctx, transaction)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}
