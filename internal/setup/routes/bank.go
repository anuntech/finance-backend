package routes

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/setup/adapters"
	"github.com/anuntech/finance-backend/internal/setup/factory"
	"github.com/anuntech/finance-backend/internal/setup/middlewares"
	"go.mongodb.org/mongo-driver/mongo"
)

func BankRoutes(server *http.ServeMux, db *mongo.Database, workspaceDb *mongo.Database) {
	server.Handle("GET /bank", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			middlewares.AllowCacheHeader(adapters.AdaptRoute(factory.MakeGetBanksController(db))),
			workspaceDb,
		),
	))

	server.Handle("GET /bank/{id}", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			middlewares.AllowCacheHeader(adapters.AdaptRoute(factory.MakeGetBankByIdController(db))),
			workspaceDb,
		),
	))

	server.Handle("GET /bank/search", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			middlewares.AllowCacheHeader(adapters.AdaptRoute(factory.MakeGetBankByNameController(db))),
			workspaceDb,
		),
	))
}
