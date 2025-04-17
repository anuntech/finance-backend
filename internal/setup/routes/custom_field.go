package routes

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/setup/adapters"
	"github.com/anuntech/finance-backend/internal/setup/factory"
	"github.com/anuntech/finance-backend/internal/setup/middlewares"
	"go.mongodb.org/mongo-driver/mongo"
)

func CustomFieldRoutes(server *http.ServeMux, db *mongo.Database, workspaceDb *mongo.Database) {
	// Create a new custom field
	server.Handle("POST /custom-field", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeCreateCustomFieldController(db)),
			workspaceDb,
		),
	))

	// Get all custom fields
	server.Handle("GET /custom-field", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeGetCustomFieldsController(db)),
			workspaceDb,
		),
	))

	// Get a specific custom field by ID
	server.Handle("GET /custom-field/{customFieldId}", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeGetCustomFieldByIdController(db)),
			workspaceDb,
		),
	))

	// Update a custom field
	server.Handle("PUT /custom-field/{customFieldId}", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeUpdateCustomFieldController(db)),
			workspaceDb,
		),
	))

	// Delete custom fields
	server.Handle("DELETE /custom-field", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeDeleteCustomFieldController(db)),
			workspaceDb,
		),
	))
}
