package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// make const PASSWORD variable to store password string
const PASSWORD = "LpyaTun3O4AgHXSG"

func ConnectDB() (*mongo.Client, error) {
    // Set client options
    clientOptions := options.Client().ApplyURI("mongodb+srv://zyqhpz:"+PASSWORD+"@eventeq.obgaljj.mongodb.net/?retryWrites=true&w=majority")

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
