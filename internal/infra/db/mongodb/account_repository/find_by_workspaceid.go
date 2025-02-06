package account_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (f *FindManyByUserIdAndWorkspaceIdMongoRepository) Find(userId string, workspaceId string) ([]models.Account, error) {
	collection := f.Db.Collection("account")

	workspaceIdObjectId, err := primitive.ObjectIDFromHex(workspaceId)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"workspace_id": workspaceIdObjectId}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	var accounts []models.Account
	if err = cursor.All(context.Background(), &accounts); err != nil {
		return nil, err
	}

	return accounts, nil
}
