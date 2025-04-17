package helpers

import (
	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/mongo"
)

func CalculateTransactionBalanceWithEdits(transactions []models.Transaction, db *mongo.Database, isConfirmed bool) float64 {
	var balance float64

	for _, t := range transactions {
		if isConfirmed && !t.IsConfirmed {
			continue
		}

		balance += CalculateOneTransactionBalance(&t)
	}

	return balance
}

func CalculateOneTransactionBalance(transaction *models.Transaction) float64 {
	var balance float64
	var multiplier float64

	switch transaction.Type {
	case "EXPENSE":
		multiplier = -1
	case "RECIPE":
		multiplier = 1
	}

	balance += transaction.Balance.Value * multiplier

	balance -= transaction.Balance.Discount * multiplier

	if transaction.Balance.DiscountPercentage > 0 {
		discountAmount := balance * (transaction.Balance.DiscountPercentage / 100)
		balance -= discountAmount
	}

	balance += transaction.Balance.Interest * multiplier

	if transaction.Balance.InterestPercentage > 0 {
		interestAmount := balance * (transaction.Balance.InterestPercentage / 100)
		balance += interestAmount * multiplier
	}

	return balance
}
