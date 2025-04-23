package credit_card_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	presentationHelpers "github.com/anuntech/finance-backend/internal/presentation/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindCreditCardsRepository handles fetching credit cards
type FindCreditCardsRepository struct {
	Db *mongo.Database
}

// NewFindCreditCardsRepository creates a new FindCreditCardsRepository
func NewFindCreditCardsRepository(db *mongo.Database) *FindCreditCardsRepository {
	return &FindCreditCardsRepository{Db: db}
}

// Find retrieves credit cards by global filters
func (r *FindCreditCardsRepository) Find(globalFilters *presentationHelpers.GlobalFilterParams) ([]models.CreditCard, error) {
	collection := r.Db.Collection("credit_card")

	filter := bson.M{"workspace_id": globalFilters.WorkspaceId}

	opts := options.Find().SetSort(bson.M{"name": 1})

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var creditCards []models.CreditCard
	if err := cursor.All(ctx, &creditCards); err != nil {
		return nil, err
	}

	return creditCards, nil
}
