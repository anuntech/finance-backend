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

type accountToSaveInterface struct {
	Id          primitive.ObjectID `bson:"_id"`
	Name        string             `bson:"name"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
	WorkspaceId primitive.ObjectID `bson:"workspace_id"`
	BankId      primitive.ObjectID `bson:"bank_id"`
}

func (c *CreateAccountMongoRepository) Create(account *models.AccountInput) (*models.Account, error) {
	collection := c.Db.Collection("account")

	accountToSave := accountToSaveInterface{
		Id:          primitive.NewObjectID(),
		Name:        account.Name,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		WorkspaceId: account.WorkspaceId,
		BankId:      account.BankId,
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
	}, nil
}
