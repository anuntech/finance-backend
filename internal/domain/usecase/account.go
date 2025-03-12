package usecase

import (
	"github.com/anuntech/finance-backend/internal/domain/models"
	presentationHelpers "github.com/anuntech/finance-backend/internal/presentation/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateAccountRepository interface {
	Create(*models.Account) (*models.Account, error)
}

type FindAccountByWorkspaceIdRepository interface {
	Find(globalFilters *presentationHelpers.GlobalFilterParams) ([]models.Account, error)
}

type FindAccountByIdRepository interface {
	Find(accountId primitive.ObjectID, workspaceId primitive.ObjectID) (*models.Account, error)
}

type DeleteAccountRepository interface {
	Delete(accountIds []primitive.ObjectID, workspaceId primitive.ObjectID) error
}

type UpdateAccountRepository interface {
	Update(primitive.ObjectID, *models.Account) (*models.Account, error)
}

type ImportAccountsRepository interface {
	Import(accounts []models.Account, workspaceId primitive.ObjectID) ([]models.Account, error)
}
