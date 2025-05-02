package bank_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
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

func (r *FindByNameMongoRepository) FindByName(name string) (*models.Bank, error) {
	collection := r.Db.Collection("bank")

	filter := bson.M{"name": name}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	cursor := collection.FindOne(ctx, filter)
	if cursor.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if cursor.Err() != nil {
		return nil, cursor.Err()
	}

	var bank models.Bank
	if err := cursor.Decode(&bank); err != nil {
		return nil, err
	}

	return &bank, nil
}
