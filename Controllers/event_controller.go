package api

import (
	"context"
	"log"
	"strconv"
	"time"

	db "github.com/zyqhpz/be-eventeq/Database"
	model "github.com/zyqhpz/be-eventeq/Models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	util "github.com/zyqhpz/be-eventeq/util"

	"github.com/gofiber/fiber/v2"
)

/*
	* Connect to the "items" collection
	@param client *mongo.Client
*/
func ConnectDBEvents(client *mongo.Client) (*mongo.Collection) {
	// Get a handle to the "users" collection
	collection := client.Database("eventeq").Collection("events")
	return collection
}

/*
	* GET /api/events
	* Get all events
*/
func GetEvents(c *fiber.Ctx) error {

	client, err  := db.ConnectDB()

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	defer client.Disconnect(ctx)
	collectionEvents := ConnectDBEvents(client)

	cursor, err := collectionEvents.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	// Iterate through the documents and print them
	var events []model.Event
	for cursor.Next(ctx) {
		var event model.Event
		if err := cursor.Decode(&event); err != nil {
			log.Fatal(err)
		}
		events = append(events, event)
	}
	return c.JSON(events)
}

/*
	* GET /api/eventsActive
	* Get all items with status 1
*/
func GetActiveEvents(c *fiber.Ctx) error {

	client, err  := db.ConnectDB()

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	defer client.Disconnect(ctx)
	collectionEvents := ConnectDBEvents(client)

	cursor, err := collectionEvents.Find(ctx, bson.M{"status": 1})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	// Iterate through the documents and print them
	var events []model.Event
	for cursor.Next(ctx) {
		var event model.Event
		if err := cursor.Decode(&event); err != nil {
			log.Fatal(err)
		}
		events = append(events, event)
	}
	return c.JSON(events)
}

/*
	* GET /api/eventsWithUser
	* Get all events with user
*/
func GetEventsWithUser(c *fiber.Ctx) error {

	type Data struct {
		ID        		primitive.ObjectID 	`bson:"_id,omitempty"`
		Name        	string 				`bson:"name"`
		Description 	string 				`bson:"description"`
		Location		model.Location		`bson:"location"`
		StartDate   	string 				`bson:"start_date"`
		EndDate     	string 				`bson:"end_date"`
		Status	  		int 				`bson:"status"`
		OrganizedBy 	model.User
		CreatedAt 		time.Time 			`bson:"created_at"`
		UpdatedAt 		time.Time 			`bson:"updated_at"`
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
				"localField":   "organized_by",
				"foreignField": "_id",
				"as":           "organized_by",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$organized_by",
				"preserveNullAndEmptyArrays": true,
			},
		},
	}

	ctx := context.Background()
	defer client.Disconnect(ctx)
	collectionEvents := ConnectDBEvents(client)

	cursor, err := collectionEvents.Aggregate(ctx, pipeline)

	if err != nil {
		log.Fatal(err)
	}

	defer cursor.Close(ctx)

	// Iterate through the documents and print them
	var events []Data
	for cursor.Next(ctx) {
		var event Data
		if err := cursor.Decode(&event); err != nil {
			log.Fatal(err)
		}
		events = append(events, event)
	}
	return c.JSON(events)
}

/*
	* GET /api/eventsActiveWithUser
	* Get all events with user and status 1
*/
func GetEventsActiveWithUser(c *fiber.Ctx) error {

	type User struct {
		ID 					primitive.ObjectID 	`bson:"_id"`
		FirstName 			string 				`bson:"first_name"`
		LastName 			string 				`bson:"last_name"`
		IsAvatarImageSet 	bool 				`bson:"isAvatarImageSet"`
		ProfileImage 		primitive.ObjectID 	`bson:"profile_image"`
		Location 			model.Location 		`bson:"location"`
	}

	type Data struct {
		ID        		primitive.ObjectID 	`bson:"_id"`
		Name        	string 				`bson:"name"`
		Description 	string 				`bson:"description"`
		Location		model.Location		`bson:"location"`
		StartDate   	string 				`bson:"start_date"`
		EndDate     	string 				`bson:"end_date"`
		Status	  		int 				`bson:"status"`
		Organized_By 	User
		CreatedAt 		time.Time 			`bson:"created_at"`
		UpdatedAt 		time.Time 			`bson:"updated_at"`
	}

	client, err  := db.ConnectDB()

	if err != nil {
		log.Fatal(err)
	}

	// create a pipeline for the aggregation to get all events with status 1 and user
	pipeline := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "organized_by",
				"foreignField": "_id",
				"as":           "organized_by",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$organized_by",
				"preserveNullAndEmptyArrays": true,
			},
		},
				{
			"$match": bson.M{"status": 1},
		},
	}

	ctx := context.Background()
	defer client.Disconnect(ctx)
	collectionEvents := ConnectDBEvents(client)

	cursor, err := collectionEvents.Aggregate(ctx, pipeline)

	if err != nil {
		log.Fatal(err)
	}

	defer cursor.Close(ctx)

	// Iterate through the documents and print them
	var events []Data
	for cursor.Next(ctx) {
		var event Data
		if err := cursor.Decode(&event); err != nil {
			log.Fatal(err)
		}
		events = append(events, event)
	}
	return c.JSON(events)
}

