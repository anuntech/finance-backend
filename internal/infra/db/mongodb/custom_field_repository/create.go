package custom_field_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateCustomFieldRepository struct {
	Db *mongo.Database
}

func NewCreateCustomFieldRepository(db *mongo.Database) *CreateCustomFieldRepository {
	return &CreateCustomFieldRepository{
		Db: db,
	}
}

func (r *CreateCustomFieldRepository) Create(customField *models.CustomField) (*models.CustomField, error) {
	collection := r.Db.Collection("custom_field")

	// Convert string IDs to ObjectIDs
	workspaceId, err := primitive.ObjectIDFromHex(customField.WorkspaceId)
	if err != nil {
		return nil, err
	}

	// Create a new ID for the custom field
	id := primitive.NewObjectID()

	// Prepare the document for insertion
	document := map[string]interface{}{
		"_id":          id.Hex(),
		"workspace_id": workspaceId,
		"name":         customField.Name,
		"type":         customField.Type,
		"options":      customField.Options,
		"required":     customField.Required,
		"created_at":   time.Now().UTC(),
		"updated_at":   time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()
	_, err = collection.InsertOne(ctx, document)
	if err != nil {
		return nil, err
	}

	// Update the ID in the returned model
	customField.Id = id.Hex()

	return customField, nil
}
