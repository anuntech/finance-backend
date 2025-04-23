package credit_card_repository

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// FindByNameMongoRepository handles fetching a credit card by name within a workspace
type FindByNameMongoRepository struct {
	Db *mongo.Database
}

// NewFindByNameMongoRepository creates a new FindByNameMongoRepository
func NewFindByNameMongoRepository(db *mongo.Database) *FindByNameMongoRepository {
	return &FindByNameMongoRepository{Db: db}
}

// FindByNameAndWorkspaceId returns a credit card matching the name and workspace
func (r *FindByNameMongoRepository) FindByNameAndWorkspaceId(name string, workspaceId primitive.ObjectID) (*models.CreditCard, error) {
	collection := r.Db.Collection("credit_card")

	filter := bson.M{"name": name, "workspace_id": workspaceId}
	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	var card models.CreditCard
	err := collection.FindOne(ctx, filter).Decode(&card)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &card, nil
}
