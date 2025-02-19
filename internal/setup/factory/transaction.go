package factory

import (
	"github.com/anuntech/finance-backend/internal/presentation/controllers/transaction"
)

func MakeCreateTransactionController() *transaction.CreateTransactionController {
	return transaction.NewCreateTransactionController()
}
