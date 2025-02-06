package usecase

import "github.com/anuntech/finance-backend/internal/domain/models"

type FindAllRepository interface {
	Find() ([]models.Bank, error)
}

type FindByIdRepository interface {
	Find(string) (*models.Bank, error)
}

type CreateRepository interface {
	Create(*models.Bank) (*models.Bank, error)
}

type UpdateRepository interface {
	Update(string, *models.Bank) (*models.Bank, error)
}

type DeleteRepository interface {
	Delete(string) error
}
