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

func (r *DeleteRecipeRepository) Delete(recipeId string, workspaceId string) error {
	collection := r.Db.Collection("recipe")

	recipeIdObjectId, err := primitive.ObjectIDFromHex(recipeId)
	if err != nil {
		return err
	}

	workspaceIdObjectId, err := primitive.ObjectIDFromHex(workspaceId)
	if err != nil {
		return err
	}

	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": recipeIdObjectId, "workspaceId": workspaceIdObjectId})
	if err != nil {
		return err
	}

	return nil
}
