package factory

import (
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/account_repository"
	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/recipe_repository"
	controllers "github.com/anuntech/finance-backend/internal/presentation/controllers/recipe"
	"go.mongodb.org/mongo-driver/mongo"
)

func MakeCreateRecipeController(db *mongo.Database) *controllers.CreateRecipeController {
	createRecipe := recipe_repository.NewCreateRecipeRepository(db)
	findAccountById := account_repository.NewFindByIdMongoRepository(db)
	return controllers.NewCreateRecipeController(createRecipe, findAccountById)
}
