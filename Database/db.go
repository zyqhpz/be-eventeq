package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() (*mongo.Client, error) {

    password, exists := os.LookupEnv("MONGODB_PASSWORD")

    if !exists {
		log.Fatal("MONGODB_PASSWORD environment variable not set")
	}

    // Set client options
    clientOptions := options.Client().ApplyURI("mongodb+srv://zyqhpz:"+password+"@eventeq.obgaljj.mongodb.net/?retryWrites=true&w=majority")

    // Connect to MongoDB Atlas
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        return nil, err
    }

    // Check the connection
    err = client.Ping(ctx, nil)
    if err != nil {
        return nil, err
    }

    return client, nil
}
