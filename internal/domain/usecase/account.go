package usecase

import "github.com/anuntech/finance-backend/internal/domain/models"

type CreateAccount interface {
	Create(*models.AccountInput) (*models.Account, error)
}

type FindManyByUserIdAndWorkspaceId interface {
	FindManyByUserIdAndWorkspaceId(string, string) ([]models.Account, error)
}
