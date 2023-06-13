package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
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

func CreateNewBooking(c *fiber.Ctx) error {
	bodyBytes := c.Body()
	var bodyMap map[string]interface{}
	err := json.Unmarshal(bodyBytes, &bodyMap)
	if err != nil {
		log.Println("error", err)
	}

	type Item struct {
		ItemID 		primitive.ObjectID 	`bson:"id"`
		Name		string				`bson:"name"`
		Price		float64 			`bson:"price"`
		Quantity 	int32 				`bson:"quantity"`
	}

	type Booking struct {
		ID        	primitive.ObjectID 	`bson:"_id,omitempty"`
		UserID 		primitive.ObjectID 	`bson:"user_id"`
		OwnerID		primitive.ObjectID 	`bson:"owner_id"`
		Items 		[]Item 				`bson:"items"`
		StartDate 	string 				`bson:"start_date"`
		EndDate 	string 				`bson:"end_date"`
		SubTotal 	float64 			`bson:"sub_total"`
		ServiceFee 	float64 			`bson:"service_fee"`
		GrandTotal 	float64 			`bson:"grand_total"`
		Status 		int32 				`bson:"status"`
		CreatedAt 	time.Time 			`bson:"created_at"`
		UpdatedAt 	time.Time 			`bson:"updated_at"`
	}
	
	booking := new(Booking)
	
	// Parse the user_id as a string
	if userID, ok := bodyMap["user_id"].(string); ok {
		uid, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid user id",
			})
		}
		booking.UserID = uid
	} else {
		return fmt.Errorf("user_id is not a string")
	}

	// Parse the owner_id as a string
	if ownerID, ok := bodyMap["owner_id"].(string); ok {
		oid, err := primitive.ObjectIDFromHex(ownerID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid owner id",
			})
		}
		booking.OwnerID = oid
	} else {
		return fmt.Errorf("owner_id is not a string")
	}

	// Parse the start_date as a string
	if startDate, ok := bodyMap["start_date"].(string); ok {
		booking.StartDate = startDate
	} else {
		return fmt.Errorf("start_date is not a string")
	}

	// Parse the end_date as a string
	if endDate, ok := bodyMap["end_date"].(string); ok {
		booking.EndDate = endDate
	} else {
		return fmt.Errorf("end_date is not a string")
	}

	// Parse the sub_total as a float64
	if subTotal, ok := bodyMap["sub_total"].(float64); ok {
		booking.SubTotal = subTotal
	} else {
		return fmt.Errorf("sub_total is not a float64")
	}

	// Parse the service_fee as a float64
	if serviceFee, ok := bodyMap["service_fee"].(float64); ok {
		booking.ServiceFee = serviceFee
	} else {
		return fmt.Errorf("service_fee is not a float64")
	}

	// Parse the grand_total as a float64
	if grandTotal, ok := bodyMap["grand_total"].(float64); ok {
		booking.GrandTotal = grandTotal
	} else {
		return fmt.Errorf("grand_total is not a float64")
	}

	client, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	// Select the `bookings` collection from the database
	collection := ConnectDBBookings(client)
	ctx := context.Background()

	if (bodyMap["items"] != nil) {
		items := bodyMap["items"].([]interface{})
		for _, item := range items {
			itemMap := item.(map[string]interface{})
			// itemID := itemMap["id"].(string)
			itemName := itemMap["name"].(string)
			itemPrice := itemMap["price"].(float64)
			itemQuantity := itemMap["quantity"].(float64)
			
			// Parse the item_id as a string
			if itemID, ok := itemMap["id"].(string); ok {
				iid, err := primitive.ObjectIDFromHex(itemID)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"message": "Invalid item id",
					})
				}
				item := new(Item)
				item.ItemID = iid
				item.Name = itemName
				item.Price = itemPrice
				item.Quantity = int32(itemQuantity)
				booking.Items = append(booking.Items, *item)
			} else {
				return fmt.Errorf("item_id is not a string")
			}
		}
	}

	// booking.Status = 0 // 0 = pending, 1 = accepted, 2 = rejected, 3 = completed
	booking.Status = 0 // 0 = upcoming, 1 = active, 2 = completed, 3 = cancelled
	booking.CreatedAt = time.Now()
	booking.UpdatedAt = time.Now()

	// Insert the `bookings` document in the database
	result, err := collection.InsertOne(ctx, booking)
	if err != nil {
		// Return an error response if there is a database error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create booking",
		})
	}

	defer client.Disconnect(ctx)

	// Return a success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Booking created",
		"data": result,
	})
}

