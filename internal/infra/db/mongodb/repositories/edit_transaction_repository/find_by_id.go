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

func (r *FindByIdEditTransactionRepository) Find(mainId primitive.ObjectID, mainCount int, workspaceId primitive.ObjectID) (*models.Transaction, error) {
	collection := r.Db.Collection("edit_transaction")

	filter := bson.M{"main_id": mainId, "workspace_id": workspaceId, "main_count": mainCount}
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

func (r *FindByIdEditTransactionRepository) FindMany(params []struct {
	MainId      primitive.ObjectID
	MainCount   int
	WorkspaceId primitive.ObjectID
}) ([]*models.Transaction, error) {
	if len(params) == 0 {
		return []*models.Transaction{}, nil
	}

	collection := r.Db.Collection("edit_transaction")

	// Build the query with $or to fetch multiple transactions in one query
	var orConditions []bson.M
	for _, param := range params {
		orConditions = append(orConditions, bson.M{
			"main_id":      param.MainId,
			"main_count":   param.MainCount,
			"workspace_id": param.WorkspaceId,
		})
	}

	filter := bson.M{"$or": orConditions}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []*models.Transaction
	if err = cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}
