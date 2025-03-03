package helpers

import (
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
)

func CalculateRecurringTransactionsBalance(transactions []models.Transaction, year int, month int) float64 {
	var balance float64
	for _, t := range transactions {
		var refDate time.Time
		if !t.IsConfirmed {
			refDate = t.DueDate
		} else {
			refDate = *t.ConfirmationDate
		}

		months := MonthsBetween(refDate, year, month)

		balance += (float64(months) + 1) * CalculateOneTransactionBalance(t)
	}

	return balance
}
