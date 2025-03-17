package account_repository

import (
	"context"
	"sync"
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

	// Optimize by fetching all transactions for all accounts in batch
	if err := c.calculateAccountBalances(accounts, globalFilters); err != nil {
		return nil, err
	}

	return accounts, nil
}

// New optimized method that maintains the same calculation logic but with batch processing
func (c *FindAccountsRepository) calculateAccountBalances(accounts []models.Account, globalFilters *presentationHelpers.GlobalFilterParams) error {
	if len(accounts) == 0 {
		return nil
	}

	// Extract all account IDs
	accountIDs := make([]primitive.ObjectID, len(accounts))
	accountMap := make(map[primitive.ObjectID]*models.Account)
	for i, account := range accounts {
		accountIDs[i] = account.Id
		accountMap[account.Id] = &accounts[i]
		// Reset balances to avoid accumulation issues
		accounts[i].Balance = account.Balance        // Preserve initial balance
		accounts[i].CurrentBalance = account.Balance // Start with same base
	}

	// Prepare date range
	startOfMonth := time.Date(globalFilters.Year, time.Month(globalFilters.Month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	// Get all transactions in one query (for total balance)
	balanceFilter := bson.M{
		"workspace_id": globalFilters.WorkspaceId,
		"account_id":   bson.M{"$in": accountIDs},
		"$or": []bson.M{
			{
				"due_date":     bson.M{"$lt": endOfMonth},
				"is_confirmed": false,
			},
			{
				"confirmation_date": bson.M{"$lt": endOfMonth},
				"is_confirmed":      true,
			},
		},
	}

	balanceTransactions, err := c.fetchTransactions(balanceFilter)
	if err != nil {
		return err
	}

	// Group transactions by account and frequency
	balanceByAccountAndFrequency := make(map[primitive.ObjectID]map[string][]models.Transaction)
	currentBalanceByAccountAndFrequency := make(map[primitive.ObjectID]map[string][]models.Transaction)

	// Initialize maps for all accounts
	for _, accountID := range accountIDs {
		balanceByAccountAndFrequency[accountID] = make(map[string][]models.Transaction)
		balanceByAccountAndFrequency[accountID]["DO_NOT_REPEAT"] = []models.Transaction{}
		balanceByAccountAndFrequency[accountID]["RECURRING"] = []models.Transaction{}
		balanceByAccountAndFrequency[accountID]["REPEAT"] = []models.Transaction{}

		currentBalanceByAccountAndFrequency[accountID] = make(map[string][]models.Transaction)
		currentBalanceByAccountAndFrequency[accountID]["DO_NOT_REPEAT"] = []models.Transaction{}
		currentBalanceByAccountAndFrequency[accountID]["RECURRING"] = []models.Transaction{}
		currentBalanceByAccountAndFrequency[accountID]["REPEAT"] = []models.Transaction{}
	}

	// Group balance transactions
	for _, tx := range balanceTransactions {
		if tx.AccountId != nil {
			accountID := *tx.AccountId
			if _, exists := balanceByAccountAndFrequency[accountID]; exists {
				balanceByAccountAndFrequency[accountID][tx.Frequency] = append(
					balanceByAccountAndFrequency[accountID][tx.Frequency], tx)
			}
		}
	}

	// Group current balance transactions
	for _, tx := range balanceTransactions {
		if tx.AccountId != nil {
			accountID := *tx.AccountId
			if _, exists := currentBalanceByAccountAndFrequency[accountID]; exists {
				currentBalanceByAccountAndFrequency[accountID][tx.Frequency] = append(
					currentBalanceByAccountAndFrequency[accountID][tx.Frequency], tx)
			}
		}
	}

	// Use a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Calculate balances for each account concurrently
	for accountID, account := range accountMap {
		wg.Add(1)
		go func(accID primitive.ObjectID, acc *models.Account) {
			defer wg.Done()

			// Calculate balance using the same logic as the original methods
			doNotRepeatBalance := helpers.CalculateTransactionBalanceWithEdits(
				balanceByAccountAndFrequency[accID]["DO_NOT_REPEAT"], c.Db, false)
			recurringBalance := helpers.CalculateRecurringTransactionsBalance(
				balanceByAccountAndFrequency[accID]["RECURRING"], globalFilters.Year, globalFilters.Month, c.Db, false)
			repeatBalance := helpers.CalculateRepeatTransactionsBalance(
				balanceByAccountAndFrequency[accID]["REPEAT"], globalFilters.Year, globalFilters.Month, c.Db, false)

			// The total balance includes both confirmed and unconfirmed transactions
			acc.Balance += doNotRepeatBalance + recurringBalance + repeatBalance
		}(accountID, account)

		wg.Add(1)
		go func(accID primitive.ObjectID, acc *models.Account) {
			defer wg.Done()

			doNotRepeatCurrentBalance := helpers.CalculateTransactionBalanceWithEdits(
				currentBalanceByAccountAndFrequency[accID]["DO_NOT_REPEAT"], c.Db, true)
			recurringCurrentBalance := helpers.CalculateRecurringTransactionsBalance(
				currentBalanceByAccountAndFrequency[accID]["RECURRING"], globalFilters.Year, globalFilters.Month, c.Db, true)
			repeatCurrentBalance := helpers.CalculateRepeatTransactionsBalance(
				currentBalanceByAccountAndFrequency[accID]["REPEAT"], globalFilters.Year, globalFilters.Month, c.Db, true)

			acc.CurrentBalance += doNotRepeatCurrentBalance + recurringCurrentBalance + repeatCurrentBalance
		}(accountID, account)
	}

	wg.Wait()

	return nil
}

// Helper method to fetch transactions
func (c *FindAccountsRepository) fetchTransactions(filter bson.M) ([]models.Transaction, error) {
	collection := c.Db.Collection("transaction")
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	var transactions []models.Transaction
	if err := cursor.All(context.Background(), &transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}
