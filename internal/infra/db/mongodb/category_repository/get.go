package category_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	presentationHelpers "github.com/anuntech/finance-backend/internal/presentation/helpers"
	"go.mongodb.org/mongo-driver/bson"
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

	return categories, nil
}
