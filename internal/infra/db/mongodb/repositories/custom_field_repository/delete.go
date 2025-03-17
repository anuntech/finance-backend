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
	if err != nil {
		return err
	}

	transactionCollection := r.Db.Collection("transaction")

	update := bson.M{
		"$pull": bson.M{
			"custom_fields": bson.M{
				"custom_field_id": bson.M{"$in": customFieldIds},
			},
		},
	}

	transactionFilter := bson.M{
		"workspace_id": workspaceId,
		"custom_fields": bson.M{
			"$elemMatch": bson.M{
				"custom_field_id": bson.M{"$in": customFieldIds},
			},
		},
	}

	updateCtx, updateCancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer updateCancel()

	_, err = transactionCollection.UpdateMany(updateCtx, transactionFilter, update)
	return err
}
