package custom_field_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	presentationHelpers "github.com/anuntech/finance-backend/internal/presentation/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FindCustomFieldsRepository struct {
	Db *mongo.Database
}

func NewFindCustomFieldsRepository(db *mongo.Database) *FindCustomFieldsRepository {
	return &FindCustomFieldsRepository{
		Db: db,
	}
}

func (r *FindCustomFieldsRepository) Find(globalFilters *presentationHelpers.GlobalFilterParams) ([]models.CustomField, error) {
	collection := r.Db.Collection("custom_field")

	filter := bson.M{"workspace_id": globalFilters.WorkspaceId}
	if globalFilters.Type != "" {
		filter["transaction_type"] = globalFilters.Type
	}

	// Define sorting options to get in order
	opts := options.Find().SetSort(bson.M{"name": 1})

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var customFields []models.CustomField
	if err := cursor.All(ctx, &customFields); err != nil {
		return nil, err
	}

	return customFields, nil
}
