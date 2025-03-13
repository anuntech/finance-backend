package account_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
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
	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	// Remove these accounts from all transactions
	transactionCollection := d.Db.Collection("transaction")
	_, err = transactionCollection.UpdateMany(
		ctx,
		bson.M{"account_id": bson.M{"$in": accountIds}, "workspace_id": workspaceId},
		bson.M{"$unset": bson.M{"account_id": ""}},
	)

	return err
}
