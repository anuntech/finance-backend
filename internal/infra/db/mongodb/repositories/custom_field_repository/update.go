package custom_field_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UpdateCustomFieldRepository struct {
	Db *mongo.Database
}

func NewUpdateCustomFieldRepository(db *mongo.Database) *UpdateCustomFieldRepository {
	return &UpdateCustomFieldRepository{
		Db: db,
	}
}

func (r *UpdateCustomFieldRepository) Update(customFieldId primitive.ObjectID, customField *models.CustomField) (*models.CustomField, error) {
	collection := r.Db.Collection("custom_field")

	workspaceId, err := primitive.ObjectIDFromHex(customField.WorkspaceId)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": customFieldId, "workspace_id": workspaceId}

	// Prepare the document for update
	update := bson.M{
		"$set": bson.M{
			"name":       customField.Name,
			"type":       customField.Type,
			"options":    customField.Options,
			"required":   customField.Required,
			"updated_at": time.Now().UTC(),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	// Get the updated document
	var updatedCustomField models.CustomField
	err = collection.FindOne(ctx, filter).Decode(&updatedCustomField)
	if err != nil {
		return nil, err
	}

	return &updatedCustomField, nil
}
