package main

import (
	// model "example/be-eventeq/Models"
	"context"
	"fmt"
	"log"

	controller "github.com/zyqhpz/be-eventeq/Controllers"
	db "github.com/zyqhpz/be-eventeq/Database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var client *mongo.Client

func setupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	
	/* User */
	app.Get("/api/user", controller.GetUsers).Name("user.get")
	app.Post("/api/user/login/*", controller.LoginUser).Name("user.login")

	app.Get("/api/user/auth/", controller.LoginUserAuth).Name("user.login.auth")
	app.Post("/api/user/logout/", controller.LogoutUser).Name("user.logout")

	app.Post("/api/user/register/*", controller.RegisterUser).Name("user.register")
	app.Put("/api/user/update/:email", controller.UpdateUserById).Name("user.update")
	app.Put("/api/user/update/:email/password", controller.UpdateUserPasswordById).Name("user.updatePassword")

	/* Item */
	app.Get("/api/item", controller.GetItems).Name("item.get")
	app.Post("/api/item/create", controller.AddItem).Name("item.create")
	app.Get("/api/item/:id", controller.GetItemById).Name("item.getById")
	app.Get("/api/item/image/:id", controller.GetItemImageById).Name("item.getImage")

	app.Get("/api/planet", GetPlanets).Name("planet.get")

	app.Post("/user/test/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Login Success"})
	})
}

type Planet struct {
	ID        		primitive.ObjectID 	`bson:"_id,omitempty"`
	Name      		string             	`bson:"name,omitempty"`
	HasRings  		bool             	`bson:"hasRings,omitempty"`
	OrderFromSun  	int32              	`bson:"orderFromSun,omitempty"`
}


func GetPlanets(c *fiber.Ctx) error {
	// Set client options
	client, err  := db.ConnectDB()
	ctx := context.Background()
	defer client.Disconnect(ctx)

	// Get a handle to the "planets" collection
	collection := client.Database("sample_guides").Collection("planets")

	// Find all documents in the collection
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	// Iterate through the documents and print them
	var planets []Planet
	for cursor.Next(ctx) {
		var planet Planet
		if err := cursor.Decode(&planet); err != nil {
			log.Fatal(err)
		}
		planets = append(planets, planet)
	}

	// Print the results
	fmt.Printf("%d planets found:\n", len(planets))
	for _, planet := range planets {
		fmt.Printf("%s has %d from Sun\n", planet.Name, planet.OrderFromSun)
	}

	return c.JSON(planets)
}


func main() {
    app := fiber.New(fiber.Config{
		// EnablePrintRoutes: true,
		// DisableStartupMessage: true,
	})
	
	app.Use(cors.New(
		cors.Config{
			AllowCredentials: true,
		},
	))
	setupRoutes(app)
    app.Listen(":8080")
}