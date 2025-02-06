package usecase

import "github.com/anuntech/finance-backend/internal/domain/models"

type CreateAccountRepository interface {
	Create(*models.AccountInput) (*models.Account, error)
}

type FindAccountByWorkspaceIdRepository interface {
	Find(string, string) ([]models.Account, error)
}

type FindAccountByIdRepository interface {
	Find(string) (*models.Account, error)
}

type DeleteAccountRepository interface {
	Delete(string) error
}

type UpdateAccountRepository interface {
	Update(string, *models.AccountInput) (*models.Account, error)
}
