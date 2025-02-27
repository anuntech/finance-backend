package usecase

import (
	"github.com/anuntech/finance-backend/internal/domain/models"
	"github.com/anuntech/finance-backend/internal/presentation/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateTransactionRepository interface {
	Create(transaction *models.Transaction) (*models.Transaction, error)
}

type FindTransactionsByWorkspaceIdInputRepository struct {
	WorkspaceId primitive.ObjectID
	Month       int
	Year        int
	Type        string
}

type FindTransactionsByWorkspaceIdRepository interface {
	Find(data *helpers.GlobalFilterParams) ([]models.Transaction, error)
}

type FindTransactionByIdRepository interface {
	Find(transactionId primitive.ObjectID, workspaceId primitive.ObjectID) (*models.Transaction, error)
}

type UpdateTransactionRepository interface {
	Update(transactionId primitive.ObjectID, transaction *models.Transaction) (*models.Transaction, error)
}

type DeleteTransactionRepository interface {
	Delete(transactionIds []primitive.ObjectID, workspaceId primitive.ObjectID) error
}
