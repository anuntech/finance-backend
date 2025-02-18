package usecase

import (
	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateAccountRepository interface {
	Create(*models.AccountInput) (*models.Account, error)
}

type FindAccountByWorkspaceIdRepository interface {
	Find(primitive.ObjectID) ([]models.Account, error)
}

type FindAccountByIdRepository interface {
	Find(primitive.ObjectID, primitive.ObjectID) (*models.Account, error)
}

type DeleteAccountRepository interface {
	Delete(accountIds []primitive.ObjectID, workspaceId primitive.ObjectID) error
}

type UpdateAccountRepository interface {
	Update(primitive.ObjectID, *models.AccountInput) (*models.Account, error)
}

type ImportAccountsRepository interface {
	Import(accounts []models.AccountInput, workspaceId primitive.ObjectID) ([]models.Account, error)
}
