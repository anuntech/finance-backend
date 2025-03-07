package usecase

import "github.com/anuntech/finance-backend/internal/domain/models"

type CreateEditTransactionRepository interface {
	Create(transaction *models.Transaction) (*models.Transaction, error)
}
