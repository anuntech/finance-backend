package category_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UpdateCategoryRepository struct {
	Db *mongo.Database
}

func NewUpdateCategoryRepository(db *mongo.Database) *UpdateCategoryRepository {
	return &UpdateCategoryRepository{
		Db: db,
	}
}

func (r *UpdateCategoryRepository) CreateSubCategory(subCategory *models.SubCategoryCategory, categoryId primitive.ObjectID, workspaceId primitive.ObjectID) (*models.SubCategoryCategory, error) {
	collection := r.Db.Collection("category")

	subCategory.Id = primitive.NewObjectID()

	update := bson.M{
		"$push": bson.M{
			"sub_categories": subCategory,
		},
	}

	_, err := collection.UpdateOne(context.Background(), bson.M{"_id": categoryId, "workspace_id": workspaceId}, update)
	if err != nil {
		return nil, err
	}

	return subCategory, nil
}

func (r *UpdateCategoryRepository) DeleteSubCategory(categoryId primitive.ObjectID, subCategoryId primitive.ObjectID, workspaceId primitive.ObjectID) error {
	collection := r.Db.Collection("category")

	filter := bson.M{
		"_id":            categoryId,
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

func (r *UpdateCategoryRepository) UpdateCategory(category *models.Category) error {
	collection := r.Db.Collection("category")

	filter := bson.M{"_id": category.Id, "workspace_id": category.WorkspaceId}
	update := bson.M{
		"$set": bson.M{
			"name":       category.Name,
			"icon":       category.Icon,
			"updated_at": time.Now(),
			"type":       category.Type,
		},
	}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}