/*
	* GET /api/event/:id
	* Get an event by id
*/
func GetEventById(c *fiber.Ctx) error {
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

	// Select the `events` collection from the database
	collectionEvents := ConnectDBEvents(client)

	type Data struct {
		ID        		primitive.ObjectID 	`bson:"_id,omitempty"`
		Name        	string 				`bson:"name"`
		Description 	string 				`bson:"description"`
		Location		model.Location		`bson:"location"`
		StartDate   	string 				`bson:"start_date"`
		EndDate     	string 				`bson:"end_date"`
		Status	  		int 				`bson:"status"`
		OrganizedBy 	model.User
		CreatedAt 		time.Time 			`bson:"created_at"`
		UpdatedAt 		time.Time 			`bson:"updated_at"`
	}

	// create a pipeline for the aggregation
	pipeline := []bson.M{
		{
			"$match": bson.M{"_id": id},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "organized_by",
				"foreignField": "_id",
				"as":           "organized_by",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$organized_by",
				"preserveNullAndEmptyArrays": true,
			},
		},
	}

	// execute the aggregation pipeline
	cursor, err := collectionEvents.Aggregate(ctx, pipeline)
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
	* GET /api/event/user/:id/
	* Get events by User ID
*/
func GetEventsByUserId(c *fiber.Ctx) error {

	// Retrieve the `id` parameter from the request URL
	idParam := c.Params("id")

	// Convert the `id` parameter to a MongoDB `ObjectID`
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		// Return an error response if the `id` parameter is not a valid `ObjectID`
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid event ID",
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

	// Select the `events` collection from the database
	collectionEvents := ConnectDBEvents(client)
	ctx := context.Background()

	// Query for the Event document and filter by the User ID in organizedBy
	cursor, err := collectionEvents.Find(ctx, bson.M{"organized_by": id})
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
	var events []model.Event
	for cursor.Next(ctx) {
		var event model.Event
		if err := cursor.Decode(&event); err != nil {
			log.Fatal(err)
		}
		events = append(events, event)
	}
	return c.JSON(events)
}

/*
	* POST /api/event/add
	* Add an event
*/
func AddEvent(c *fiber.Ctx) error {

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	id := form.Value["userID"][0]
	userID, _ := primitive.ObjectIDFromHex(id)

	// Get all files from form
	name := form.Value["name"][0]
	description := form.Value["description"][0]
	startDate := form.Value["start_date"][0]
	endDate := form.Value["end_date"][0]
	state := form.Value["state"][0]
	district := form.Value["district"][0]

	client, err  := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	defer client.Disconnect(ctx)

	location := model.Location{
		State: state,
		District: district,
	}

	event := model.Event{
		ID: primitive.NewObjectID(),
		Name: name,
		Description: description,
		Location: location,
		StartDate: startDate,
		EndDate: endDate,
		Status: 1,
		OrganizedBy: userID,
		CreatedAt: util.GetCurrentTime(),
		UpdatedAt: util.GetCurrentTime(),
	}
	
	// Select the `events` collection from the database
	collectionEvents := ConnectDBEvents(client)

	// Insert new item into database
	res, err := collectionEvents.InsertOne(context.Background(), event)
	if err != nil {
		return err
	}

	log.Println("[Event] Inserted new event: ", res.InsertedID)

	// Return response
	return c.JSON(fiber.Map{
		"status": "success",
		"message": "Event added successfully",
		"event_id": res.InsertedID,
	})
}

func UpdateEvent(c *fiber.Ctx) error {

	id := c.Params("id")

	objectID, err := primitive.ObjectIDFromHex(id)

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	// Get all files from form
	name := form.Value["name"][0]
	description := form.Value["description"][0]
	startDate := form.Value["start_date"][0]
	endDate := form.Value["end_date"][0]

	state := form.Value["state"][0]
	district := form.Value["district"][0]

	status := form.Value["status"][0]

	// convert string to int
	statusInt, _ := strconv.Atoi(status)

	client, err  := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	defer client.Disconnect(ctx)

	location := model.Location{
		State: state,
		District: district,
	}

	event := model.Event{
		ID: objectID,
		Name: name,
		Description: description,
		Location: location,
		StartDate: startDate,
		EndDate: endDate,
		Status: statusInt,
		UpdatedAt: util.GetCurrentTime(),
	}

	// update item into database
	collectionEvents := ConnectDBEvents(client)

	filter := bson.M{"_id": event.ID}
	update := bson.M{"$set": bson.M{
		"name": event.Name,
		"description": event.Description,
		"location": event.Location,
		"start_date": event.StartDate,
		"end_date": event.EndDate,
		"status": event.Status,
		"updated_at": event.UpdatedAt,
	}}

	res, err := collectionEvents.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	log.Println("[Event] Updated event: ", res.UpsertedID)

	// Return response
	return c.JSON(fiber.Map{
		"status": "success",
		"message": "Event updated successfully",
		"event_id": res.UpsertedID,
	})
}