package transaction_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
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

	editTransactionCollection := r.Db.Collection("edit_transaction")
	_, err = editTransactionCollection.DeleteMany(ctx, bson.M{"main_id": bson.M{"$in": transactionIds}})
	if err != nil {
		return err
	}

	return nil
}

func (r *DeleteTransactionRepository) DeleteEditTransactions(editTransactionParams []struct {
	MainId      primitive.ObjectID
	MainCount   int
	WorkspaceId primitive.ObjectID
}, findTransactionById usecase.FindTransactionByIdRepository) error {
	if len(editTransactionParams) == 0 {
		return nil
	}

	collection := r.Db.Collection("edit_transaction")
	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	for _, param := range editTransactionParams {
		// First check if the edit transaction already exists
		filter := bson.M{
			"main_id":      param.MainId,
			"main_count":   param.MainCount,
			"workspace_id": param.WorkspaceId,
		}

		var existingEditTx models.Transaction
		err := collection.FindOne(ctx, filter).Decode(&existingEditTx)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				// Edit transaction doesn't exist, need to create it
				// First get the main transaction to copy its data
				mainTransaction, err := findTransactionById.Find(param.MainId, param.WorkspaceId)
				if err != nil || mainTransaction == nil {
					// Skip if main transaction doesn't exist
					continue
				}

				// Create a copy of the main transaction
				newEditTx := *mainTransaction
				newEditTx.Id = primitive.NewObjectID()
				newEditTx.MainId = &param.MainId
				newEditTx.MainCount = &param.MainCount
				newEditTx.IsDeleted = true
				newEditTx.CreatedAt = time.Now().UTC()
				newEditTx.UpdatedAt = time.Now().UTC()

				// Insert the new edit transaction
				_, err = collection.InsertOne(ctx, newEditTx)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// Edit transaction exists, update it to mark as deleted
			update := bson.M{
				"$set": bson.M{
					"is_deleted": true,
					"updated_at": time.Now().UTC(),
				},
			}

			_, err = collection.UpdateOne(ctx, filter, update)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
