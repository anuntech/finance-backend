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

func (d *DeleteAccountMongoRepository) Delete(id string) error {
	collection := d.Db.Collection("account")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectId}
	_, err = collection.DeleteOne(context.Background(), filter)
	return err
}
