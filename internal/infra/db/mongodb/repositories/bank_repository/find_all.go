package bank_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindAllMongoRepository struct {
	Db *mongo.Database
}

func NewFindAllMongoRepository(db *mongo.Database) *FindAllMongoRepository {
	return &FindAllMongoRepository{
		Db: db,
	}
}

func (r *FindAllMongoRepository) Find() ([]models.Bank, error) {
	collection := r.Db.Collection("bank")

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var banks []models.Bank
	if err = cursor.All(ctx, &banks); err != nil {
		return nil, err
	}

	return banks, nil
}
