package helpers

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
)

type CalculateCategoryBalanceParams struct {
	Db            *mongo.Database
	GlobalFilters *helpers.GlobalFilterParams
	CategoryId    primitive.ObjectID
}

func NewCalculateCategoryBalanceParams(db *mongo.Database, globalFilters *helpers.GlobalFilterParams, categoryId primitive.ObjectID) *CalculateCategoryBalanceParams {
	return &CalculateCategoryBalanceParams{
		Db:            db,
		GlobalFilters: globalFilters,
		CategoryId:    categoryId,
	}
}

func (p *CalculateCategoryBalanceParams) CalculateCategoryBalance() float64 {
	return p.calculateDoNotRepeatCategoryBalance()
}

func (p *CalculateCategoryBalanceParams) calculateDoNotRepeatCategoryBalance() float64 {
	collection := p.Db.Collection("transaction")
	startOfMonth := time.Date(p.GlobalFilters.Year, time.Month(p.GlobalFilters.Month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	filter := bson.M{
		"workspace_id": p.GlobalFilters.WorkspaceId,
		"category_id":  p.CategoryId,
		"$or": []bson.M{
			{
				"due_date": bson.M{
					"$lt": endOfMonth,
				},
				"is_confirmed": false,
			},
			{
				"confirmation_date": bson.M{
					"$lt": endOfMonth,
				},
				"is_confirmed": true,
			},
		},
		"frequency": "DO_NOT_REPEAT",
	}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return 0.0
	}

	var transactions []models.Transaction
	if err := cursor.All(context.Background(), &transactions); err != nil {
		return 0.0
	}

	return CalculateTransactionBalance(transactions)
}
