package routes

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/setup/adapters"
	"github.com/anuntech/finance-backend/internal/setup/factory"
	"github.com/anuntech/finance-backend/internal/setup/middlewares"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateChat(server *http.ServeMux, db *mongo.Database) {
	server.Handle("POST /chat", middlewares.VerifyAccessToken(adapters.AdaptRoute(factory.MakeCreateChatController(db))))
}
