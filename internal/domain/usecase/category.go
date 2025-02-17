package usecase

import (
	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateCategoryRepository interface {
	Create(category *models.Category) (*models.Category, error)
}

type FindCategorysByWorkspaceIdRepository interface {
	Find(workspaceId primitive.ObjectID, categoryType string) ([]models.Category, error)
}

type UpdateCategoryRepository interface {
	CreateSubCategory(subCategory *models.SubCategoryCategory, categoryId primitive.ObjectID, workspaceId primitive.ObjectID) (*models.SubCategoryCategory, error)
	DeleteSubCategory(categoryId primitive.ObjectID, subCategoryId primitive.ObjectID, workspaceId primitive.ObjectID) error
	UpdateCategory(category *models.Category) error
	UpdateSubCategory(subCategory *models.SubCategoryCategory, categoryId primitive.ObjectID, subCategoryId primitive.ObjectID, workspaceId primitive.ObjectID) error
}

type FindCategoryByIdRepository interface {
	Find(categoryId primitive.ObjectID, workspaceId primitive.ObjectID) (*models.Category, error)
}

type DeleteCategoryRepository interface {
	Delete(categoryId primitive.ObjectID, workspaceId primitive.ObjectID) error
}

type ImportCategoriesRepository interface {
	Import(categories []models.Category, workspaceId primitive.ObjectID) ([]models.Category, error)
}
