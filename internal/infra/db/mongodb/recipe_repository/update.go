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

func (r *UpdateRecipeRepository) CreateSubCategory(subCategory models.SubRecipeCategory, recipeId string, workspaceId string) (*models.SubRecipeCategory, error) {
	collection := r.Db.Collection("recipe")

	recipeObjectId, err := primitive.ObjectIDFromHex(recipeId)
	if err != nil {
		return nil, err
	}

	workspaceObjectId, err := primitive.ObjectIDFromHex(workspaceId)
	if err != nil {
		return nil, err
	}

	subCategory.Id = primitive.NewObjectID()
	update := bson.M{
		"$push": bson.M{
			"subCategories": subCategory,
		},
	}

	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": recipeObjectId, "workspaceId": workspaceObjectId}, update)
	if err != nil {
		return nil, err
	}

	return &subCategory, nil
}

func (r *UpdateRecipeRepository) DeleteSubCategory(recipeId string, subCategoryId string, workspaceId string) error {
	collection := r.Db.Collection("recipe")

	recipeObjectId, err := primitive.ObjectIDFromHex(recipeId)
	if err != nil {
		return err
	}

	subCategoryObjectId, err := primitive.ObjectIDFromHex(subCategoryId)
	if err != nil {
		return err
	}

	workspaceObjectId, err := primitive.ObjectIDFromHex(workspaceId)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id":           recipeObjectId,
		"workspaceId":   workspaceObjectId,
		"subCategories": bson.M{"$elemMatch": bson.M{"id": subCategoryObjectId}},
	}

	update := bson.M{
		"$pull": bson.M{
			"subCategories": bson.M{"id": subCategoryObjectId},
		},
	}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	return err
}
