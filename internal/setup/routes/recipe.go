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

	// server.Handle("GET /recipe", middlewares.VerifyAccessToken(
	// 	middlewares.IsAllowed(
	// 		adapters.AdaptRoute(factory.MakeGetRecipesController(db)),
	// 		workspaceDb,
	// 	),
	// ))
}
