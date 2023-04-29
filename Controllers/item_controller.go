package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
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

	type FileIDs struct {
		ID []primitive.ObjectID `bson:"_id"`
	}

	type Item struct {
		ID primitive.ObjectID `bson:"_id"`
		Name string `bson:"name"`
		Description string `bson:"description"`
		Price float64 `bson:"price"`
		Quantity int	`bson:"quantity"`
		Image []primitive.ObjectID `bson:"image"`
		OwnedBy primitive.ObjectID `bson:"ownedBy"`
		CreatedAt time.Time `bson:"createdAt"`
		UpdatedAt time.Time `bson:"updatedAt"`
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
	* GET /api/item/user/:id/
	* Get items by User ID
*/
func GetItemsByUserId(c *fiber.Ctx) error {

	type Item struct {
		ID primitive.ObjectID `bson:"_id"`
		Name string `bson:"name"`
		Description string `bson:"description"`
		Price float64 `bson:"price"`
		Quantity int	`bson:"quantity"`
		Image primitive.ObjectID `bson:"image"`
		OwnedBy primitive.ObjectID `bson:"ownedBy"`
		CreatedAt time.Time `bson:"createdAt"`
		UpdatedAt time.Time `bson:"updatedAt"`
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
	price, _ := strconv.ParseFloat(form.Value["price"][0], 64)
	quantity, _ := strconv.Atoi(form.Value["quantity"][0])
	id := form.Value["userID"][0]
	
	userID, err := primitive.ObjectIDFromHex(id)

	file, err := c.FormFile("image")
	if err != nil {
		return err
	}

	log.Println(form)
	log.Println(file)

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

	type Images struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	// create an array to store Images id
	var images []Images

	// append the image id to the array
	images = append(images, Images{ID: uploadStream.FileID.(primitive.ObjectID)})

	type Item struct {
		Name string `bson:"name"`
		Description string `bson:"description"`
		Price float64 `bson:"price"`
		Quantity int	`bson:"quantity"`
		Images []primitive.ObjectID `bson:"images"`
		OwnedBy primitive.ObjectID `bson:"ownedBy"`
		CreatedAt time.Time `bson:"created_at"`
		UpdatedAt time.Time `bson:"updated_at"`
	}

	item := Item{
		Name: name,
		Description: description,
		Price: price,
		Quantity: quantity,
		Images: images,
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

	// Return response
	return c.JSON(fiber.Map{
		"status": "success",
		"message": "Item created successfully",
		"item_id": res.InsertedID,
	})
}

/*
	* POST /api/item/add
	* Add an item
*/
func AddItemImages(c *fiber.Ctx) error {

	// Parse the multipart form data
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	count, _ := strconv.Atoi(form.Value["count"][0])

	// make files array
	var files[] *multipart.FileHeader

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
	
	for i := 0; i < count; i++ {
		imageCount := "images-" + strconv.Itoa(i)
		// log.Println(imageCount)
		
		file, err := c.FormFile(imageCount)
		if err != nil {
			// return err
		}

		log.Println(file.Filename)

		// log.Println(form.File[imageCount])

		files = append(files, file)

		// log.Println(file)
	}

	// // Upload each file to the GridFS
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

	type FileIDs struct {
		ID []primitive.ObjectID `bson:"_id"`
	}

	type Item struct {
		ID primitive.ObjectID `bson:"_id"`
		Images []primitive.ObjectID `bson:"images"`
		CreatedAt time.Time `bson:"created_at"`
		UpdatedAt time.Time `bson:"updated_at"`
	}

	
	item := Item{
		ID: primitive.NewObjectID(),
		Images: fileIDs,
		CreatedAt: util.GetCurrentTime(),
		UpdatedAt: util.GetCurrentTime(),
	}

	// // // Insert new item into database
	collectionItems := db.Collection("items")
	res, err := collectionItems.InsertOne(context.Background(), item)
	if err != nil {
		return err
	}

	log.Println("images uploaded successfully")

	// Return response
	return c.JSON(fiber.Map{
		"status": "success",
		"message": "Item created successfully",
		"item_id": res.InsertedID,
		"images": fileIDs,
	})
}

func ReadBlobs(c *fiber.Ctx) error {

		// Parse the multipart form data
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

		// Get the blob URLs from the form data
	blobs := form.Value["images"]

	log.Println(blobs)


	// Create a new multipart message
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	// Iterate over the blobs and add each as a new file part in the multipart message
	for i, blob := range blobs {
		// Create a new request with the blob URL as the body
		req, err := http.NewRequest(http.MethodGet, blob, nil)
		if err != nil {
			return err
		}

		// Add a Content-Disposition header to the request with the filename
		filename := fmt.Sprintf("image%d.jpg", i)
		req.Header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="images"; filename="%s"`, filename))

		// Send the request and read the response body as a multipart.FileHeader
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		// Add the file part to the multipart message
		part, err := writer.CreateFormFile("images", filename)
		if err != nil {
			return err
		}
		_, err = io.Copy(part, res.Body)
		if err != nil {
			return err
		} 
	}

	// Close the multipart message and set the Content-Type header
	err = writer.Close()
	if err != nil {
		return err
	}

	// Create a new request with the multipart message as the body
	req, err := http.NewRequest(http.MethodPost, "localhost:8080/api/item/createImage", buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request to the server
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
defer res.Body.Close()

// Return response
return c.JSON(fiber.Map{
	"status": "success",
	"message": "Item created successfully",
	// "item_id": res.InsertedID,
	// "images": fileIDs,
})
}