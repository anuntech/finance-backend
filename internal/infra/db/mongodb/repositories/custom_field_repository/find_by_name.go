package custom_field_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindCustomFieldByNameRepository struct {
	db *mongo.Database
}

func NewFindCustomFieldByNameRepository(db *mongo.Database) *FindCustomFieldByNameRepository {
	return &FindCustomFieldByNameRepository{db}
}

func (r *FindCustomFieldByNameRepository) FindByNameAndWorkspaceId(name string, workspaceId primitive.ObjectID) (*models.CustomField, error) {
	collection := r.db.Collection("custom_field")

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	// Buscar campo personalizado por nome e workspace

	var customField models.CustomField
	err := collection.FindOne(ctx, bson.M{
		"name":         name,
		"workspace_id": workspaceId,
	}).Decode(&customField)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &customField, nil
}
