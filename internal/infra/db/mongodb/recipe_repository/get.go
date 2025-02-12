package recipe_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindRecipesByWorkspaceIdRepository struct {
	Db *mongo.Database
}

func NewFindRecipesByWorkspaceIdRepository(db *mongo.Database) *FindRecipesByWorkspaceIdRepository {
	return &FindRecipesByWorkspaceIdRepository{
		Db: db,
	}
}

func (r *FindRecipesByWorkspaceIdRepository) Find(workspaceId primitive.ObjectID) ([]models.Recipe, error) {
	collection := r.Db.Collection("recipe")

	cursor, err := collection.Find(context.Background(), bson.M{"workspaceId": workspaceId})
	if err != nil {
		return nil, err
	}

	var recipes []models.Recipe
	if err = cursor.All(context.Background(), &recipes); err != nil {
		return nil, err
	}

	return recipes, nil
}
