package usecase

import "github.com/anuntech/finance-backend/internal/domain/models"

type FindAllBankRepository interface {
	Find() ([]models.Bank, error)
}

type FindBankByIdRepository interface {
	Find(string) (*models.Bank, error)
}

type FindBankByNameRepository interface {
	FindByName(string) (*models.Bank, error)
}
