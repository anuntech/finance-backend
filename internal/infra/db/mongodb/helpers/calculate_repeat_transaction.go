package helpers

import (
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
)

func CalculateRepeatTransactionsBalance(transactions []models.Transaction, year int, month int) float64 {
	var balance float64
	for _, t := range transactions {
		var refDate time.Time
		if !t.IsConfirmed {
			refDate = t.DueDate
		} else {
			refDate = *t.ConfirmationDate
		}

		monthsBetween := MonthsBetween(refDate, year, month)

		if monthsBetween >= t.RepeatSettings.Count {
			balance += float64(t.RepeatSettings.Count) * CalculateOneTransactionBalance(t)
			continue
		}

		balance += (float64(monthsBetween) + 1) * CalculateOneTransactionBalance(t)
	}

	return balance
}
