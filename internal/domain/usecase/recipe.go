package usecase

import "github.com/anuntech/finance-backend/internal/domain/models"

type CreateRecipeRepository interface {
	Create(recipe models.Recipe) (*models.Recipe, error)
}
