package custom_field_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DeleteCustomFieldRepository struct {
	Db *mongo.Database
}

func NewDeleteCustomFieldRepository(db *mongo.Database) *DeleteCustomFieldRepository {
	return &DeleteCustomFieldRepository{
		Db: db,
	}
}

func (r *DeleteCustomFieldRepository) Delete(customFieldIds []primitive.ObjectID, workspaceId primitive.ObjectID) error {
	collection := r.Db.Collection("custom_field")

	filter := bson.M{
		"_id":          bson.M{"$in": customFieldIds},
		"workspace_id": workspaceId,
	}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err := collection.DeleteMany(ctx, filter)
	return err
}
