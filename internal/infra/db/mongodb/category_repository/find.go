package category_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	presentationHelpers "github.com/anuntech/finance-backend/internal/presentation/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindCategoriesRepository struct {
	Db *mongo.Database
}

func NewFindCategoriesRepository(db *mongo.Database) *FindCategoriesRepository {
	return &FindCategoriesRepository{
		Db: db,
	}
}

func (r *FindCategoriesRepository) Find(globalFilters *presentationHelpers.GlobalFilterParams) ([]models.Category, error) {
	collection := r.Db.Collection("category")

	filter := bson.M{"workspace_id": globalFilters.WorkspaceId}
	if globalFilters.Type != "" {
		filter["type"] = globalFilters.Type
	}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var categories []models.Category
	if err = cursor.All(context.Background(), &categories); err != nil {
		return nil, err
	}

	if globalFilters.Month == 0 {
		return categories, nil
	}

	for _, category := range categories {
		for i := range category.SubCategories {
			category.SubCategories[i].Amount = r.calculateSubCategoryBalance(category.SubCategories[i].Id, globalFilters)
		}
	}

	return categories, nil
}

func (c *FindCategoriesRepository) calculateSubCategoryBalance(subCategoryId primitive.ObjectID, globalFilters *presentationHelpers.GlobalFilterParams) float64 {
	return c.calculateDoNotRepeatSubCategoryBalance(subCategoryId, globalFilters)
}

func (c *FindCategoriesRepository) calculateDoNotRepeatSubCategoryBalance(subCategoryId primitive.ObjectID, globalFilters *presentationHelpers.GlobalFilterParams) float64 {
	collection := c.Db.Collection("transaction")
	startOfMonth := time.Date(globalFilters.Year, time.Month(globalFilters.Month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	filter := bson.M{
		"workspace_id":    globalFilters.WorkspaceId,
		"sub_category_id": subCategoryId,
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

	return helpers.CalculateTransactionBalance(transactions)
}
