package account_repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DeleteAccountMongoRepository struct {
	Db *mongo.Database
}

func NewDeleteAccountMongoRepository(db *mongo.Database) *DeleteAccountMongoRepository {
	return &DeleteAccountMongoRepository{
		Db: db,
	}
}

func (d *DeleteAccountMongoRepository) Delete(accountIds []primitive.ObjectID, workspaceId primitive.ObjectID) error {
	collection := d.Db.Collection("account")

	filter := bson.M{"_id": bson.M{"$in": accountIds}, "workspace_id": workspaceId}
	_, err := collection.DeleteMany(context.Background(), filter)
	return err
}
