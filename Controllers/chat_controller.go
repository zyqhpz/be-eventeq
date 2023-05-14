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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	// "go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/gridfs"
	// "go.mongodb.org/mongo-driver/mongo/options"

	// util "github.com/zyqhpz/be-eventeq/util"

	"github.com/gofiber/fiber/v2"
)

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