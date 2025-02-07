package factory

import (
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/account_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/recipe_repository"
	controllers "github.com/anuntech/finance-backend/internal/presentation/controllers/recipe"
	controllers_recipe "github.com/anuntech/finance-backend/internal/presentation/controllers/recipe/sub_category"
	"go.mongodb.org/mongo-driver/mongo"
)

func MakeCreateRecipeController(db *mongo.Database) *controllers.CreateRecipeController {
	createRecipe := recipe_repository.NewCreateRecipeRepository(db)
	findAccountById := account_repository.NewFindByIdMongoRepository(db)
	return controllers.NewCreateRecipeController(createRecipe, findAccountById)
}

func MakeGetRecipesController(db *mongo.Database) *controllers.GetRecipesController {
	findRecipesByWorkspaceId := recipe_repository.NewFindRecipesByWorkspaceIdRepository(db)
	return controllers.NewGetRecipesController(findRecipesByWorkspaceId)
}

func MakeCreateSubCategoryController(db *mongo.Database) *controllers_recipe.CreateSubCategoryController {
	updateRecipe := recipe_repository.NewUpdateRecipeRepository(db)
	findRecipeById := recipe_repository.NewFindRecipeByIdRepository(db)
	return controllers_recipe.NewCreateSubCategoryController(updateRecipe, findRecipeById)
}
