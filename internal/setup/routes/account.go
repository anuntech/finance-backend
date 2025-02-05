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
}
