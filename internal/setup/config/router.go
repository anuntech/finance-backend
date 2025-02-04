package config

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/setup/routes"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(server *http.ServeMux, db *mongo.Database) {
	routes.AccountRoutes(server, db)
}
