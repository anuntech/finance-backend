package credit_card_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// FindCreditCardByIdRepository handles fetching a credit card by its ID
type FindCreditCardByIdRepository struct {
	Db *mongo.Database
}

// NewFindCreditCardByIdRepository creates a new FindCreditCardByIdRepository
func NewFindCreditCardByIdRepository(db *mongo.Database) *FindCreditCardByIdRepository {
	return &FindCreditCardByIdRepository{Db: db}
}

// Find returns a credit card by its ID and workspace
func (r *FindCreditCardByIdRepository) Find(creditCardId primitive.ObjectID, workspaceId primitive.ObjectID) (*models.CreditCard, error) {
	collection := r.Db.Collection("credit_card")

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	var creditCard models.CreditCard
	err := collection.FindOne(ctx, bson.M{"_id": creditCardId, "workspace_id": workspaceId}).Decode(&creditCard)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &creditCard, nil
}
