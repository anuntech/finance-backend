package category_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
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

func (r *DeleteCategoryRepository) Delete(categoryIds []primitive.ObjectID, workspaceId primitive.ObjectID) error {
	collection := r.Db.Collection("category")

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err := collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": categoryIds}, "workspace_id": workspaceId})
	if err != nil {
		return err
	}

	transactionCollection := r.Db.Collection("transaction")
	_, err = transactionCollection.UpdateMany(
		ctx,
		bson.M{"category_id": bson.M{"$in": categoryIds}, "workspace_id": workspaceId},
		bson.M{"$unset": bson.M{"category_id": ""}},
	)
	if err != nil {
		return err
	}

	return nil
}
