package config

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DBName = "sitor"
var UserCollection = "users"

var (
	clientInstance      *mongo.Client
	clientInstanceError error
	dbOnce              sync.Once
)

func getClient() *mongo.Client {
	dbOnce.Do(func() {
		_ = godotenv.Load(".env")
		mongoString := os.Getenv("MONGOSTRING")
		if mongoString == "" {
			log.Fatal("MONGOSTRING environment variable is not set")
		}
		clientInstance, clientInstanceError = mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoString))
		if clientInstanceError != nil {
			log.Fatalf("MongoConnect failed: %v", clientInstanceError)
		}
	})
	if clientInstance == nil {
		log.Fatal("MongoDB client is nil after connection attempt")
	}
	return clientInstance
}

// Helper to get default DB
func GetDB() *mongo.Database {
	client := getClient()
	if client == nil {
		log.Fatal("MongoDB client is not initialized")
	}
	return client.Database(DBName)
}
