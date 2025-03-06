package helpers

import (
	"fmt"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
)

func CalculateRepeatTransactionsBalance(transactions []models.Transaction, year int, month int) float64 {
	var balance float64
	fmt.Println(transactions)
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
		case "YEARLY":
			balance += repeatYearlyTransaction(&t, refDate, year)
		case "QUARTERLY":
			balance += repeatQuarterlyTransaction(&t, refDate, year, month)
		}
	}

	return balance
}

func repeatMonthlyTransaction(t *models.Transaction, refDate time.Time, year int, month int) float64 {
	monthsBetween := MonthsBetween(refDate, year, month)

	effectiveInstallment := int(t.RepeatSettings.InitialInstallment) + monthsBetween

	installmentValue := CalculateOneTransactionBalance(t) / float64(t.RepeatSettings.Count)

	if effectiveInstallment >= t.RepeatSettings.Count {
		return CalculateOneTransactionBalance(t)
	}

	return installmentValue * float64(effectiveInstallment)
}

func repeatYearlyTransaction(t *models.Transaction, refDate time.Time, year int) float64 {
	yearsBetween := YearsBetween(refDate, year)

	effectiveInstallment := int(t.RepeatSettings.InitialInstallment) + yearsBetween

	installmentValue := CalculateOneTransactionBalance(t) / float64(t.RepeatSettings.Count)

	if effectiveInstallment >= t.RepeatSettings.Count {
		return CalculateOneTransactionBalance(t)
	}

	return installmentValue * float64(effectiveInstallment)
}

func repeatQuarterlyTransaction(t *models.Transaction, refDate time.Time, year int, month int) float64 {
	quartersBetween := QuartersBetween(refDate, year, month)

	effectiveInstallment := int(t.RepeatSettings.InitialInstallment) + quartersBetween

	installmentValue := CalculateOneTransactionBalance(t) / float64(t.RepeatSettings.Count)

	if effectiveInstallment >= t.RepeatSettings.Count {
		return CalculateOneTransactionBalance(t)
	}

	return installmentValue * float64(effectiveInstallment)
}
