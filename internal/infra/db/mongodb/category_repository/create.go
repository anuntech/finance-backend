package category_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateCategoryRepository struct {
	Db *mongo.Database
}

func NewCreateCategoryRepository(db *mongo.Database) *CreateCategoryRepository {
	return &CreateCategoryRepository{
		Db: db,
	}
}

func (r *CreateCategoryRepository) Create(category *models.Category) (*models.Category, error) {
	collection := r.Db.Collection("category")

	categoryToSave := &models.Category{
		Id:            primitive.NewObjectID(),
		Name:          category.Name,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		WorkspaceId:   category.WorkspaceId,
		SubCategories: category.SubCategories,
		Icon:          category.Icon,
		Type:          category.Type,
	}

	_, err := collection.InsertOne(context.Background(), categoryToSave)
	if err != nil {
		return nil, err
	}

	return categoryToSave, nil
}
