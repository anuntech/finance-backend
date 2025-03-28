package category_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindByNameMongoRepository struct {
	Db *mongo.Database
}

func NewFindByNameMongoRepository(db *mongo.Database) *FindByNameMongoRepository {
	return &FindByNameMongoRepository{
		Db: db,
	}
}

func (r *FindByNameMongoRepository) FindByNameAndWorkspaceId(name string, workspaceId primitive.ObjectID) (*models.Category, error) {
	collection := r.Db.Collection("category")

	filter := bson.M{"name": name, "workspace_id": workspaceId}
	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	var category models.Category
	err := collection.FindOne(ctx, filter).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &category, nil
}
