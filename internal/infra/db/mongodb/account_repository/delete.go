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

func (d *DeleteAccountMongoRepository) Delete(id primitive.ObjectID) error {
	collection := d.Db.Collection("account")

	filter := bson.M{"_id": id}
	_, err := collection.DeleteOne(context.Background(), filter)
	return err
}
