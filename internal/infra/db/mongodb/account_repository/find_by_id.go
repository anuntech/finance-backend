package account_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
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

func (f *FindByIdMongoRepository) Find(id string) (*models.Account, error) {
	collection := f.Db.Collection("account")

	filter := bson.M{"_id": id}
	cursor := collection.FindOne(context.Background(), filter)
	if cursor.Err() != nil {
		return nil, cursor.Err()
	}

	var account models.Account
	if err := cursor.Decode(&account); err != nil {
		return nil, err
	}

	return &account, nil
}
