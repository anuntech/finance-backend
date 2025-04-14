package routes

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/setup/adapters"
	"github.com/anuntech/finance-backend/internal/setup/factory"
	"github.com/anuntech/finance-backend/internal/setup/middlewares"
	"go.mongodb.org/mongo-driver/mongo"
)

func AccountRoutes(server *http.ServeMux, db *mongo.Database, workspaceDb *mongo.Database) {
	server.Handle("POST /account", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeCreateAccountController(db)),
			workspaceDb,
		),
	))

	server.Handle("GET /account", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeGetAccountsController(db)),
			workspaceDb,
		),
	))

	server.Handle("GET /account/{id}", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeGetAccountByIdController(db)),
			workspaceDb,
		),
	))

	server.Handle("DELETE /account", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeDeleteAccountController(db)),
			workspaceDb,
		),
	))

	server.Handle("POST /account/import", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeImportAccountController(db)),
			workspaceDb,
		),
	))

	server.Handle("PUT /account/{id}", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeUpdateAccountController(db)),
			workspaceDb,
		),
	))

	server.Handle("POST /account/transfer", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeTransferenceAccountController(db)),
			workspaceDb,
		),
	))
}
