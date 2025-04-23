package credit_card_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateCreditCardRepository struct {
	Db *mongo.Database
}

func NewCreateCreditCardRepository(db *mongo.Database) *CreateCreditCardRepository {
	return &CreateCreditCardRepository{Db: db}
}

func (r *CreateCreditCardRepository) Create(creditCard *models.CreditCard) (*models.CreditCard, error) {
	collection := r.Db.Collection("credit_card")

	workspaceId, err := primitive.ObjectIDFromHex(creditCard.WorkspaceId)
	if err != nil {
		return nil, err
	}

	id := primitive.NewObjectID()
	doc := map[string]interface{}{
		"_id":          id,
		"workspace_id": workspaceId,
		"name":         creditCard.Name,
		"due_date":     creditCard.DueDate,
		"close_date":   creditCard.CloseDate,
		"limit":        creditCard.Limit,
		"balance":      creditCard.Balance,
	}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err = collection.InsertOne(ctx, doc)
	if err != nil {
		return nil, err
	}

	creditCard.Id = id.Hex()
	return creditCard, nil
}
