package routes

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/setup/adapters"
	"github.com/anuntech/finance-backend/internal/setup/factory"
	"github.com/anuntech/finance-backend/internal/setup/middlewares"
	"go.mongodb.org/mongo-driver/mongo"
)

func TransactionRoutes(server *http.ServeMux, db *mongo.Database, workspaceDb *mongo.Database) {
	server.Handle("POST /transaction", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeCreateTransactionController(workspaceDb, db)),
			workspaceDb,
		),
	))

	server.Handle("GET /transaction", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeGetTransactionController(workspaceDb, db)),
			workspaceDb,
		),
	))

	server.Handle("PUT /transaction/{id}", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeUpdateTransactionController(workspaceDb, db)),
			workspaceDb,
		),
	))

	server.Handle("DELETE /transaction", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeDeleteTransactionController(db)),
			workspaceDb,
		),
	))

	server.Handle("POST /transaction/edit", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeCreateEditTransactionController(workspaceDb, db)),
			workspaceDb,
		),
	))
}
