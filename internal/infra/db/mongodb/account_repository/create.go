package account_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateAccountMongoRepository struct {
	Db *mongo.Database
}

func NewCreateAccountMongoRepository(db *mongo.Database) *CreateAccountMongoRepository {
	return &CreateAccountMongoRepository{
		Db: db,
	}
}

func (c *CreateAccountMongoRepository) Create(account *models.AccountInput) (*models.Account, error) {
	collection := c.Db.Collection("account")

	accountToSave := models.Account{
		Id:          primitive.NewObjectID(),
		Name:        account.Name,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		WorkspaceId: account.WorkspaceId,
		BankId:      account.BankId,
		Balance:     account.Balance,
	}

	_, err := collection.InsertOne(context.Background(), accountToSave)
	if err != nil {
		return nil, err
	}

	return &models.Account{
		Id:          accountToSave.Id,
		Name:        accountToSave.Name,
		CreatedAt:   accountToSave.CreatedAt,
		UpdatedAt:   accountToSave.UpdatedAt,
		WorkspaceId: accountToSave.WorkspaceId,
		BankId:      accountToSave.BankId,
		Balance:     accountToSave.Balance,
	}, nil
}
