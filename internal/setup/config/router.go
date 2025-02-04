package config

import (
	"net/http"

	"github.com/willchat-ofc/api-willchat-golang/internal/setup/routes"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(server *http.ServeMux, db *mongo.Database) {
	routes.CreateChat(server, db)
}
