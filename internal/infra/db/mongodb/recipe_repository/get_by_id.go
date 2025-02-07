package recipe_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindRecipeByIdRepository struct {
	Db *mongo.Database
}

func NewFindRecipeByIdRepository(db *mongo.Database) *FindRecipeByIdRepository {
	return &FindRecipeByIdRepository{
		Db: db,
	}
}

func (r *FindRecipeByIdRepository) Find(recipeId string, workspaceId string) (*models.Recipe, error) {
	collection := r.Db.Collection("recipe")

	recipeObjectId, err := primitive.ObjectIDFromHex(recipeId)
	if err != nil {
		return nil, err
	}

	workspaceObjectId, err := primitive.ObjectIDFromHex(workspaceId)
	if err != nil {
		return nil, err
	}

	var recipe models.Recipe
	err = collection.FindOne(context.Background(), bson.M{"_id": recipeObjectId, "workspaceId": workspaceObjectId}).Decode(&recipe)
	if err != nil {
		return nil, err
	}

	return &recipe, nil
}
