package usecase

import "github.com/anuntech/finance-backend/internal/domain/models"

type CreateRecipeRepository interface {
	Create(recipe models.Recipe) (*models.Recipe, error)
}

type FindRecipesByWorkspaceIdRepository interface {
	Find(workspaceId string) ([]models.Recipe, error)
}

type UpdateRecipeRepository interface {
	CreateSubCategory(subCategory models.SubRecipeCategory, recipeId string, workspaceId string) (*models.SubRecipeCategory, error)
	DeleteSubCategory(recipeId string, subCategoryId string, workspaceId string) error
}

type FindRecipeByIdRepository interface {
	Find(recipeId string, workspaceId string) (*models.Recipe, error)
}
