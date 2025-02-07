package recipe_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UpdateRecipeRepository struct {
	Db *mongo.Database
}

func NewUpdateRecipeRepository(db *mongo.Database) *UpdateRecipeRepository {
	return &UpdateRecipeRepository{
		Db: db,
	}
}

func (r *UpdateRecipeRepository) CreateSubCategory(subCategory models.SubRecipeCategory, recipeId string, workspaceId string) error {
	collection := r.Db.Collection("recipe")

	recipeObjectId, err := primitive.ObjectIDFromHex(recipeId)
	if err != nil {
		return err
	}

	workspaceObjectId, err := primitive.ObjectIDFromHex(workspaceId)
	if err != nil {
		return err
	}

	update := bson.M{
		"$push": bson.M{
			"subCategories": subCategory,
		},
	}

	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": recipeObjectId, "workspaceId": workspaceObjectId}, update)
	return err
}
