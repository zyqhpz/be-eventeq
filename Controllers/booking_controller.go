package api

import (
	"context"
	"log"
	"time"

	db "github.com/zyqhpz/be-eventeq/Database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gofiber/fiber/v2"

	model "github.com/zyqhpz/be-eventeq/Models"
)

func ConnectDBBookings(client *mongo.Client) *mongo.Collection {
	// Get a handle to the "users" collection
	collection := client.Database("eventeq").Collection("bookings")
	return collection
}

func GetItemDetailsForBooking(c *fiber.Ctx) error {
	// get id from params
	ownerId := c.Params("ownerId")

	// convert id to primitive.ObjectID
	oid, err := primitive.ObjectIDFromHex(ownerId)

	type Item struct {
		ID          primitive.ObjectID `bson:"_id,omitempty"`
		Name        string             `bson:"name"`
		Description string             `bson:"description"`
		Price       float64            `bson:"price"`
		Quantity    int                `bson:"quantity"`
		Images      []string           `bson:"images"`
	}

	type Body struct {
		ID          primitive.ObjectID `bson:"_id,omitempty"`
		FirstName   string             `bson:"first_name"`
		LastName    string             `bson:"last_name"`
		Email       string             `bson:"email"`
		// PhoneNumber string             `bson:"phone_number"`
		Location    model.Location     `bson:"location"`
		Items       []Item             `bson:"items"`
	}

	client, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	// Select the `items` collection from the database
	itemsCollection := ConnectDBItems(client)
	ctx := context.Background()

	// Query for the Item document and filter by the User ID in ownedBy
	cursor, err := itemsCollection.Find(ctx, bson.M{"ownedBy": oid})
	if err != nil {
		// Return an error response if the document is not found
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Item not found",
			})
		}
		// Return an error response if there is a database error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get item from database",
		})
	}
	defer cursor.Close(ctx)

	// Iterate through the documents and print them
	var items []Item
	for cursor.Next(ctx) {
		var item Item
		if err := cursor.Decode(&item); err != nil {
			log.Fatal(err)
		}
		items = append(items, item)
	}

	defer client.Disconnect(ctx)
	usersCollection := ConnectDBUsers(client)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var body Body
	body.Items = items

	err = usersCollection.FindOne(ctx, bson.M{"_id": oid}).Decode(&body)
	if err != nil {
		log.Fatal(err)
	}

	return c.JSON(body)
}