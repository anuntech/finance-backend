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
			category.SubCategories[i].Amount = r.calculateAllSubCategoryBalance(category.SubCategories[i].Id, globalFilters)
		}
	}

	return categories, nil
}

func (c *FindCategoriesRepository) calculateAllSubCategoryBalance(subCategoryId primitive.ObjectID, globalFilters *presentationHelpers.GlobalFilterParams) float64 {
	doNotRepeatBalance := c.calculateSubCategoryBalance(subCategoryId, globalFilters, "DO_NOT_REPEAT")
	recurringBalance := c.calculateSubCategoryBalance(subCategoryId, globalFilters, "RECURRING")
	repeatBalance := c.calculateSubCategoryBalance(subCategoryId, globalFilters, "REPEAT")

	return doNotRepeatBalance + recurringBalance + repeatBalance
}

func (c *FindCategoriesRepository) calculateSubCategoryBalance(subCategoryId primitive.ObjectID, globalFilters *presentationHelpers.GlobalFilterParams, frequency string) float64 {
	collection := c.Db.Collection("transaction")
	startOfMonth := time.Date(globalFilters.Year, time.Month(globalFilters.Month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	filter := bson.M{
		"workspace_id": globalFilters.WorkspaceId,

		"$and": []bson.M{
			{"$or": []bson.M{
				{"sub_category_id": subCategoryId},
				{"tags.sub_tag_id": subCategoryId},
			}},
			{"$or": []bson.M{
				{"$and": []bson.M{
					{
						"due_date": bson.M{
							"$lt": endOfMonth,
						},
						"is_confirmed": false,
					},
				}},
				{
					"$and": []bson.M{
						{
							"confirmation_date": bson.M{
								"$lt": endOfMonth,
							},
							"is_confirmed": true,
						},
					},
				},
			}},
		},
		"frequency": frequency,
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return 0.0
	}

	var transactions []models.Transaction
	if err := cursor.All(context.Background(), &transactions); err != nil {
		return 0.0
	}

	switch frequency {
	case "DO_NOT_REPEAT":
		return helpers.CalculateTransactionBalance(transactions)
	case "RECURRING":
		return helpers.CalculateRecurringTransactionsBalance(transactions, globalFilters.Year, globalFilters.Month, c.Db)
	case "REPEAT":
		return helpers.CalculateRepeatTransactionsBalance(transactions, globalFilters.Year, globalFilters.Month, c.Db)
	}
	return 0.0
}
