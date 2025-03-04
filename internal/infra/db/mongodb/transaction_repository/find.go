package transaction_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	presentationHelpers "github.com/anuntech/finance-backend/internal/presentation/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionRepository struct {
	db *mongo.Database
}

func NewTransactionRepository(db *mongo.Database) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Find(filters *presentationHelpers.GlobalFilterParams) ([]models.Transaction, error) {
	collection := r.db.Collection("transaction")

	startOfMonth := time.Date(filters.Year, time.Month(filters.Month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	filter := bson.M{
		"workspace_id": filters.WorkspaceId,
	}
	if filters.Type != "" {
		filter["type"] = filters.Type
	}

	if filters.Month != 0 {
		orRepeatAndRecurringLogic := []bson.M{
			{
				"$and": []bson.M{
					{"is_confirmed": false},
					{"due_date": bson.M{"$lt": endOfMonth}},
				},
			},
			{
				"$and": []bson.M{
					{"is_confirmed": true},
					{"confirmation_date": bson.M{"$lt": endOfMonth}},
				},
			},
		}

		filter["$or"] = []bson.M{
			{
				"frequency": "DO_NOT_REPEAT",
				"$or": []bson.M{
					{
						"$and": []bson.M{
							{"is_confirmed": false},
							{"due_date": bson.M{"$gte": startOfMonth, "$lt": endOfMonth}},
						},
					},
					{
						"$and": []bson.M{
							{"is_confirmed": true},
							{"confirmation_date": bson.M{"$gte": startOfMonth, "$lt": endOfMonth}},
						},
					},
				},
			},
			{
				"frequency": "RECURRING",
				"$or":       orRepeatAndRecurringLogic,
			},
			{
				"frequency": "REPEAT",
				"$or":       orRepeatAndRecurringLogic,
			},
		}
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
