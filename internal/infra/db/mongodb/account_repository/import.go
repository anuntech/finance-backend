package account_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ImportAccountsMongoRepository struct {
	db *mongo.Database
}

func NewImportAccountsMongoRepository(db *mongo.Database) *ImportAccountsMongoRepository {
	return &ImportAccountsMongoRepository{
		db: db,
	}
}

func (r *ImportAccountsMongoRepository) Import(accounts []models.AccountInput, workspaceId primitive.ObjectID) ([]models.Account, error) {
	collection := r.db.Collection("account")

	var docs []interface{}
	for i := range accounts {
		account := models.Account{
			Id:          primitive.NewObjectID(),
			Name:        accounts[i].Name,
			Balance:     accounts[i].Balance,
			BankId:      accounts[i].BankId,
			WorkspaceId: workspaceId,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		docs = append(docs, account)
	}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	_, err := collection.InsertMany(ctx, docs)
	if err != nil {
		return nil, err
	}

	var importedAccounts []models.Account
	for _, doc := range docs {
		importedAccounts = append(importedAccounts, doc.(models.Account))
	}

	return importedAccounts, nil
}
