package routes

import (
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

func AccountRoutes(server *http.ServeMux, db *mongo.Database) {
	// server.Handle("POST /account", middlewares.VerifyAccessToken(adapters.AdaptRoute()))
}
