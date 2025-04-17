package custom_field_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindCustomFieldByIdRepository struct {
	Db *mongo.Database
}

func NewFindCustomFieldByIdRepository(db *mongo.Database) *FindCustomFieldByIdRepository {
	return &FindCustomFieldByIdRepository{
		Db: db,
	}
}

func (r *FindCustomFieldByIdRepository) Find(customFieldId primitive.ObjectID, workspaceId primitive.ObjectID) (*models.CustomField, error) {
	collection := r.Db.Collection("custom_field")

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	var customField models.CustomField
	err := collection.FindOne(ctx, bson.M{"_id": customFieldId, "workspace_id": workspaceId}).Decode(&customField)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &customField, nil
}
