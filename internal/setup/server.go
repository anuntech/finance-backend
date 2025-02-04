package setup

import (
	"net/http"

	"github.com/willchat-ofc/api-willchat-golang/internal/infra/db/mongodb/helpers"
	"github.com/willchat-ofc/api-willchat-golang/internal/setup/config"
)

func Server() *http.ServeMux {
	mux := http.NewServeMux()

	db := helpers.MongoHelper()

	config.SetupRoutes(mux, db)

	return mux
}
