package category_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
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
	collection := r.db.Collection("category")

	var docs []interface{}
	for i := range categories {
		categories[i].Id = primitive.NewObjectID()
		categories[i].WorkspaceId = workspaceId
		categories[i].CreatedAt = time.Now().UTC()
		categories[i].UpdatedAt = time.Now().UTC()
		docs = append(docs, categories[i])
	}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err := collection.InsertMany(ctx, docs)
	if err != nil {
		return nil, err
	}

	return categories, nil
}
