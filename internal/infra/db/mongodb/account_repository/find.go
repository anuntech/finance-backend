package account_repository

import (
	"context"
	"time"

	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	presentationHelpers "github.com/anuntech/finance-backend/internal/presentation/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindAccountsRepository struct {
	Db *mongo.Database
}

func NewFindAccountsRepository(db *mongo.Database) *FindAccountsRepository {
	return &FindAccountsRepository{
		Db: db,
	}
}

func (c *FindAccountsRepository) Find(globalFilters *presentationHelpers.GlobalFilterParams) ([]models.Account, error) {
	collection := c.Db.Collection("account")

	filter := bson.M{"workspace_id": globalFilters.WorkspaceId}
	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var accounts []models.Account
	if err = cursor.All(context.Background(), &accounts); err != nil {
		return nil, err
	}

	if globalFilters.Month == 0 {
		return accounts, nil
	}

	for index, account := range accounts {
		accounts[index].Balance = c.calculateAccountBalance(account.Id, globalFilters)
	}

	return accounts, nil
}

func (c *FindAccountsRepository) calculateAccountBalance(accountId primitive.ObjectID, globalFilters *presentationHelpers.GlobalFilterParams) float64 {
	doNotRepeatBalance := c.calculateDoNotRepeatAccountBalance(accountId, globalFilters)
	recurringBalance := c.calculateRecurringAccountBalance(accountId, globalFilters)
	return doNotRepeatBalance + recurringBalance
}

func (c *FindAccountsRepository) calculateDoNotRepeatAccountBalance(accountId primitive.ObjectID, globalFilters *presentationHelpers.GlobalFilterParams) float64 {
	collection := c.Db.Collection("transaction")
	startOfMonth := time.Date(globalFilters.Year, time.Month(globalFilters.Month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	filter := bson.M{
		"workspace_id": globalFilters.WorkspaceId,
		"account_id":   accountId,
		"$or": []bson.M{
			{
				"due_date": bson.M{
					"$lt": endOfMonth,
				},
				"is_confirmed": false,
			},
			{
				"confirmation_date": bson.M{
					"$lt": endOfMonth,
				},
				"is_confirmed": true,
			},
		},
		"frequency": "DO_NOT_REPEAT",
	}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return 0.0
	}

	var transactions []models.Transaction
	if err := cursor.All(context.Background(), &transactions); err != nil {
		return 0.0
	}

	return helpers.CalculateTransactionBalance(transactions)
}

func (c *FindAccountsRepository) calculateRecurringAccountBalance(accountId primitive.ObjectID, globalFilters *presentationHelpers.GlobalFilterParams) float64 {
	collection := c.Db.Collection("transaction")

	startOfMonth := time.Date(globalFilters.Year, time.Month(globalFilters.Month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	filter := bson.M{
		"workspace_id": globalFilters.WorkspaceId,
		"account_id":   accountId,
		"frequency":    "RECURRING",
		"$or": []bson.M{
			{
				"due_date": bson.M{
					"$lt": endOfMonth,
				},
				"is_confirmed": false,
			},
			{
				"confirmation_date": bson.M{
					"$lt": endOfMonth,
				},
				"is_confirmed": true,
			},
		},
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return 0.0
	}

	var recurringTransactions []models.Transaction
	if err := cursor.All(context.Background(), &recurringTransactions); err != nil {
		return 0.0
	}

	return helpers.CalculateRecurringTransactionsBalance(recurringTransactions, globalFilters.Year, globalFilters.Month)
}
