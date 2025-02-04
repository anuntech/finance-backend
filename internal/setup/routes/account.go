package routes

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/setup/adapters"
	"github.com/anuntech/finance-backend/internal/setup/factory"
	"github.com/anuntech/finance-backend/internal/setup/middlewares"
	"go.mongodb.org/mongo-driver/mongo"
)

func AccountRoutes(server *http.ServeMux, db *mongo.Database) {
	server.Handle("POST /account", middlewares.VerifyAccessToken(adapters.AdaptRoute(factory.MakeCreateAccountController(db))))
}
