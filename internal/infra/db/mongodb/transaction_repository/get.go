package transaction_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/domain/usecase"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionRepository struct {
	db *mongo.Database
}

func NewTransactionRepository(db *mongo.Database) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Find(filters *usecase.FindTransactionsByWorkspaceIdInputRepository) ([]models.Transaction, error) {
	collection := r.db.Collection("transactions")

	startOfMonth := time.Date(filters.Year, time.Month(filters.Month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	filter := bson.M{
		"workspace_id": filters.WorkspaceId,
		"due_date": bson.M{
			"$gte": startOfMonth,
			"$lt":  endOfMonth,
		},
		"confirmation_date": bson.M{
			"$gte": startOfMonth,
			"$lt":  endOfMonth,
		},
		"type": filters.Type,
	}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err = cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}