func GetUpcomingBookingListByUserID(c *fiber.Ctx) error {
	// get id from params
	userId := c.Params("userId")

	// convert id to primitive.ObjectID
	uid, err := primitive.ObjectIDFromHex(userId)

	type Item struct {
		ItemID 		primitive.ObjectID 	`bson:"id"`
		Name		string				`bson:"name"`
		Price		float64 			`bson:"price"`
		Quantity 	int32 				`bson:"quantity"`
	}

	type Booking struct {
		ID        	primitive.ObjectID 	`bson:"_id,omitempty"`
		UserID 		primitive.ObjectID 	`bson:"user_id"`
		OwnerID		primitive.ObjectID 	`bson:"owner_id"`
		Items 		[]Item 				`bson:"items"`
		StartDate 	string 				`bson:"start_date"`
		EndDate 	string 				`bson:"end_date"`
		SubTotal 	float64 			`bson:"sub_total"`
		ServiceFee 	float64 			`bson:"service_fee"`
		GrandTotal 	float64 			`bson:"grand_total"`
		Status 		int32 				`bson:"status"`
		CreatedAt 	time.Time 			`bson:"created_at"`
		UpdatedAt 	time.Time 			`bson:"updated_at"`
	}

	client, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	// Select the `bookings` collection from the database
	bookingsCollection := ConnectDBBookings(client)
	ctx := context.Background()

	// Query for the Item document and filter by the User ID in ownedBy
	cursor, err := bookingsCollection.Find(ctx, bson.M{"user_id": uid})
	if err != nil {
		// Return an error response if the document is not found
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Booking not found",
			})
		}
		// Return an error response if there is a database error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get booking from database",
		})
	}
	defer cursor.Close(ctx)

	// Iterate through the documents and print them
	var bookings []Booking
	for cursor.Next(ctx) {
		var booking Booking
		if err := cursor.Decode(&booking); err != nil {
			log.Fatal(err)
		}
		startDateString := booking.StartDate

		// convert DD/MM/YYYY to date object for comparison
		layout := "02/01/2006" // Specify the layout to match the input date format
		date, _ := time.Parse(layout, startDateString)

		// compare date with current date
		if date.After(time.Now()) {
			bookings = append(bookings, booking)
		} else {
			// update status in database
			updateResult, _ := bookingsCollection.UpdateOne(
				ctx,
				bson.M{"_id": booking.ID},
				bson.D{
					{Key: "$set", Value: bson.D{{Key: "status", Value: 1}}},
				},
			)

			booking.Status = 1 // 0 = upcoming, 1 = active, 2 = completed, 3 = cancelled
			booking.UpdatedAt = time.Now()

			log.Printf("Running update query for bookings, matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
		}
	}

	// sort by start date (ascending)
	sort.Slice(bookings, func(i, j int) bool {
		layout := "02/01/2006" // Specify the layout to match the input date format

		// Parse the date string into a time.Time object
		date1, _ := time.Parse(layout, bookings[i].StartDate)
		date2, _ := time.Parse(layout, bookings[j].StartDate)

		return date1.Before(date2)
	})

	defer client.Disconnect(ctx)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return c.JSON(bookings)
}

