package config

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/setup/routes"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(server *http.ServeMux, db *mongo.Database, workspaceDb *mongo.Database) {
	apiServer := http.NewServeMux()
	routes.AccountRoutes(apiServer, db, workspaceDb)
	routes.CategoryRoutes(apiServer, db, workspaceDb)
	routes.BankRoutes(apiServer, db, workspaceDb)
	routes.TransactionRoutes(apiServer, db, workspaceDb)
	routes.CustomFieldRoutes(apiServer, db, workspaceDb)
	routes.CreditCardRoutes(apiServer, db, workspaceDb)

	server.Handle("/api/", http.StripPrefix("/api", apiServer))
}
