package edit_transaction_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DeleteEditTransactionRepository struct {
	Db *mongo.Database
}

func NewDeleteEditTransactionRepository(db *mongo.Database) *DeleteEditTransactionRepository {
	return &DeleteEditTransactionRepository{
		Db: db,
	}
}

func (r *DeleteEditTransactionRepository) Delete(editTransactionParams []struct {
	MainId      primitive.ObjectID
	MainCount   int
	WorkspaceId primitive.ObjectID
}) error {
	collection := r.Db.Collection("edit_transaction")

	if len(editTransactionParams) == 0 {
		return nil
	}

	// All items share the same workspace ID
	workspaceId := editTransactionParams[0].WorkspaceId

	// Create a filter for multiple (mainId, mainCount) pairs
	var orConditions []bson.M
	for _, param := range editTransactionParams {
		orConditions = append(orConditions, bson.M{
			"main_id":    param.MainId,
			"main_count": param.MainCount,
		})
	}

	filter := bson.M{
		"$or":          orConditions,
		"workspace_id": workspaceId,
	}

	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"updated_at": time.Now().UTC(),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err := collection.UpdateMany(ctx, filter, update)
	return err
}
