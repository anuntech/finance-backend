package helpers

import (
	"context"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func CalculateTransactionBalanceWithEdits(transactions []models.Transaction, db *mongo.Database, isConfirmed bool) float64 {
	var balance float64
	editCollection := db.Collection("edit_transaction")

	for _, t := range transactions {
		cursor, err := editCollection.Find(context.Background(), bson.M{
			"main_id":      t.Id,
			"workspace_id": t.WorkspaceId,
		})

		if err == nil {
			var editTransactions []models.Transaction
			if err := cursor.All(context.Background(), &editTransactions); err == nil && len(editTransactions) > 0 {
				for _, editTransaction := range editTransactions {
					originalValue := CalculateOneTransactionBalance(&t)
					if isConfirmed {
						if editTransaction.IsConfirmed {
							balance += CalculateOneTransactionBalance(&editTransaction) - originalValue
						} else {
							balance -= originalValue
						}
						continue
					}

					balance += CalculateOneTransactionBalance(&editTransaction) - originalValue
				}
				continue
			}
		}

		if isConfirmed && !t.IsConfirmed {
			continue
		}

		balance += CalculateOneTransactionBalance(&t)
	}

	return balance
}

func CalculateOneTransactionBalance(transaction *models.Transaction) float64 {
	var balance float64

	switch transaction.Type {
	case "EXPENSE":
		balance -= transaction.Balance.Value
		balance += transaction.Balance.Discount
		balance -= transaction.Balance.Interest
	case "RECIPE":
		balance += transaction.Balance.Value
		balance -= transaction.Balance.Discount
		balance += transaction.Balance.Interest
	}

	return balance
}
