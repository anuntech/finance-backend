package helpers

import (
	"github.com/anuntech/finance-backend/internal/domain/models"
)

func CalculateTransactionBalance(transactions []models.Transaction) float64 {
	var balance float64

	for _, transaction := range transactions {
		switch transaction.Type {
		case "EXPENSE":
			balance -= transaction.Balance.Value
			balance -= transaction.Balance.Parts
			balance -= transaction.Balance.Labor
			balance += transaction.Balance.Discount
			balance -= transaction.Balance.Interest
		case "RECIPE":
			balance += transaction.Balance.Value
			balance += transaction.Balance.Parts
			balance += transaction.Balance.Labor
			balance -= transaction.Balance.Discount
			balance += transaction.Balance.Interest
		}
	}

	return balance
}
