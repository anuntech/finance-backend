package category_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
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

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err := collection.UpdateOne(ctx, bson.M{"_id": categoryId, "workspace_id": workspaceId}, update)
	if err != nil {
		return nil, err
	}

	return subCategory, nil
}

func (r *UpdateCategoryRepository) DeleteSubCategory(subCategoryIds []primitive.ObjectID, categoryId primitive.ObjectID, workspaceId primitive.ObjectID) error {
	collection := r.Db.Collection("category")

	filter := bson.M{
		"_id":            categoryId,
		"workspace_id":   workspaceId,
		"sub_categories": bson.M{"$elemMatch": bson.M{"_id": bson.M{"$in": subCategoryIds}}},
	}

	update := bson.M{
		"$pull": bson.M{
			"sub_categories": bson.M{"_id": bson.M{"$in": subCategoryIds}},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err := collection.UpdateOne(ctx, filter, update)
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

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *UpdateCategoryRepository) UpdateSubCategory(subCategory *models.SubCategoryCategory, categoryId primitive.ObjectID, subCategoryId primitive.ObjectID, workspaceId primitive.ObjectID) error {
	collection := r.Db.Collection("category")

	filter := bson.M{
		"_id":               categoryId,
		"workspace_id":      workspaceId,
		"sub_categories.id": subCategoryId,
	}

	update := bson.M{
		"$set": bson.M{
			"sub_categories.$.name": subCategory.Name,
			"sub_categories.$.icon": subCategory.Icon,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *UpdateCategoryRepository) CreateSubCategories(subCategories []models.SubCategoryCategory, categoryId primitive.ObjectID, workspaceId primitive.ObjectID) ([]models.SubCategoryCategory, error) {
	collection := r.Db.Collection("category")

	// Gera IDs para as novas subcategorias
	for i := range subCategories {
		subCategories[i].Id = primitive.NewObjectID()
	}

	update := bson.M{
		"$push": bson.M{
			"sub_categories": bson.M{
				"$each": subCategories,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err := collection.UpdateOne(ctx,
		bson.M{"_id": categoryId, "workspace_id": workspaceId},
		update)
	if err != nil {
		return nil, err
	}

	return subCategories, nil
}
