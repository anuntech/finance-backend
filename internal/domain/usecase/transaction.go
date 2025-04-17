package usecase

import (
	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateTransactionRepository interface {
	Create(transaction *models.Transaction) (*models.Transaction, error)
}

type FindTransactionsByWorkspaceIdInputRepository struct {
	Month       int
	Year        int
	Type        string
	InitialDate string
	FinalDate   string
	WorkspaceId primitive.ObjectID
	AccountIds  []primitive.ObjectID
	Limit       int
	Offset      int
}

type FindTransactionsByWorkspaceIdRepository interface {
	Find(data *FindTransactionsByWorkspaceIdInputRepository) ([]models.Transaction, error)
}

type FindTransactionByIdRepository interface {
	Find(transactionId primitive.ObjectID, workspaceId primitive.ObjectID) (*models.Transaction, error)
}

type UpdateTransactionRepository interface {
	Update(transactionId primitive.ObjectID, transaction *models.Transaction) (*models.Transaction, error)
}

type DeleteTransactionRepository interface {
	Delete(transactionIds []primitive.ObjectID, workspaceId primitive.ObjectID) error
	DeleteEditTransactions(editTransactionParams []struct {
		MainId      primitive.ObjectID
		MainCount   int
		WorkspaceId primitive.ObjectID
	}, findTransactionById FindTransactionByIdRepository) error
}
