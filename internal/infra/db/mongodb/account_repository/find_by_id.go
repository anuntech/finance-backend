package account_repository

import (
	"context"
	"log"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindByIdMongoRepository struct {
	Db *mongo.Database
}

func NewFindByIdMongoRepository(db *mongo.Database) *FindByIdMongoRepository {
	return &FindByIdMongoRepository{
		Db: db,
	}
}

func (f *FindByIdMongoRepository) Find(id string, workspaceId string) (*models.Account, error) {
	collection := f.Db.Collection("account")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	log.Printf("ObjectID: %s", objectId.Hex())

	workspaceObjectId, err := primitive.ObjectIDFromHex(workspaceId)
	if err != nil {
		return nil, err
	}

	log.Printf("WorkspaceObjectID: %s", workspaceObjectId.Hex())

	filter := bson.M{"_id": objectId, "workspace_id": workspaceObjectId}
	cursor := collection.FindOne(context.Background(), filter)
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
