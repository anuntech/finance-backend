package edit_transaction_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindByIdEditTransactionRepository struct {
	Db *mongo.Database
}

func NewFindByIdEditTransactionRepository(db *mongo.Database) *FindByIdEditTransactionRepository {
	return &FindByIdEditTransactionRepository{
		Db: db,
	}
}

func (r *FindByIdEditTransactionRepository) Find(id primitive.ObjectID, workspaceId primitive.ObjectID) (*models.Transaction, error) {
	collection := r.Db.Collection("edit_transaction")

	filter := bson.M{"main_id": id, "workspace_id": workspaceId}
	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	cursor := collection.FindOne(ctx, filter)
	if cursor.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if cursor.Err() != nil {
		return nil, cursor.Err()
	}

	var transaction models.Transaction
	if err := cursor.Decode(&transaction); err != nil {
		return nil, err
	}

	return &transaction, nil
}
