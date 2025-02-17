package routes

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/setup/adapters"
	"github.com/anuntech/finance-backend/internal/setup/factory"
	"github.com/anuntech/finance-backend/internal/setup/middlewares"
	"go.mongodb.org/mongo-driver/mongo"
)

func CategoryRoutes(server *http.ServeMux, db *mongo.Database, workspaceDb *mongo.Database) {
	server.Handle("POST /category", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeCreateCategoryController(db)),
			workspaceDb,
		),
	))

	server.Handle("GET /category", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeGetCategorysController(db)),
			workspaceDb,
		),
	))

	server.Handle("POST /category/sub-category", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeCreateSubCategoryController(db)),
			workspaceDb,
		),
	))

	server.Handle("PUT /category/sub-category/{categoryId}/{subCategoryId}", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeUpdateSubCategoryController(db)),
			workspaceDb,
		),
	))

	server.Handle("DELETE /category/sub-category/{categoryId}/{subCategoryId}", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeDeleteSubCategoryController(db)),
			workspaceDb,
		),
	))

	server.Handle("DELETE /category/{categoryId}", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeDeleteCategoryController(db)),
			workspaceDb,
		),
	))

	server.Handle("GET /category/{categoryId}", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeGetCategoryByIdController(db)),
			workspaceDb,
		),
	))

	server.Handle("PUT /category/{categoryId}", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeUpdateCategoryController(db)),
			workspaceDb,
		),
	))

	server.Handle("POST /category/import", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeImportCategoryController(db)),
			workspaceDb,
		),
	))
}
