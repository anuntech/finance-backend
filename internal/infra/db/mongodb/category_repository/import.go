package category_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ImportCategoriesRepository struct {
	db *mongo.Database
}

func NewImportCategoriesRepository(db *mongo.Database) *ImportCategoriesRepository {
	return &ImportCategoriesRepository{
		db: db,
	}
}

func (r *ImportCategoriesRepository) Import(categories []models.Category, workspaceId primitive.ObjectID) ([]models.Category, error) {
	ctx := context.Background()

	collection := r.db.Collection("categories")

	var docs []interface{}
	for i := range categories {
		categories[i].Id = primitive.NewObjectID()
		categories[i].WorkspaceId = workspaceId
		categories[i].CreatedAt = time.Now()
		categories[i].UpdatedAt = time.Now()
		docs = append(docs, categories[i])
	}

	_, err := collection.InsertMany(ctx, docs)
	if err != nil {
		return nil, err
	}

	return categories, nil
}
