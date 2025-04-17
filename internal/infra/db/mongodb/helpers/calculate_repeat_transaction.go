package helpers

import (
	"context"
	"sync"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func CalculateRepeatTransactionsBalance(transactions []models.Transaction, year int, month int, db *mongo.Database, isConfirmed bool) float64 {
	var balance float64
	editCollection := db.Collection("edit_transaction")

	var wg sync.WaitGroup

	for _, t := range transactions {
		wg.Add(1)
		go func(t models.Transaction) {
			defer wg.Done()
			// Get the current count for this transaction in the specified month
			currentCount := CalculateCurrentCount(&t, year, month)

			// Find all edits for this transaction with main_count <= currentCount
			cursor, err := editCollection.Find(context.Background(), bson.M{
				"main_id":      t.Id,
				"workspace_id": t.WorkspaceId,
				"main_count":   bson.M{"$lte": currentCount},
			})

			// Process each edit if we found any
			if err == nil {
				var editTransactions []models.Transaction
				if err := cursor.All(context.Background(), &editTransactions); err == nil && len(editTransactions) > 0 {
					// Apply the balance adjustments for each edit
					for _, editTransaction := range editTransactions {
						oneRecurringValue := CalculateOneTransactionBalance(&t) / float64(t.RepeatSettings.Count)
						if !isConfirmed {
							balance += CalculateOneTransactionBalance(&editTransaction) - oneRecurringValue
							continue
						}

						if editTransaction.IsConfirmed {
							balance += CalculateOneTransactionBalance(&editTransaction)
							continue
						}
						balance -= oneRecurringValue

					}
				}
			}

			if isConfirmed && !t.IsConfirmed {
				return
			}

			// If no edits were found or there was an error, calculate normally based on interval
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
		}(t)
	}

	wg.Wait()

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
