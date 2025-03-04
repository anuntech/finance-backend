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

		switch t.RepeatSettings.Interval {
		case "MONTHLY":
			balance += repeatMonthlyTransaction(&t, refDate, year, month)
		}

	}

	return balance
}

func repeatMonthlyTransaction(t *models.Transaction, refDate time.Time, year int, month int) float64 {
	monthsBetween := MonthsBetween(refDate, year, month)

	installmentValue := CalculateOneTransactionBalance(t) / float64(t.RepeatSettings.Count)

	if monthsBetween+1 >= t.RepeatSettings.Count {
		return CalculateOneTransactionBalance(t)
	}

	return installmentValue * float64(monthsBetween+1)
}
