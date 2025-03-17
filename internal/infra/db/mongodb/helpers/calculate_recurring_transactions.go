package helpers

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func CalculateRecurringTransactionsBalance(transactions []models.Transaction, year int, month int, db *mongo.Database, isConfirmed bool) float64 {
	var balance float64
	editCollection := db.Collection("edit_transaction")

	for _, t := range transactions {
		var refDate time.Time
		if !t.IsConfirmed {
			refDate = t.DueDate
		} else {
			refDate = *t.ConfirmationDate
		}
		currentCount := MonthsBetween(refDate, year, month) + 1

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
					oneRecurringValue := CalculateOneTransactionBalance(&t)

					if !isConfirmed {
						balance += CalculateOneTransactionBalance(&editTransaction) - oneRecurringValue
						continue
					}

					if editTransaction.IsConfirmed {
						balance += CalculateOneTransactionBalance(&editTransaction) - oneRecurringValue
						continue
					}
					balance -= oneRecurringValue

				}

			}
		}

		if isConfirmed && !t.IsConfirmed {
			continue
		}

		months := MonthsBetween(refDate, year, month)
		balance += (float64(months) + 1) * CalculateOneTransactionBalance(&t)
	}

	return balance
}
