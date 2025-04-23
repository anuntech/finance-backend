package usecase

import (
	"github.com/anuntech/finance-backend/internal/domain/models"
	presentationHelpers "github.com/anuntech/finance-backend/internal/presentation/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateCreditCardRepository defines the interface for creating credit cards
type CreateCreditCardRepository interface {
	Create(*models.CreditCard) (*models.CreditCard, error)
}

// FindCreditCardsRepository defines the interface for retrieving multiple credit cards
type FindCreditCardsRepository interface {
	Find(globalFilters *presentationHelpers.GlobalFilterParams) ([]models.CreditCard, error)
}

// FindCreditCardByIdRepository defines the interface for retrieving a single credit card by ID
type FindCreditCardByIdRepository interface {
	Find(creditCardId primitive.ObjectID, workspaceId primitive.ObjectID) (*models.CreditCard, error)
}

// FindCreditCardByNameAndWorkspaceIdRepository defines the interface for finding credit cards by name within a workspace
type FindCreditCardByNameAndWorkspaceIdRepository interface {
	FindByNameAndWorkspaceId(name string, workspaceId primitive.ObjectID) (*models.CreditCard, error)
}

// UpdateCreditCardRepository defines the interface for updating credit cards
type UpdateCreditCardRepository interface {
	Update(creditCardId primitive.ObjectID, creditCard *models.CreditCard) (*models.CreditCard, error)
}

// DeleteCreditCardRepository defines the interface for deleting credit cards
type DeleteCreditCardRepository interface {
	Delete(creditCardIds []primitive.ObjectID, workspaceId primitive.ObjectID) error
}
