package category_repository

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

type FindCategoriesRepository struct {
	Db *mongo.Database
}

func NewFindCategoriesRepository(db *mongo.Database) *FindCategoriesRepository {
	return &FindCategoriesRepository{
		Db: db,
	}
}

func (r *FindCategoriesRepository) Find(globalFilters *presentationHelpers.GlobalFilterParams) ([]models.Category, error) {
	collection := r.Db.Collection("category")

	filter := bson.M{"workspace_id": globalFilters.WorkspaceId}
	if globalFilters.Type != "" {
		filter["type"] = globalFilters.Type
	}

	ctx, cancel := context.WithTimeout(context.Background(), helpers.Timeout)
	defer cancel()

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var categories []models.Category
	if err = cursor.All(context.Background(), &categories); err != nil {
		return nil, err
	}

	if globalFilters.Month == 0 {
		return categories, nil
	}

	// Optimize by processing all subcategories in batch
	if err := r.calculateCategoryBalances(categories, globalFilters); err != nil {
		return nil, err
	}

	return categories, nil
}

// New optimized method that maintains the same calculation logic but with batch processing
func (r *FindCategoriesRepository) calculateCategoryBalances(categories []models.Category, globalFilters *presentationHelpers.GlobalFilterParams) error {
	// Collect all subcategory IDs from all categories
	var subCategoryIDs []primitive.ObjectID
	subCategoryMap := make(map[primitive.ObjectID]*models.SubCategoryCategory)

	for i := range categories {
		for j := range categories[i].SubCategories {
			subID := categories[i].SubCategories[j].Id
			subCategoryIDs = append(subCategoryIDs, subID)
			// Store reference to subcategory for later update
			subCategoryMap[subID] = &categories[i].SubCategories[j]
			// Reset amounts to avoid accumulation issues
			categories[i].SubCategories[j].Amount = 0
			categories[i].SubCategories[j].CurrentAmount = 0
		}
	}

	if len(subCategoryIDs) == 0 {
		return nil
	}

	// Prepare date range
	startOfMonth := time.Date(globalFilters.Year, time.Month(globalFilters.Month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	// Build filter for balance calculation (includes both confirmed and unconfirmed)
	balanceFilter := bson.M{
		"workspace_id": globalFilters.WorkspaceId,
		"$and": []bson.M{
			{"$or": []bson.M{
				{"sub_category_id": bson.M{"$in": subCategoryIDs}},
				{"tags.sub_tag_id": bson.M{"$in": subCategoryIDs}},
			}},
			{"$or": []bson.M{
				{"$and": []bson.M{
					{
						"due_date": bson.M{
							"$lt": endOfMonth,
						},
						"is_confirmed": false,
					},
				}},
				{"$and": []bson.M{
					{
						"confirmation_date": bson.M{
							"$lt": endOfMonth,
						},
						"is_confirmed": true,
					},
				}},
			}},
		},
	}

	// Get all transactions in batch
	balanceTransactions, err := r.fetchTransactions(balanceFilter)
	if err != nil {
		return err
	}

	// Group transactions by subcategory ID and frequency
	transactionsBySubCategoryAndFrequency := make(map[primitive.ObjectID]map[string][]models.Transaction)
	currentTransactionsBySubCategoryAndFrequency := make(map[primitive.ObjectID]map[string][]models.Transaction)

	// Initialize maps for each subcategory
	for _, subCategoryID := range subCategoryIDs {
		transactionsBySubCategoryAndFrequency[subCategoryID] = make(map[string][]models.Transaction)
		transactionsBySubCategoryAndFrequency[subCategoryID]["DO_NOT_REPEAT"] = []models.Transaction{}
		transactionsBySubCategoryAndFrequency[subCategoryID]["RECURRING"] = []models.Transaction{}
		transactionsBySubCategoryAndFrequency[subCategoryID]["REPEAT"] = []models.Transaction{}

		currentTransactionsBySubCategoryAndFrequency[subCategoryID] = make(map[string][]models.Transaction)
		currentTransactionsBySubCategoryAndFrequency[subCategoryID]["DO_NOT_REPEAT"] = []models.Transaction{}
		currentTransactionsBySubCategoryAndFrequency[subCategoryID]["RECURRING"] = []models.Transaction{}
		currentTransactionsBySubCategoryAndFrequency[subCategoryID]["REPEAT"] = []models.Transaction{}
	}

	// Group balance transactions by subcategory and frequency
	for _, tx := range balanceTransactions {
		// Handle direct subcategory reference
		if tx.SubCategoryId != nil {
			subCategoryID := *tx.SubCategoryId
			if _, exists := transactionsBySubCategoryAndFrequency[subCategoryID]; exists {
				transactionsBySubCategoryAndFrequency[subCategoryID][tx.Frequency] = append(
					transactionsBySubCategoryAndFrequency[subCategoryID][tx.Frequency], tx)
			}
		}

		// Handle tag references
		for _, tag := range tx.Tags {
			subTagID := tag.SubTagId
			if _, exists := transactionsBySubCategoryAndFrequency[subTagID]; exists {
				transactionsBySubCategoryAndFrequency[subTagID][tx.Frequency] = append(
					transactionsBySubCategoryAndFrequency[subTagID][tx.Frequency], tx)
			}
		}
	}

	// Group current balance transactions by subcategory and frequency
	for _, tx := range balanceTransactions {
		// Handle direct subcategory reference
		if tx.SubCategoryId != nil {
			subCategoryID := *tx.SubCategoryId
			if _, exists := currentTransactionsBySubCategoryAndFrequency[subCategoryID]; exists {
				currentTransactionsBySubCategoryAndFrequency[subCategoryID][tx.Frequency] = append(
					currentTransactionsBySubCategoryAndFrequency[subCategoryID][tx.Frequency], tx)
			}
		}

		// Handle tag references
		for _, tag := range tx.Tags {
			subTagID := tag.SubTagId
			if _, exists := currentTransactionsBySubCategoryAndFrequency[subTagID]; exists {
				currentTransactionsBySubCategoryAndFrequency[subTagID][tx.Frequency] = append(
					currentTransactionsBySubCategoryAndFrequency[subTagID][tx.Frequency], tx)
			}
		}
	}

	// Use WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Calculate amounts for each subcategory using concurrent processing
	for subCategoryID, subCategory := range subCategoryMap {
		wg.Add(1)
		go func(scID primitive.ObjectID, sc *models.SubCategoryCategory) {
			defer wg.Done()

			// Calculate balance
			doNotRepeatBalance := helpers.CalculateTransactionBalanceWithEdits(
				transactionsBySubCategoryAndFrequency[scID]["DO_NOT_REPEAT"], r.Db, false)

			// Set the total balance (includes both confirmed and unconfirmed)
			sc.Amount += doNotRepeatBalance
		}(subCategoryID, subCategory)

		wg.Add(1)
		go func(scID primitive.ObjectID, sc *models.SubCategoryCategory) {
			defer wg.Done()

			recurringBalance := helpers.CalculateRecurringTransactionsBalance(
				transactionsBySubCategoryAndFrequency[scID]["RECURRING"], globalFilters.Year, globalFilters.Month, r.Db, false)

			sc.Amount += recurringBalance
		}(subCategoryID, subCategory)

		wg.Add(1)
		go func(scID primitive.ObjectID, sc *models.SubCategoryCategory) {
			defer wg.Done()

			repeatBalance := helpers.CalculateRepeatTransactionsBalance(
				transactionsBySubCategoryAndFrequency[scID]["REPEAT"], globalFilters.Year, globalFilters.Month, r.Db, false)

			sc.Amount += repeatBalance
		}(subCategoryID, subCategory)

		wg.Add(1)
		go func(scID primitive.ObjectID, sc *models.SubCategoryCategory) {
			defer wg.Done()

			// Calculate current balance
			doNotRepeatCurrentBalance := helpers.CalculateTransactionBalanceWithEdits(
				currentTransactionsBySubCategoryAndFrequency[scID]["DO_NOT_REPEAT"], r.Db, true)

			// Set the current balance (includes only confirmed transactions)
			sc.CurrentAmount += doNotRepeatCurrentBalance
		}(subCategoryID, subCategory)

		wg.Add(1)
		go func(scID primitive.ObjectID, sc *models.SubCategoryCategory) {
			defer wg.Done()

			recurringCurrentBalance := helpers.CalculateRecurringTransactionsBalance(
				currentTransactionsBySubCategoryAndFrequency[scID]["RECURRING"], globalFilters.Year, globalFilters.Month, r.Db, true)

			sc.CurrentAmount += recurringCurrentBalance
		}(subCategoryID, subCategory)

		wg.Add(1)
		go func(scID primitive.ObjectID, sc *models.SubCategoryCategory) {
			defer wg.Done()

			repeatCurrentBalance := helpers.CalculateRepeatTransactionsBalance(
				currentTransactionsBySubCategoryAndFrequency[scID]["REPEAT"], globalFilters.Year, globalFilters.Month, r.Db, true)

			sc.CurrentAmount += repeatCurrentBalance
		}(subCategoryID, subCategory)
	}

	// Wait for all subcategory calculations to complete
	wg.Wait()

	// Calculate total amounts for each category based on their subcategories
	for i := range categories {
		categories[i].Amount = 0
		categories[i].CurrentAmount = 0

		for _, subCategory := range categories[i].SubCategories {
			categories[i].Amount += subCategory.Amount
			categories[i].CurrentAmount += subCategory.CurrentAmount
		}
	}

	return nil
}

// Helper method to fetch transactions
func (r *FindCategoriesRepository) fetchTransactions(filter bson.M) ([]models.Transaction, error) {
	collection := r.Db.Collection("transaction")
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

//
// I used these methods below to filter transactions date with golang tools
//

// func (r *FindCategoriesRepository) fetchTransactions(globalFilters *presentationHelpers.GlobalFilterParams, subCategoryIDs []primitive.ObjectID) ([]models.Transaction, error) {
// 	collection := r.Db.Collection("transaction")

// 	filter := bson.M{
// 		"workspace_id": globalFilters.WorkspaceId,
// 		"$or": []bson.M{
// 			{"sub_category_id": bson.M{"$in": subCategoryIDs}},
// 			{"tags.sub_tag_id": bson.M{"$in": subCategoryIDs}},
// 		},
// 	}

// 	cursor, err := collection.Find(context.Background(), filter)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var transactions []models.Transaction
// 	if err := cursor.All(context.Background(), &transactions); err != nil {
// 		return nil, err
// 	}

// 	return transactions, nil
// }

// func (r *FindCategoriesRepository) filterTransactions(transactions []models.Transaction, globalFilters *presentationHelpers.GlobalFilterParams) []models.Transaction {
// 	filtered := []models.Transaction{}
// 	startOfMonth := time.Date(globalFilters.Year, time.Month(globalFilters.Month), 1, 0, 0, 0, 0, time.UTC)
// 	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

// 	for _, tx := range transactions {
// 		if !tx.IsConfirmed && !tx.DueDate.IsZero() && tx.DueDate.Before(endOfMonth) {
// 			filtered = append(filtered, tx)
// 			continue
// 		}

// 		if tx.IsConfirmed && tx.ConfirmationDate != nil && tx.ConfirmationDate.Before(endOfMonth) {
// 			filtered = append(filtered, tx)
// 		}
// 	}

// 	return filtered
// }
