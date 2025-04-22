package credit_card_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UpdateCreditCardRepository handles updating credit cards
type UpdateCreditCardRepository struct {
	Db *mongo.Database
}

// NewUpdateCreditCardRepository creates a new UpdateCreditCardRepository
func NewUpdateCreditCardRepository(db *mongo.Database) *UpdateCreditCardRepository {
	return &UpdateCreditCardRepository{Db: db}
}

// Update modifies an existing credit card
func (r *UpdateCreditCardRepository) Update(creditCardId primitive.ObjectID, creditCard *models.CreditCard) (*models.CreditCard, error) {
	collection := r.Db.Collection("credit_card")

	workspaceId, err := primitive.ObjectIDFromHex(creditCard.WorkspaceId)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": creditCardId, "workspace_id": workspaceId}
	update := bson.M{"$set": bson.M{
		"name":       creditCard.Name,
		"due_date":   creditCard.DueDate,
		"close_date": creditCard.CloseDate,
		"limit":      creditCard.Limit,
		"updated_at": time.Now().UTC(),
	}}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	// Retrieve updated document
	var updated models.CreditCard
	err = collection.FindOne(ctx, filter).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}
