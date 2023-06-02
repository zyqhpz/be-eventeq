package api

import (
	// "bytes"
	"context"
	// "io"
	"log"
	// "mime/multipart"
	// "strconv"

	"time"

	db "github.com/zyqhpz/be-eventeq/Database"
	util "github.com/zyqhpz/be-eventeq/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	// "go.mongodb.org/mongo-driver/mongo/gridfs"
	// "go.mongodb.org/mongo-driver/mongo/options"

	// util "github.com/zyqhpz/be-eventeq/util"

	"github.com/gofiber/fiber/v2"
)

/*
	* Connect to the "chat" collection
	@param client *mongo.Client
*/
func ConnectDBChats(client *mongo.Client) *mongo.Collection {
	// Get a handle to the "users" collection
	collection := client.Database("eventeq").Collection("chats")
	return collection
}

/*
* GET /api/chat/users
* Get all users
 */
func GetChatUsers(c *fiber.Ctx) error {

	// get id from params
	id := c.Params("id")

	// convert id to primitive.ObjectID
	oid, err := primitive.ObjectIDFromHex(id)

	type Body struct {
		ID        primitive.ObjectID 	`bson:"_id,omitempty"`
		FirstName string             	`bson:"first_name"`
		LastName  string             	`bson:"last_name"`
		Email     string             	`bson:"email"`

		// Timestamp
		CreatedAt time.Time 			`bson:"created_at"`
		UpdatedAt time.Time 			`bson:"updated_at"`
	}

	client, err  := db.ConnectDB()

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	defer client.Disconnect(ctx)
	collectionUsers := ConnectDBUsers(client)

	// Find all users except the current user (id)
	cursor, err := collectionUsers.Find(ctx, bson.M{"_id": bson.M{"$ne": oid}})

	// cursor, err := collectionUsers.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	// Iterate through the documents and print them
	var users []Body
	for cursor.Next(ctx) {
		var user Body
		if err := cursor.Decode(&user); err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	return c.JSON(users)
}

func SendMessage(c *fiber.Ctx) error {

	type Request struct {
		Message  	string  `bson:"message"`
		Sender     	string  `bson:"sender"`
		Receiver   	string  `bson:"receiver"`
	}

	req := new(Request)
	if err := c.BodyParser(req); err != nil {
		log.Println("Error parsing JSON request body:", err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	message := req.Message

	sender, err := primitive.ObjectIDFromHex(req.Sender)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid sender ID",
		})
	}

	receiver, err := primitive.ObjectIDFromHex(req.Receiver)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid receiver ID",
		})
	}
	
	type Body struct {
		ID		 	primitive.ObjectID 	`bson:"_id,omitempty"`
		Message  	string             	`bson:"message"`
		Sender	 	primitive.ObjectID 	`bson:"sender"`
		Receiver 	primitive.ObjectID 	`bson:"receiver"`
		CreatedAt 	time.Time 			`bson:"created_at"`
		UpdatedAt 	time.Time 			`bson:"updated_at"`
	}

	body := Body{
		ID: primitive.NewObjectID(),
		Message: message,
		Sender: sender,
		Receiver: receiver,
		CreatedAt: util.GetCurrentTime(),
		UpdatedAt: util.GetCurrentTime(),
	}

	client, err  := db.ConnectDB()

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	defer client.Disconnect(ctx)
	collectionChats := ConnectDBChats(client)

	// Insert new item into database
	res, err := collectionChats.InsertOne(ctx, body)
	if err != nil {
		return err
	}

	// Return response
	return c.JSON(fiber.Map{
		"status": "success",
		"message": "Successfully sent message",
		"data": res,
	})
}

func FetchMessages(c *fiber.Ctx) error {
	type Request struct {
		Sender     	string  `json:"sender"`
		Receiver   	string  `json:"receiver"`
	}

	req := new(Request)
	if err := c.BodyParser(req); err != nil {
		log.Println("Error parsing JSON request body:", err)
		log.Println("Request body:", req)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	sender, err := primitive.ObjectIDFromHex(req.Sender)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid sender ID",
		})
	}

	receiver, err := primitive.ObjectIDFromHex(req.Receiver)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid receiver ID",
		})
	}

	client, err  := db.ConnectDB()

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	defer client.Disconnect(ctx)
	collecttionChats := ConnectDBChats(client)

	// Find all chats with sender and receiver
	cursor, err := collecttionChats.Find(ctx, bson.M{"sender": sender, "receiver": receiver})

	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	type Body struct {
		ID        	primitive.ObjectID 	`bson:"_id,omitempty"`
		Message  	string             	`bson:"message"`
		Sender	 	primitive.ObjectID 	`bson:"sender"`
		Receiver 	primitive.ObjectID 	`bson:"receiver"`
		CreatedAt 	time.Time 			`bson:"created_at"`
	}

	// Iterate through the documents and print them
	var chats []Body
	for cursor.Next(ctx) {
		var chat Body
		if err := cursor.Decode(&chat); err != nil {
			log.Fatal(err)
		}
		chats = append(chats, chat)
	}
	return c.JSON(chats)
}
