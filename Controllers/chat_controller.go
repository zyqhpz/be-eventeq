package api

import (
	// "bytes"
	"context"
	"encoding/json"
	"sync"

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

	"github.com/gofiber/contrib/websocket"
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

	// create condition to find all chats with sender and receiver or switch sender and receiver
		// Construct the query
	filter := bson.M{
		"$or": []bson.M{
			{
				"$and": []bson.M{
					{"sender": sender},
					{"receiver": receiver},
				},
			},
			{
				"$and": []bson.M{
					{"sender": receiver},
					{"receiver": sender},
				},
			},
		},
	}

	// Find all chats with sender and receiver
	cursor, err := collecttionChats.Find(ctx, filter)

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


	if chats == nil {
		// Find all chats with switch sender and receiver
		cursor, err := collecttionChats.Find(ctx, bson.M{"sender": receiver, "receiver": sender})

		if err != nil {
			log.Fatal(err)
		}
		defer cursor.Close(ctx)

		// Iterate through the documents and print them
		for cursor.Next(ctx) {
			var chat Body
			if err := cursor.Decode(&chat); err != nil {
				log.Fatal(err)
			}
			chats = append(chats, chat)
		}
	}

	return c.JSON(chats)
}

var clients = make(map[*websocket.Conn]bool)
var clientsLock = sync.RWMutex{}

// WebSocket route
func WebSocketChat(c *websocket.Conn) {
	// Handle WebSocket connection
	log.Println("New client connected")

	// Add the client to the map of connected clients
	clientsLock.Lock()
	clients[c] = true
	clientsLock.Unlock()

	defer func() {
		log.Println("Closing client connection")
		c.Close()

		// Remove the client from the map of connected clients
		clientsLock.Lock()
		delete(clients, c)
		clientsLock.Unlock()
	}()

	// Read messages from the client
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Println("Read error:", err)
			}
			break
		}

		// convert msg to JSON
		var jsonMsg map[string]interface{}
		err = json.Unmarshal([]byte(msg), &jsonMsg)
		if err != nil {
			log.Println("Error unmarshalling JSON:", err)
			break
		}

		// get sender, receiver and message from JSON
		sender := jsonMsg["sender"].(string)
		receiver := jsonMsg["receiver"].(string)
		message := jsonMsg["message"].(string)

		// convert sender and receiver to primitive.ObjectID
		senderOID, err := primitive.ObjectIDFromHex(sender)
		if err != nil {
			log.Println("Error converting sender to primitive.ObjectID:", err)
			break
		}

		receiverOID, err := primitive.ObjectIDFromHex(receiver)
		if err != nil {
			log.Println("Error converting receiver to primitive.ObjectID:", err)
			break
		}

		// create body
		type Body struct {
			ID        	primitive.ObjectID 	`bson:"_id,omitempty"`
			Message  	string             	`bson:"message"`
			Sender	 	primitive.ObjectID 	`bson:"sender"`
			Receiver 	primitive.ObjectID 	`bson:"receiver"`
			CreatedAt 	time.Time 			`bson:"created_at"`
		}

		body := Body{
			ID: primitive.NewObjectID(),
			Message: jsonMsg["message"].(string),
			Sender: senderOID,
			Receiver: receiverOID,
			CreatedAt: util.GetCurrentTimeUTC(),
		}

		client, err  := db.ConnectDB()

		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()
		defer client.Disconnect(ctx)
		collectionChats := ConnectDBChats(client)

		// Insert new item into database
		_, err = collectionChats.InsertOne(ctx, body)
		if err != nil {
			log.Println("Error inserting new chat into database:", err)
			break
		}

		log.Printf("Received message from id %s to id %s: %s", sender, receiver, message)

		response := map[string]interface{}{
			"sender": sender,
			"receiver": receiver,
			"message": message,
		}

		res, _ := json.Marshal(response)

		// Echo the message back to the client and broadcast to all clients
		clientsLock.RLock()
		for client := range clients {
			err = client.WriteMessage(websocket.TextMessage, res)
			if err != nil {
				log.Println("Write error:", err)
			}
		}
		clientsLock.RUnlock()
	}
}
