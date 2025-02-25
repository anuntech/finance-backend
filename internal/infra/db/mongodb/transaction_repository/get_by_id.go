package transaction_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type GetTransactionByIdRepository struct {
	Db *mongo.Database
}

func NewGetTransactionByIdRepository(db *mongo.Database) *GetTransactionByIdRepository {
	return &GetTransactionByIdRepository{
		Db: db,
	}
}

func (r *GetTransactionByIdRepository) Find(transactionId primitive.ObjectID, workspaceId primitive.ObjectID) (*models.Transaction, error) {
	collection := r.Db.Collection("transaction")

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	var transaction models.Transaction

	err := collection.FindOne(ctx, bson.M{"_id": transactionId, "workspace_id": workspaceId}).Decode(&transaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &transaction, nil
}
