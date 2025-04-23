package credit_card_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// DeleteCreditCardRepository handles deleting credit cards
type DeleteCreditCardRepository struct {
	Db *mongo.Database
}

// NewDeleteCreditCardRepository creates a new DeleteCreditCardRepository
func NewDeleteCreditCardRepository(db *mongo.Database) *DeleteCreditCardRepository {
	return &DeleteCreditCardRepository{Db: db}
}

// Delete removes credit cards matching the given IDs and workspace
func (r *DeleteCreditCardRepository) Delete(creditCardIds []primitive.ObjectID, workspaceId primitive.ObjectID) error {
	collection := r.Db.Collection("credit_card")
	filter := bson.M{"_id": bson.M{"$in": creditCardIds}, "workspace_id": workspaceId}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err := collection.DeleteMany(ctx, filter)
	return err
}
