package helpers

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
)

type CalculateAccountBalanceParams struct {
	Db            *mongo.Database
	GlobalFilters *helpers.GlobalFilterParams
	AccountId     primitive.ObjectID
}

func CalculateAccountBalance(params *CalculateAccountBalanceParams) float64 {
	collection := params.Db.Collection("transaction")

	filter := bson.M{
		"account_id": params.AccountId,
		"$or": []bson.M{
			{
				"frequency": "DO_NOT_REPEAT",
			},
			{
				"frequency": "REPEAT",
			},
		},
	}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return 0.0
	}

	return calculateAccountBalance(cursor)
}

func calculateAccountBalance(cursor *mongo.Cursor) float64 {
	balance := 0.0

	var transactions []models.Transaction
	if err := cursor.All(context.Background(), &transactions); err != nil {
		return 0.0
	}

	for _, transaction := range transactions {
		switch transaction.Type {
		case "EXPENSE":
			balance -= transaction.Balance.Value
			balance -= transaction.Balance.Parts
			balance -= transaction.Balance.Labor
			balance += transaction.Balance.Discount
			balance -= transaction.Balance.Interest
		case "INCOME":
			balance += transaction.Balance.Value
			balance += transaction.Balance.Parts
			balance += transaction.Balance.Labor
			balance -= transaction.Balance.Discount
			balance += transaction.Balance.Interest
		}
	}

	return balance
}
