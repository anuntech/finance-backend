package helpers

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func CalculateRepeatTransactionsBalance(transactions []models.Transaction, year int, month int, db *mongo.Database) float64 {
	var balance float64
	for _, t := range transactions {

		editCollection := db.Collection("edit_transaction")
		var editTransaction models.Transaction

		result := editCollection.FindOne(context.Background(), bson.M{
			"main_id":      t.Id,
			"workspace_id": t.WorkspaceId,
			"main_count":   CalculateCurrentCount(&t, year, month),
		})

		if result.Err() == nil && result.Decode(&editTransaction) == nil {
			// If an edited transaction exists for this count, use its full balance
			oneRepeatToRemove := CalculateOneTransactionBalance(&t) / float64(t.RepeatSettings.Count)
			balance += CalculateOneTransactionBalance(&editTransaction) - oneRepeatToRemove
		}

		// Otherwise, calculate normally based on interval
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

	if effectiveInstallment >= t.RepeatSettings.Count {
		return CalculateOneTransactionBalance(t)
	}

	installmentValue := CalculateOneTransactionBalance(t) / float64(t.RepeatSettings.Count)
	return installmentValue * float64(effectiveInstallment)
}

func repeatYearlyTransaction(t *models.Transaction, refDate time.Time, year int) float64 {
	yearsBetween := YearsBetween(refDate, year)

	effectiveInstallment := int(t.RepeatSettings.InitialInstallment) + yearsBetween

	if effectiveInstallment >= t.RepeatSettings.Count {
		return CalculateOneTransactionBalance(t)
	}

	installmentValue := CalculateOneTransactionBalance(t) / float64(t.RepeatSettings.Count)
	return installmentValue * float64(effectiveInstallment)
}

func repeatQuarterlyTransaction(t *models.Transaction, refDate time.Time, year int, month int) float64 {
	quartersBetween := QuartersBetween(refDate, year, month)

	effectiveInstallment := int(t.RepeatSettings.InitialInstallment) + quartersBetween

	if effectiveInstallment >= t.RepeatSettings.Count {
		return CalculateOneTransactionBalance(t)
	}

	installmentValue := CalculateOneTransactionBalance(t) / float64(t.RepeatSettings.Count)
	return installmentValue * float64(effectiveInstallment)
}
