package account_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	presentationHelpers "github.com/anuntech/finance-backend/internal/presentation/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindManyByUserIdAndWorkspaceIdMongoRepository struct {
	Db *mongo.Database
}

func NewFindManyByUserIdAndWorkspaceIdMongoRepository(db *mongo.Database) *FindManyByUserIdAndWorkspaceIdMongoRepository {
	return &FindManyByUserIdAndWorkspaceIdMongoRepository{
		Db: db,
	}
}

func (f *FindManyByUserIdAndWorkspaceIdMongoRepository) Find(globalFilters *presentationHelpers.GlobalFilterParams) ([]models.Account, error) {
	collection := f.Db.Collection("account")

	filter := bson.M{"workspace_id": globalFilters.WorkspaceId}
	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var accounts []models.Account
	if err = cursor.All(context.Background(), &accounts); err != nil {
		return nil, err
	}

	return accounts, nil
}
