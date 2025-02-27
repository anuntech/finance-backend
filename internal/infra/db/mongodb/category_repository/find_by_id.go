package category_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindCategoryByIdRepository struct {
	Db *mongo.Database
}

func NewFindCategoryByIdRepository(db *mongo.Database) *FindCategoryByIdRepository {
	return &FindCategoryByIdRepository{
		Db: db,
	}
}

func (r *FindCategoryByIdRepository) Find(categoryId primitive.ObjectID, workspaceId primitive.ObjectID) (*models.Category, error) {
	collection := r.Db.Collection("category")

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	var category models.Category
	err := collection.FindOne(ctx, bson.M{"_id": categoryId, "workspace_id": workspaceId}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &category, nil
}
