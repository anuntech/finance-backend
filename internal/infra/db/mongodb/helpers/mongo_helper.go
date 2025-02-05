package helpers

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Ctx = context.TODO()

func MongoHelper(URI string, databaseName string) *mongo.Database {
	clientOptions := options.Client().ApplyURI(URI)
	client, err := mongo.Connect(Ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(Ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("MongoDB connection established with database", databaseName)

	return client.Database(databaseName)
}
