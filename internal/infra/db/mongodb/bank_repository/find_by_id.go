package bank_repository

import (
	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
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

func (r *FindByIdMongoRepository) Find(id string) (*models.Bank, error) {
	collection := r.Db.Collection("bank")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": objectId}
	cursor := collection.FindOne(helpers.Ctx, filter)
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
