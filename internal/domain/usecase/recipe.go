package usecase

import (
	"github.com/anuntech/finance-backend/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateRecipeRepository interface {
	Create(recipe *models.Recipe) (*models.Recipe, error)
}

type FindRecipesByWorkspaceIdRepository interface {
	Find(workspaceId primitive.ObjectID) ([]models.Recipe, error)
}

type UpdateRecipeRepository interface {
	CreateSubCategory(subCategory *models.SubRecipeCategory, recipeId primitive.ObjectID, workspaceId primitive.ObjectID) (*models.SubRecipeCategory, error)
	DeleteSubCategory(recipeId primitive.ObjectID, subCategoryId primitive.ObjectID, workspaceId primitive.ObjectID) error
	UpdateRecipe(recipe *models.Recipe) error
}

type FindRecipeByIdRepository interface {
	Find(recipeId primitive.ObjectID, workspaceId primitive.ObjectID) (*models.Recipe, error)
}

type DeleteRecipeRepository interface {
	Delete(recipeId primitive.ObjectID, workspaceId primitive.ObjectID) error
}
