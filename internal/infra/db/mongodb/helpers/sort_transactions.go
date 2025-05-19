package helpers

import (
	"sort"

	"github.com/anuntech/finance-backend/internal/domain/models"
)

// SortTransactionsByDueDate sorts a slice of transactions by their due date
// in descending order (newest transactions first)
func SortTransactionsByDueDate(transactions []models.Transaction) {
	sort.Slice(transactions, func(i, j int) bool {
		// Sort in descending order (newest first)
		return transactions[j].DueDate.Before(transactions[i].DueDate)
	})
}
