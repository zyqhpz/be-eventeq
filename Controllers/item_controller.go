package api

import (
	"bytes"
	"context"
	"io"
	"log"
	"mime/multipart"
	"strconv"
	"time"

	db "github.com/zyqhpz/be-eventeq/Database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"

	util "github.com/zyqhpz/be-eventeq/util"

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
	ID primitive.ObjectID `bson:"_id"`
	Name string `bson:"name"`
	Description string `bson:"description"`
	Price float64 `bson:"price"`
	Quantity int	`bson:"quantity"`
	Image primitive.ObjectID `bson:"image"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

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

	type Item struct {
		ID 			primitive.ObjectID 		`bson:"_id"`
		Name 		string 					`bson:"name"`
		Description string 					`bson:"description"`
		Category	string 					`bson:"category"`
		Price 		float64 				`bson:"price"`
		Quantity 	int						`bson:"quantity"`
		Images 		[]primitive.ObjectID 	`bson:"images"`
		OwnedBy 	primitive.ObjectID 		`bson:"ownedBy"`
		CreatedAt 	time.Time 				`bson:"created_at"`
		UpdatedAt 	time.Time 				`bson:"updated_at"`
	}

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
	var items []Item
	for cursor.Next(ctx) {
		var item Item
		if err := cursor.Decode(&item); err != nil {
			log.Fatal(err)
		}
		items = append(items, item)
	}
	return c.JSON(items)
}

/*
	* GET /api/item/withUser
	* Get all items with user
*/
func GetItemsWithUser(c *fiber.Ctx) error {

	type User struct {
		ID primitive.ObjectID `bson:"_id"`
		FirstName string `bson:"first_name"`
		LastName string `bson:"last_name"`
		IsAvatarImageSet bool `bson:"isAvatarImageSet"`
	}

	type Data struct {
		ID          primitive.ObjectID `bson:"_id"`
		Name        string             `bson:"name"`
		Description string             `bson:"description"`
		Category    string             `bson:"category"`
		Price       float64            `bson:"price"`
		Quantity    int                `bson:"quantity"`
		Images      []primitive.ObjectID `bson:"images"`
		OwnedBy     User
		CreatedAt   time.Time          `bson:"created_at"`
		UpdatedAt   time.Time          `bson:"updated_at"`
	}

	client, err  := db.ConnectDB()

	if err != nil {
		log.Fatal(err)
	}

	// create a pipeline for the aggregation
	pipeline := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "ownedBy",
				"foreignField": "_id",
				"as":           "ownedBy",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$ownedBy",
				"preserveNullAndEmptyArrays": true,
			},
		},
	}

	ctx := context.Background()
	defer client.Disconnect(ctx)
	collectionItems := ConnectDBItems(client)

	cursor, err := collectionItems.Aggregate(ctx, pipeline)

	if err != nil {
		log.Fatal(err)
	}

	defer cursor.Close(ctx)

	// Iterate through the documents and print them
	var items []Data
	for cursor.Next(ctx) {
		var item Data
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
	ctx := context.Background()

	// Select the `items` collection from the database
	itemsCollection := ConnectDBItems(client)

	type User struct {
		ID primitive.ObjectID `bson:"_id"`
		FirstName string `bson:"first_name"`
		LastName string `bson:"last_name"`
		IsAvatarImageSet bool `bson:"isAvatarImageSet"`
	}

	type Data struct {
		ID          primitive.ObjectID `bson:"_id"`
		Name        string             `bson:"name"`
		Description string             `bson:"description"`
		Price       float64            `bson:"price"`
		Quantity    int                `bson:"quantity"`
		Images      []primitive.ObjectID `bson:"images"`
		OwnedBy     User
		CreatedAt   time.Time `bson:"created_at"`
		UpdatedAt   time.Time `bson:"updated_at"`
	}

	// create a pipeline for the aggregation
	pipeline := []bson.M{
		{
			"$match": bson.M{"_id": id},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "ownedBy",
				"foreignField": "_id",
				"as":           "ownedBy",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$ownedBy",
				"preserveNullAndEmptyArrays": true,
			},
		},
	}

	// execute the aggregation pipeline
	cursor, err := itemsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		// Return an error response if the aggregation fails
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get item from database",
		})
	}

	// loop through the results and decode them into Data objects
	var results []Data
	for cursor.Next(ctx) {
		var data Data
		err := cursor.Decode(&data)
		if err != nil {
			// Return an error response if the decoding fails
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to get item from database",
			})
		}
		results = append(results, data)
	}

	// handle any errors that occurred during the cursor iteration
	if err := cursor.Err(); err != nil {
		// Return an error response if the iteration fails
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get item from database",
		})
	}

	// send the results as a JSON response
	return c.JSON(results[0])
}

/*
	* GET /api/item/user/:id/
	* Get items by User ID
*/
func GetItemsByUserId(c *fiber.Ctx) error {

	type Item struct {
		ID 			primitive.ObjectID 		`bson:"_id"`
		Name 		string 					`bson:"name"`
		Description string 					`bson:"description"`
		Category	string 					`bson:"category"`
		Price 		float64 				`bson:"price"`
		Quantity 	int						`bson:"quantity"`
		Images 		[]primitive.ObjectID 	`bson:"images"`
		OwnedBy 	primitive.ObjectID 		`bson:"ownedBy"`
		CreatedAt 	time.Time 				`bson:"created_at"`
		UpdatedAt 	time.Time 				`bson:"updated_at"`
	}

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
	ctx := context.Background()

	// Query for the Item document and filter by the User ID in ownedBy
	cursor, err := itemsCollection.Find(ctx, bson.M{"ownedBy": id})
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
	return c.JSON(items)
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
	category := form.Value["category"][0]
	price, _ := strconv.ParseFloat(form.Value["price"][0], 64)
	quantity, _ := strconv.Atoi(form.Value["quantity"][0])
	id := form.Value["userID"][0]
	imagesCount, _ := strconv.Atoi(form.Value["imagesCount"][0])
	
	userID, err := primitive.ObjectIDFromHex(id)

	// make files array
	var files[] *multipart.FileHeader
	for i := 0; i < imagesCount; i++ {
		imageCount := "images-" + strconv.Itoa(i)
	
		file, err := c.FormFile(imageCount)
		if err != nil {
			return err
		}

		files = append(files, file)
	}

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

	// create an array to store Images ids
	var fileIDs []primitive.ObjectID
	for _, file := range files {

		// Open the file
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		// Create a new upload stream
		uploadStream, err := bucket.OpenUploadStream(file.Filename)
		if err != nil {
			return err
		}
		defer uploadStream.Close()

		// Copy the file data to the upload stream
		_, err = io.Copy(uploadStream, src)
		if err != nil {
			return err
		}

		// Get the ID of the uploaded file
		fileID := uploadStream.FileID
		fileIDs = append(fileIDs, fileID.(primitive.ObjectID))

		log.Println("file " + file.Filename + " uploaded successfully")
	}

	type Item struct {
		ID 			primitive.ObjectID 		`bson:"_id,omitempty"`
		Name 		string 					`bson:"name"`
		Description string 					`bson:"description"`
		Category 	string 					`bson:"category"`
		Price 		float64 				`bson:"price"`
		Quantity 	int						`bson:"quantity"`
		Images 		[]primitive.ObjectID 	`bson:"images"`
		OwnedBy 	primitive.ObjectID 		`bson:"ownedBy"`
		CreatedAt 	time.Time 				`bson:"created_at"`
		UpdatedAt 	time.Time 				`bson:"updated_at"`
	}

	item := Item{
		ID: primitive.NewObjectID(),
		Name: name,
		Description: description,
		Category: category,
		Price: price,
		Quantity: quantity,
		Images: fileIDs,
		OwnedBy: userID,
		CreatedAt: util.GetCurrentTime(),
		UpdatedAt: util.GetCurrentTime(),
	}

	// Insert new item into database
	collectionItems := db.Collection("items")
	res, err := collectionItems.InsertOne(context.Background(), item)
	if err != nil {
		return err
	}

	log.Println("[Item] Inserted new item: ", res.InsertedID)

	// Return response
	return c.JSON(fiber.Map{
		"status": "success",
		"message": "Item created successfully",
		"item_id": res.InsertedID,
	})
}