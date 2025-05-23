package usecase

import (
	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateEditTransactionRepository interface {
	Create(transaction *models.Transaction) (*models.Transaction, error)
}

type FindByIdEditTransactionRepository interface {
	Find(mainId primitive.ObjectID, mainCount int, workspaceId primitive.ObjectID) (*models.Transaction, error)
	FindMany(params []struct {
		MainId      primitive.ObjectID
		MainCount   int
		WorkspaceId primitive.ObjectID
	}) ([]*models.Transaction, error)
}

type UpdateEditTransactionRepository interface {
	Update(transaction *models.Transaction) (*models.Transaction, error)
}

type DeleteEditTransactionRepository interface {
	Delete(editTransactionParams []struct {
		MainId      primitive.ObjectID
		MainCount   int
		WorkspaceId primitive.ObjectID
	}) error
}
