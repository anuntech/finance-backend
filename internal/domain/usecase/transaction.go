package usecase

import "github.com/anuntech/finance-backend/internal/domain/models"

type CreateTransactionRepository interface {
	Create(transaction *models.Transaction) (*models.Transaction, error)
}
