package config

import (
	"context"
	"fmt"
	"os"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DBName = "sitor"
var UserCollection = "users"
var MongoString string = os.Getenv("MONGOSTRING")

var (
	clientInstance      *mongo.Client
	clientInstanceError error
	dbOnce              sync.Once
)

func getClient() *mongo.Client {
	dbOnce.Do(func() {
		clientInstance, clientInstanceError = mongo.Connect(context.TODO(), options.Client().ApplyURI(MongoString))
		if clientInstanceError != nil {
			fmt.Printf("MongoConnect: %v\n", clientInstanceError)
		}
	})
	return clientInstance
}

// Helper to get default DB
func GetDB() *mongo.Database {
	return getClient().Database(DBName)
}
