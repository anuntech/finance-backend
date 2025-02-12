package recipe_repository

import (
	"context"
	"time"

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

func (r *UpdateRecipeRepository) CreateSubCategory(subCategory models.SubRecipeCategory, recipeId primitive.ObjectID, workspaceId primitive.ObjectID) (*models.SubRecipeCategory, error) {
	collection := r.Db.Collection("recipe")

	subCategory.Id = primitive.NewObjectID()

	update := bson.M{
		"$push": bson.M{
			"sub_categories": subCategory,
		},
	}

	_, err := collection.UpdateOne(context.Background(), bson.M{"_id": recipeId, "workspace_id": workspaceId}, update)
	if err != nil {
		return nil, err
	}

	return &subCategory, nil
}

func (r *UpdateRecipeRepository) DeleteSubCategory(recipeId primitive.ObjectID, subCategoryId primitive.ObjectID, workspaceId primitive.ObjectID) error {
	collection := r.Db.Collection("recipe")

	filter := bson.M{
		"_id":            recipeId,
		"workspace_id":   workspaceId,
		"sub_categories": bson.M{"$elemMatch": bson.M{"id": subCategoryId}},
	}

	update := bson.M{
		"$pull": bson.M{
			"sub_categories": bson.M{"id": subCategoryId},
		},
	}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}

func (r *UpdateRecipeRepository) UpdateRecipe(recipe models.Recipe) error {
	collection := r.Db.Collection("recipe")

	filter := bson.M{"_id": recipe.Id, "workspace_id": recipe.WorkspaceId}
	update := bson.M{
		"$set": bson.M{
			"name":           recipe.Name,
			"sub_categories": recipe.SubCategories,
			"updated_at":     time.Now(),
		},
	}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}
