package setup

import (
	"net/http"
	"os"

	"github.com/anuntech/finance-backend/internal/infra/db/mongodb/helpers"
	"github.com/anuntech/finance-backend/internal/setup/config"
	"go.mongodb.org/mongo-driver/mongo"
)

func Server() *http.ServeMux {
	mux := http.NewServeMux()

	dbChan := make(chan *mongo.Database)
	workspaceDbChan := make(chan *mongo.Database)

	go func() {
		dbChan <- helpers.MongoHelper(os.Getenv("MONGO_URL"), "finance")
	}()

	go func() {
		workspaceDbChan <- helpers.MongoHelper(os.Getenv("WORKSPACE_MONGO_URL"), "test")
	}()

	db := <-dbChan
	workspaceDb := <-workspaceDbChan

	config.SetupRoutes(mux, db, workspaceDb)

	return mux
}
