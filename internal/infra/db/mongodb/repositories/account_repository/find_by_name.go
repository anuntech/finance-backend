package account_repository

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

func (f *FindByNameMongoRepository) FindByNameAndWorkspaceId(name string, workspaceId primitive.ObjectID) (*models.Account, error) {
	collection := f.Db.Collection("account")

	filter := bson.M{"name": name, "workspace_id": workspaceId}
	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	cursor := collection.FindOne(ctx, filter)
	if cursor.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if cursor.Err() != nil {
		return nil, cursor.Err()
	}

	var account models.Account
	if err := cursor.Decode(&account); err != nil {
		return nil, err
	}

	return &account, nil
}
