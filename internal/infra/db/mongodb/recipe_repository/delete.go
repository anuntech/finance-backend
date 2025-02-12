package recipe_repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DeleteRecipeRepository struct {
	Db *mongo.Database
}

func NewDeleteRecipeRepository(db *mongo.Database) *DeleteRecipeRepository {
	return &DeleteRecipeRepository{
		Db: db,
	}
}

func (r *DeleteRecipeRepository) Delete(recipeId primitive.ObjectID, workspaceId primitive.ObjectID) error {
	collection := r.Db.Collection("recipe")

	_, err := collection.DeleteOne(context.Background(), bson.M{"_id": recipeId, "workspaceId": workspaceId})
	if err != nil {
		return err
	}

	return nil
}
