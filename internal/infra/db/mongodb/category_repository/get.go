package category_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindCategorysByWorkspaceIdRepository struct {
	Db *mongo.Database
}

func NewFindCategorysByWorkspaceIdRepository(db *mongo.Database) *FindCategorysByWorkspaceIdRepository {
	return &FindCategorysByWorkspaceIdRepository{
		Db: db,
	}
}

func (r *FindCategorysByWorkspaceIdRepository) Find(workspaceId primitive.ObjectID, categoryType string) ([]models.Category, error) {
	collection := r.Db.Collection("category")

	filter := bson.M{"workspace_id": workspaceId}
	if categoryType != "" {
		filter["type"] = categoryType
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
