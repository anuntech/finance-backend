package category_repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DeleteCategoryRepository struct {
	Db *mongo.Database
}

func NewDeleteCategoryRepository(db *mongo.Database) *DeleteCategoryRepository {
	return &DeleteCategoryRepository{
		Db: db,
	}
}

func (r *DeleteCategoryRepository) Delete(categoryId primitive.ObjectID, workspaceId primitive.ObjectID) error {
	collection := r.Db.Collection("category")

	_, err := collection.DeleteOne(context.Background(), bson.M{"_id": categoryId, "workspace_id": workspaceId})
	if err != nil {
		return err
	}

	return nil
}
