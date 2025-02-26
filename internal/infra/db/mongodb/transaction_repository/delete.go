package transaction_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DeleteTransactionRepository struct {
	Db *mongo.Database
}

func NewDeleteTransactionRepository(db *mongo.Database) *DeleteTransactionRepository {
	return &DeleteTransactionRepository{
		Db: db,
	}
}

func (r *DeleteTransactionRepository) Delete(transactionIds []primitive.ObjectID, workspaceId primitive.ObjectID) error {
	collection := r.Db.Collection("transaction")

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err := collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": transactionIds}, "workspace_id": workspaceId})
	if err != nil {
		return err
	}

	return nil
}
