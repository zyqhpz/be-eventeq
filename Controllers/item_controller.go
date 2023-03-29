package api

import (
	"context"
	"io"
	"log"
	"strconv"
	"time"

	db "github.com/zyqhpz/be-eventeq/Database"
	model "github.com/zyqhpz/be-eventeq/Models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gofiber/fiber/v2"
)

type CreateNewItemRequest struct {
	Name string `bson:"name"`
	Description string `bson:"description"`
	Price float64 `bson:"price"`
	Quantity int	`bson:"quantity"`
	Image primitive.ObjectID `bson:"image"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}
// type CreateNewItemRequest struct {
// 	Name string `json:"name"`
// 	Description string `json:"description"`
// 	Price float64 `json:"price"`
// 	Quantity int `json:"quantity"`
// 	Image string `json:"image"`
// }

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

/*
	* POST /api/item/add
	* Add an item
*/
func AddItem(c *fiber.Ctx) error {

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	// Get all files from form

	name := form.Value["name"][0]
	description := form.Value["description"][0]
	price, _ := strconv.ParseFloat(form.Value["price"][0], 64)
	quantity, _ := strconv.Atoi(form.Value["quantity"][0])

	file, err := c.FormFile("image")
	if err != nil {
		return err
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	client, err  := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	defer client.Disconnect(ctx)

	db := client.Database("eventeq")
	bucket, err := gridfs.NewBucket(db, options.GridFSBucket().SetName("images"))

	if err != nil {
		log.Fatal(err)
	}

	// Upload file to gridfs
	uploadStream, err := bucket.OpenUploadStream(file.Filename)
	if err != nil {
		return err
	}
	defer uploadStream.Close()

	if _, err := io.Copy(uploadStream, src); err != nil {
		return err
	}


	// collectionItems := ConnectDBItems(client)

	item := CreateNewItemRequest{
		Name: name,
		Description: description,
		Price: price,
		Quantity: quantity,
		Image: uploadStream.FileID.(primitive.ObjectID),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}


	// Insert new item into database
	collectionItems := db.Collection("items")
	res, err := collectionItems.InsertOne(context.Background(), item)
	if err != nil {
		return err
	}


	// var item model.Item
	// if err := c.BodyParser(&item); err != nil {
	// 	return err
	// }

	// image := 

	// // Insert a single document
	// insertResult, err := collectionItems.InsertOne(ctx, item)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// return c.JSON(insertResult)

		// Return response
	return c.JSON(fiber.Map{
		"message": "Item created successfully",
		"item_id": res.InsertedID,
	})
}