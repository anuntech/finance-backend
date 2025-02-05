package account_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/google/uuid"
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
	Id          string    `bson:"_id"`
	Name        string    `bson:"name"`
	Image       string    `bson:"image"`
	Color       string    `bson:"color"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
	WorkspaceId string    `bson:"workspace_id"`
	UserId      string    `bson:"user_id"`
}

func (c *CreateAccountMongoRepository) Create(account *models.AccountInput) (*models.Account, error) {
	collection := c.Db.Collection("account")

	accountToSave := accountToSaveInterface{
		Id:          uuid.New().String(),
		Name:        account.Name,
		Image:       account.Image,
		Color:       account.Color,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		WorkspaceId: account.WorkspaceId,
		UserId:      account.UserId,
	}
	_, err := collection.InsertOne(context.Background(), accountToSave)
	if err != nil {
		return nil, err
	}

	return &models.Account{
		Id:        accountToSave.Id,
		Name:      accountToSave.Name,
		Image:     accountToSave.Image,
		Color:     accountToSave.Color,
		CreatedAt: accountToSave.CreatedAt,
		UpdatedAt: accountToSave.UpdatedAt,
	}, nil
}
