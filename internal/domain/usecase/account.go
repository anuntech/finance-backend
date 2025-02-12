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
	Find(string, string) (*models.Account, error)
}

type DeleteAccountRepository interface {
	Delete(string) error
}

type UpdateAccountRepository interface {
	Update(string, *models.AccountInput) (*models.Account, error)
}
