package helpers

import (
	"context"
	"time"

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

func NewCalculateAccountBalanceParams(db *mongo.Database, globalFilters *helpers.GlobalFilterParams, accountId primitive.ObjectID) *CalculateAccountBalanceParams {
	return &CalculateAccountBalanceParams{
		Db:            db,
		GlobalFilters: globalFilters,
		AccountId:     accountId,
	}
}

func (p *CalculateAccountBalanceParams) CalculateAccountBalance() float64 {
	return p.calculateDoNotRepeatAccountBalance()
}

func (p *CalculateAccountBalanceParams) calculateDoNotRepeatAccountBalance() float64 {
	collection := p.Db.Collection("transaction")
	startOfMonth := time.Date(p.GlobalFilters.Year, time.Month(p.GlobalFilters.Month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	filter := bson.M{
		"workspace_id": p.GlobalFilters.WorkspaceId,
		"account_id":   p.AccountId,
		"confirmation_date": bson.M{
			"$lt": endOfMonth,
		},
		"is_confirmed": true,
		"frequency":    "DO_NOT_REPEAT",
	}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return 0.0
	}

	return p.calculateAccountBalance(cursor)
}

func (p *CalculateAccountBalanceParams) calculateAccountBalance(cursor *mongo.Cursor) float64 {
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
		case "RECIPE":
			balance += transaction.Balance.Value
			balance += transaction.Balance.Parts
			balance += transaction.Balance.Labor
			balance -= transaction.Balance.Discount
			balance += transaction.Balance.Interest
		}
	}

	return balance
}
