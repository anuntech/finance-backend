package setup

import (
	"net/http"

	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"github.com/anuntech/finance-backend/internal/setup/config"
)

func Server() *http.ServeMux {
	mux := http.NewServeMux()

	db := helpers.MongoHelper()

	config.SetupRoutes(mux, db)

	return mux
}
