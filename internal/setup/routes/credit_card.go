package routes

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/setup/adapters"
	"github.com/anuntech/finance-backend/internal/setup/factory"
	"github.com/anuntech/finance-backend/internal/setup/middlewares"
	"go.mongodb.org/mongo-driver/mongo"
)

// CreditCardRoutes registers HTTP routes for credit card operations
func CreditCardRoutes(server *http.ServeMux, db *mongo.Database, workspaceDb *mongo.Database) {
	// Create a new credit card
	server.Handle("POST /credit-card", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeCreateCreditCardController(db)),
			workspaceDb,
		),
	))

	// Get all credit cards
	server.Handle("GET /credit-card", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeGetCreditCardsController(db)),
			workspaceDb,
		),
	))

	// Get a credit card by ID
	server.Handle("GET /credit-card/{creditCardId}", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeGetCreditCardByIdController(db)),
			workspaceDb,
		),
	))

	// Update a credit card
	server.Handle("PUT /credit-card/{creditCardId}", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeUpdateCreditCardController(db)),
			workspaceDb,
		),
	))

	// Delete credit cards
	server.Handle("DELETE /credit-card", middlewares.VerifyAccessToken(
		middlewares.IsAllowed(
			adapters.AdaptRoute(factory.MakeDeleteCreditCardController(db)),
			workspaceDb,
		),
	))
}
