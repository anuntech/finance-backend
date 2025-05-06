package helpers

import (
	"context"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Timeout padrão para as operações do MongoDB
var Timeout = 30 * time.Second

var (
	dbConnections   = make(map[string]*mongo.Database)
	mongoClientMap  = make(map[string]*mongo.Client)
	connectionMutex sync.Mutex
)

// MongoHelper retorna uma conexão com o MongoDB
func MongoHelper(connectionUrl string, databaseName string) *mongo.Database {
	connectionKey := connectionUrl + ":" + databaseName

	// Verifica se já existe uma conexão com este banco de dados
	connectionMutex.Lock()
	if db, exists := dbConnections[connectionKey]; exists {
		connectionMutex.Unlock()
		return db
	}
	connectionMutex.Unlock()

	// Se não existir, cria uma nova conexão
	clientOptions := options.Client().ApplyURI(connectionUrl)

	// Configurações importantes para melhorar a performance
	clientOptions.SetMaxPoolSize(200)                   // Ajuste esse valor de acordo com sua necessidade
	clientOptions.SetMinPoolSize(20)                    // Mantém um número mínimo de conexões abertas
	clientOptions.SetMaxConnIdleTime(200 * time.Second) // Tempo máximo que uma conexão pode ficar inativa

	// Configuração de timeout para operações de conexão
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
		return nil
	}

	// Verifica se a conexão foi estabelecida com sucesso
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Error pinging MongoDB: %v", err)
		return nil
	}

	// Armazena a conexão para reutilização
	db := client.Database(databaseName)

	connectionMutex.Lock()
	dbConnections[connectionKey] = db
	mongoClientMap[connectionKey] = client
	connectionMutex.Unlock()

	log.Printf("Connected to MongoDB: %s", databaseName)

	return db
}

// DisconnectMongo fecha todas as conexões com o MongoDB
func DisconnectMongo() {
	connectionMutex.Lock()
	defer connectionMutex.Unlock()

	for _, client := range mongoClientMap {
		ctx, cancel := context.WithTimeout(context.Background(), Timeout)
		defer cancel()

		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		} else {
			log.Printf("Disconnected from MongoDB")
		}
	}

	// Limpa os mapas de conexão
	dbConnections = make(map[string]*mongo.Database)
	mongoClientMap = make(map[string]*mongo.Client)
}
