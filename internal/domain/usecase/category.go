package usecase

import (
	"github.com/anuntech/finance-backend/internal/domain/models"
	presentationHelpers "github.com/anuntech/finance-backend/internal/presentation/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateCategoryRepository interface {
	Create(category *models.Category) (*models.Category, error)
}

type FindCategoriesRepository interface {
	Find(globalFilters *presentationHelpers.GlobalFilterParams) ([]models.Category, error)
}

type FindCategoryByNameAndWorkspaceIdRepository interface {
	FindByNameAndWorkspaceId(name string, workspaceId primitive.ObjectID) (*models.Category, error)
}

type UpdateCategoryRepository interface {
	CreateSubCategories(subCategories []models.SubCategoryCategory, categoryId primitive.ObjectID, workspaceId primitive.ObjectID) ([]models.SubCategoryCategory, error)
	CreateSubCategory(subCategory *models.SubCategoryCategory, categoryId primitive.ObjectID, workspaceId primitive.ObjectID) (*models.SubCategoryCategory, error)
	DeleteSubCategory(subCategoryIds []primitive.ObjectID, categoryId primitive.ObjectID, workspaceId primitive.ObjectID) error
	UpdateCategory(category *models.Category) error
	UpdateSubCategory(subCategory *models.SubCategoryCategory, categoryId primitive.ObjectID, subCategoryId primitive.ObjectID, workspaceId primitive.ObjectID) error
}

type FindCategoryByIdRepository interface {
	Find(categoryId primitive.ObjectID, workspaceId primitive.ObjectID) (*models.Category, error)
}

type DeleteCategoryRepository interface {
	Delete(categoryIds []primitive.ObjectID, workspaceId primitive.ObjectID) error
}

type ImportCategoriesRepository interface {
	Import(categories []models.Category, workspaceId primitive.ObjectID) ([]models.Category, error)
}