func GetActiveBookingListByUserID(c *fiber.Ctx) error {
	// get id from params
	userId := c.Params("userId")

	// convert id to primitive.ObjectID
	uid, err := primitive.ObjectIDFromHex(userId)

	type Item struct {
		ItemID 		primitive.ObjectID 	`bson:"id"`
		Name		string				`bson:"name"`
		Price		float64 			`bson:"price"`
		Quantity 	int32 				`bson:"quantity"`
	}

	type Booking struct {
		ID        	primitive.ObjectID 	`bson:"_id,omitempty"`
		UserID 		primitive.ObjectID 	`bson:"user_id"`
		OwnerID		primitive.ObjectID 	`bson:"owner_id"`
		Items 		[]Item 				`bson:"items"`
		StartDate 	string 				`bson:"start_date"`
		EndDate 	string 				`bson:"end_date"`
		SubTotal 	float64 			`bson:"sub_total"`
		ServiceFee 	float64 			`bson:"service_fee"`
		GrandTotal 	float64 			`bson:"grand_total"`
		Status 		int32 				`bson:"status"`
		CreatedAt 	time.Time 			`bson:"created_at"`
		UpdatedAt 	time.Time 			`bson:"updated_at"`
	}

	client, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	// Select the `bookings` collection from the database
	bookingsCollection := ConnectDBBookings(client)
	ctx := context.Background()

	// create filter by user_id and status
	filter := bson.M{"user_id": uid, "status": 1}

	// Query for the Item document and filter by the User ID in ownedBy
	cursor, err := bookingsCollection.Find(ctx, filter)
	if err != nil {
		// Return an error response if the document is not found
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Booking not found",
			})
		}
		// Return an error response if there is a database error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get booking from database",
		})
	}
	defer cursor.Close(ctx)

	// Iterate through the documents and print them
	var bookings []Booking
	for cursor.Next(ctx) {
		var booking Booking
		if err := cursor.Decode(&booking); err != nil {
			log.Fatal(err)
		}

		bookings = append(bookings, booking)
	}

	// sort by start date (ascending)
	sort.Slice(bookings, func(i, j int) bool {
		layout := "02/01/2006" // Specify the layout to match the input date format

		// Parse the date string into a time.Time object
		date1, _ := time.Parse(layout, bookings[i].StartDate)
		date2, _ := time.Parse(layout, bookings[j].StartDate)

		return date1.Before(date2)
	})

	defer client.Disconnect(ctx)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return c.JSON(bookings)
}

func GetActiveBookingByBookingID(c *fiber.Ctx) error {
	// get id from params
	bookingId := c.Params("bookingId")

	// convert id to primitive.ObjectID
	bid, err := primitive.ObjectIDFromHex(bookingId)

	type Item struct {
		ItemID 		primitive.ObjectID 	`bson:"id"`
		Name		string				`bson:"name"`
		Price		float64 			`bson:"price"`
		Quantity 	int32 				`bson:"quantity"`
		Images 		[]primitive.ObjectID	`bson:"images"`
	}

	type Booking struct {
		ID        	primitive.ObjectID 	`bson:"_id,omitempty"`
		UserID 		primitive.ObjectID 	`bson:"user_id"`
		OwnerID		primitive.ObjectID 	`bson:"owner_id"`
		Items 		[]Item 				`bson:"items"`
		StartDate 	string 				`bson:"start_date"`
		EndDate 	string 				`bson:"end_date"`
		SubTotal 	float64 			`bson:"sub_total"`
		ServiceFee 	float64 			`bson:"service_fee"`
		GrandTotal 	float64 			`bson:"grand_total"`
		Status 		int32 				`bson:"status"`
		CreatedAt 	time.Time 			`bson:"created_at"`
		UpdatedAt 	time.Time 			`bson:"updated_at"`
	}

	client, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	// Select the `bookings` collection from the database
	bookingsCollection := ConnectDBBookings(client)
	ctx := context.Background()

	// create filter by user_id and status
	filter := bson.M{"_id": bid}

	// Query for the Item document and filter by the User ID in ownedBy
	result := bookingsCollection.FindOne(ctx, filter)
	if err != nil {
		// Return an error response if the document is not found
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Booking not found",
			})
		}
		// Return an error response if there is a database error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get booking from database",
		})
	}

	var booking Booking
	if err := result.Decode(&booking); err != nil {
		log.Fatal(err)
	}

	// get Item image from database
	itemsCollection := ConnectDBItems(client)
	for i, item := range booking.Items {
		filter := bson.M{"_id": item.ItemID}
		result := itemsCollection.FindOne(ctx, filter)
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

		var item Item
		if err := result.Decode(&item); err != nil {
			log.Fatal(err)
		}

		booking.Items[i].Images = item.Images
	}

	defer client.Disconnect(ctx)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	return c.JSON(booking)
}