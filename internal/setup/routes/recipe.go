package routes

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/setup/adapters"
	"github.com/anuntech/finance-backend/internal/setup/factory"
	"github.com/anuntech/finance-backend/internal/setup/middlewares"
	"go.mongodb.org/mongo-driver/mongo"
)

func RecipeRoutes(server *http.ServeMux, db *mongo.Database, workspaceDb *mongo.Database) {
	server.Handle("POST /recipe", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeCreateRecipeController(db)),
			workspaceDb,
		),
	))

	server.Handle("GET /recipe", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeGetRecipesController(db)),
			workspaceDb,
		),
	))

	server.Handle("POST /recipe/sub-category", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeCreateSubCategoryController(db)),
			workspaceDb,
		),
	))

	server.Handle("DELETE /recipe/sub-category/{recipeId}/{subCategoryId}", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeDeleteSubCategoryController(db)),
			workspaceDb,
		),
	))

	server.Handle("DELETE /recipe/{recipeId}", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeDeleteRecipeController(db)),
			workspaceDb,
		),
	))

	server.Handle("GET /recipe/{recipeId}", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeGetRecipeByIdController(db)),
			workspaceDb,
		),
	))

	server.Handle("PUT /recipe/{recipeId}", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeUpdateRecipeController(db)),
			workspaceDb,
		),
	))
}
