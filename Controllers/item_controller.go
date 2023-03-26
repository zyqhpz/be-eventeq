package api

import (
	"context"
	"log"

	db "github.com/zyqhpz/be-eventeq/Database"
	model "github.com/zyqhpz/be-eventeq/Models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gofiber/fiber/v2"
)

/*
	* Connect to the "items" collection
	@param client *mongo.Client
*/
func ConnectDBItems(client *mongo.Client) (*mongo.Collection) {
	// Get a handle to the "users" collection
	collection := client.Database("eventeq").Collection("items")
	return collection
}

/*
	* GET /api/items
	* Get all items
*/
func GetItems(c *fiber.Ctx) error {
	client, err  := db.ConnectDB()

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	defer client.Disconnect(ctx)
	collectionItems := ConnectDBItems(client)

	cursor, err := collectionItems.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	// Iterate through the documents and print them
	var items []model.Item
	for cursor.Next(ctx) {
		var item model.Item
		if err := cursor.Decode(&item); err != nil {
			log.Fatal(err)
		}
		items = append(items, item)
	}
	return c.JSON(items)
}