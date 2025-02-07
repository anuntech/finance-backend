package recipe_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateRecipeRepository struct {
	Db *mongo.Database
}

func NewCreateRecipeRepository(db *mongo.Database) *CreateRecipeRepository {
	return &CreateRecipeRepository{
		Db: db,
	}
}

func (r *CreateRecipeRepository) Create(recipe models.Recipe) (*models.Recipe, error) {
	collection := r.Db.Collection("recipe")

	recipeToSave := &models.Recipe{
		Id:            primitive.NewObjectID(),
		Name:          recipe.Name,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		WorkspaceId:   recipe.WorkspaceId,
		AccountId:     recipe.AccountId,
		SubCategories: recipe.SubCategories,
	}
	_, err := collection.InsertOne(context.Background(), recipeToSave)
	if err != nil {
		return nil, err
	}

	return recipeToSave, nil
}
