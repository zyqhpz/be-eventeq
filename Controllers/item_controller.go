package api

import (
	"bytes"
	"context"
	"io"
	"log"
	"strconv"
	"time"

	db "github.com/zyqhpz/be-eventeq/Database"
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

type ItemDetailsRequest struct {
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
	var items []CreateNewItemRequest
	for cursor.Next(ctx) {
		var item CreateNewItemRequest
		if err := cursor.Decode(&item); err != nil {
			log.Fatal(err)
		}
		items = append(items, item)
	}
	return c.JSON(items)
}

/*
	* GET /api/item/:id
	* Get an item by id
*/
func GetItemById(c *fiber.Ctx) error {
	// Retrieve the `id` parameter from the request URL
	idParam := c.Params("id")

	// Convert the `id` parameter to a MongoDB `ObjectID`
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		// Return an error response if the `id` parameter is not a valid `ObjectID`
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid item ID",
		})
	}

	// Connect to the MongoDB database
	client, err := db.ConnectDB()
	if err != nil {
		// Return an error response if the database connection fails
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to connect to database",
		})
	}
	defer client.Disconnect(context.Background())

	// Select the `items` collection from the database
	itemsCollection := ConnectDBItems(client)

	// Query for the `ItemDetailsRequest` document with the specified `id`
	var item ItemDetailsRequest
	err = itemsCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&item)
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

	// Return the retrieved `ItemDetailsRequest` document as a JSON response
	return c.JSON(item)
}

/*
	* GET /api/item/image/:id
	* Get an image item by id
*/
func GetItemImageById(c *fiber.Ctx) error {
	id := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(id)

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

	// downloadStream, err := bucket.OpenDownloadStreamByName(objectID.Hex())
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer downloadStream.Close()

	// // Copy the file to the response
	// if _, err = io.Copy(c, downloadStream); err != nil {
	// 	log.Fatal(err)
	// }

	// return nil

		// Open a download stream for the file with the given ID
	downloadStream, err := bucket.OpenDownloadStream(objectID)
	if err != nil {
		return c.Status(404).SendString("Image not found")
	}
	defer downloadStream.Close()

	// Read the contents of the file into a byte buffer
	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, downloadStream)
	if err != nil {
		return c.Status(500).SendString("Error reading image data")
	}

	// Return the image data
	return c.Send(buffer.Bytes())
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