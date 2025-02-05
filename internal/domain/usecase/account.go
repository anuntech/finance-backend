package usecase

import "github.com/anuntech/finance-backend/internal/domain/models"

type CreateAccount interface {
	Create(*models.AccountInput) (*models.Account, error)
}

type FindByWorkspaceId interface {
	Find(string, string) ([]models.Account, error)
}

type FindById interface {
	Find(string) (*models.Account, error)
}

type DeleteAccount interface {
	Delete(string) error
}

type UpdateAccount interface {
	Update(string, *models.AccountInput) (*models.Account, error)
}
