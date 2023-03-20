package api

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	db "github.com/zyqhpz/be-eventeq/Database"
	model "github.com/zyqhpz/be-eventeq/Models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gofiber/fiber/v2"
)

/*
	* Connect to the "users" collection
	@param client *mongo.Client
*/
func ConnectDBUsers(client *mongo.Client) (*mongo.Collection) {
	// Get a handle to the "users" collection
	collection := client.Database("eventeq").Collection("users")

	return collection
}

/*
	* GET /api/users
	* Get all users
*/
func GetUsers(c *fiber.Ctx) error {
	client, err  := db.ConnectDB()
	ctx := context.Background()
	defer client.Disconnect(ctx)
	collectionUsers := ConnectDBUsers(client)

	cursor, err := collectionUsers.Find(ctx, bson.M{})
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

/*
	* POST /api/user/login
	* Login a user
*/
func LoginUser(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	client, _  := db.ConnectDB()
	ctx := context.Background()
	defer client.Disconnect(ctx)

	collection := ConnectDBUsers(client)

	// Define a filter to find the user with the given username and password
    filter := bson.M{"username": username, "password": password}

    // Count the number of documents that match the filter
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    count, _ := collection.CountDocuments(ctx, filter)

    // Return true if a user with the given username and password was found, false otherwise
    if (count > 0) {
		fmt.Println("Username: ", username, "Login Success")
		return c.JSON(fiber.Map{"status": "success", "message": "Login Success " + username})
	}
	fmt.Println("Username: ", username, "Login Failed")
	return c.JSON(fiber.Map{"status": "failed", "message": "Login Failed " + username})
}

/*
	* POST /api/user/register
	* Register a new user
*/
func RegisterUser(c *fiber.Ctx) error {
	first_name := c.FormValue("first_name")
	last_name := c.FormValue("last_name")
	username := c.FormValue("username")
	password := c.FormValue("password")
	email := c.FormValue("email")

	// hash the password in sha256
	password = fmt.Sprintf("%x", sha256.Sum256([]byte(password)))

	client, _  := db.ConnectDB()

	collection := ConnectDBUsers(client)

	// Define a filter to find the user with the given username and email, if any exist return error message
	filterUsername := bson.M{"username": username}
	filterEmail := bson.M{"email": email}

	// Count the number of documents that match the filter
	ctx := context.Background()
	countUsername, _ := collection.CountDocuments(ctx, filterUsername)
	countEmail, _ := collection.CountDocuments(ctx, filterEmail)

	// Return error if a user with the given username was found
	if (countUsername > 0) {
		return c.JSON(fiber.Map{"status": "failed username", "message": "Username already exists"})
	} else if (countEmail > 0) {
		return c.JSON(fiber.Map{"status": "failed email", "message": "Email already exists"})
	} else {
		// Create a new user
		user := model.User{
			FirstName: first_name,
			LastName: last_name,
			Username: username,
			Password: password,
			Email: email,
		}

		// Insert the new user into the database
		ctx := context.Background()
		result, err := collection.InsertOne(ctx, user)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Print("Inserted ID: ", result.InsertedID)

		return c.JSON(fiber.Map{"status": "success", "message": "User created successfully"})
	}
}

