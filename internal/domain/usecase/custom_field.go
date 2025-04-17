package usecase

import (
	"github.com/anuntech/finance-backend/internal/domain/models"
	presentationHelpers "github.com/anuntech/finance-backend/internal/presentation/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateCustomFieldRepository interface {
	Create(customField *models.CustomField) (*models.CustomField, error)
}

type FindCustomFieldsRepository interface {
	Find(globalFilters *presentationHelpers.GlobalFilterParams) ([]models.CustomField, error)
}

type FindCustomFieldByIdRepository interface {
	Find(customFieldId primitive.ObjectID, workspaceId primitive.ObjectID) (*models.CustomField, error)
}

type UpdateCustomFieldRepository interface {
	Update(customFieldId primitive.ObjectID, customField *models.CustomField) (*models.CustomField, error)
}

type DeleteCustomFieldRepository interface {
	Delete(customFieldIds []primitive.ObjectID, workspaceId primitive.ObjectID) error
}
