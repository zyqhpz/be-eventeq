package api

import (
	"context"
	"fmt"
	"log"
	"time"

	db "github.com/zyqhpz/be-eventeq/Database"
	model "github.com/zyqhpz/be-eventeq/Models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gofiber/fiber/v2"
)

func ConnectDBUsers() (*mongo.Collection, error) {
	client, err  := db.ConnectDB()
	ctx := context.Background()
	defer client.Disconnect(ctx)

	// Get a handle to the "users" collection
	collection := client.Database("eventeq").Collection("users")

	return collection, err
}

func LoginUser(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	client, err  := db.ConnectDB()
	ctx := context.Background()
	defer client.Disconnect(ctx)

	collection := client.Database("eventeq").Collection("users")

	// Define a filter to find the user with the given username and password
    filter := bson.M{"username": username, "password": password}

    // Count the number of documents that match the filter
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    count, err := collection.CountDocuments(ctx, filter)
    if err != nil {
        return c.JSON(fiber.Map{"message": "Login Failed " + username})
    }

    // Return true if a user with the given username and password was found, false otherwise
    if (count > 0) {
		fmt.Println("Username: ", username, "Login Success")
		return c.JSON(fiber.Map{"message": "Login Success"})
	} else {
		return c.JSON(fiber.Map{"message": "Login Failed " + username})
	}
}

func GetUsers(c *fiber.Ctx) error {

	ctx := context.Background()

	// collection, _ := ConnectDBUsers()

	client, err  := db.ConnectDB()
	collection := client.Database("eventeq").Collection("users")

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	// Iterate through the documents and print them
	var users []model.User
	for cursor.Next(ctx) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}

	return c.JSON(users)
}