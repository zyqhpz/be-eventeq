package api

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"time"

	db "github.com/zyqhpz/be-eventeq/Database"
	model "github.com/zyqhpz/be-eventeq/Models"
	"github.com/zyqhpz/be-eventeq/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/golang-jwt/jwt/v5"

	"github.com/gofiber/fiber/v2"
)

type RegisterUserRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Password  string `json:"password"`
	Email     string `json:"email"`
}

/*
	* Connect to the "users" collection
	@param client *mongo.Client
*/
func ConnectDBUsers(client *mongo.Client) (*mongo.Collection) {
	// Get a handle to the "users" collection
	collection := client.Database("eventeq").Collection("users")
	return collection
}

/*
	* GET /api/user
	* Get all users
*/
func GetUsers(c *fiber.Ctx) error {
	client, err  := db.ConnectDB()

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	defer client.Disconnect(ctx)
	collectionUsers := ConnectDBUsers(client)

	cursor, err := collectionUsers.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	// Iterate through the documents and print them
	var users []model.User
	for cursor.Next(ctx) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	return c.JSON(users)
}

/*
	* POST /api/user/login
	* Login a user
*/
func LoginUser(c *fiber.Ctx) error {

	type body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	
	req := new(body)
	if err := c.BodyParser(req); err != nil {
		log.Println("Error parsing JSON request body:", err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	email := req.Email
	password := req.Password

	client, err  := db.ConnectDB()

	if err != nil {
		log.Fatal(err)
	}
	
	ctx := context.Background()
	defer client.Disconnect(ctx)
	collection := ConnectDBUsers(client)

	// Define a filter to find the user with the given username and password
    filter := bson.M{"email": email, "password": fmt.Sprintf("%x", sha256.Sum256([]byte(password)))}

    // Count the number of documents that match the filter
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    count, _ := collection.CountDocuments(ctx, filter)

    // Return true if a user with the given username and password was found, false otherwise
    if (count > 0) {
		// Get the user ID
		var result struct {
			ID primitive.ObjectID `bson:"_id"`
		}

		err := collection.FindOne(ctx, filter).Decode(&result)

		if err != nil {
			log.Fatal(err)
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": result.ID,
			"email": email,
			"exp": time.Now().Add(time.Hour * 72).Unix(),
		})

		tokenString, err := token.SignedString([]byte("secret"))
		if err != nil {
			c.JSON(http.StatusBadRequest)
			return err
		}

		c.Cookie(&fiber.Cookie{
			Name: "jwt",
			Value: tokenString,
			Expires: time.Now().Add(time.Hour * 72),
			HTTPOnly: true,
		})

		log.Println("Email: ", email, "Login Success")
		return c.JSON(fiber.Map{
			"status": "success",
			"message": "Login Success " + email,
		})
	}
	log.Println("Email: ", email, "Login Failed")
	return c.JSON(fiber.Map{"status": "failed", "message": "Login Failed " + email})
}

/*
	* GET /api/user/auth
	* Check if user is logged in
*/
func LoginUserAuth(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")

	// fmt.Println(cookie)

	if cookie == "" {
		return c.JSON(fiber.Map{"status": "failed", "message": "Not logged in"})
	}

	token, err := jwt.ParseWithClaims(cookie, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})

	if err != nil {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{"status": "failed", "message": "Not logged in"})
	}

	claims := token.Claims.(*jwt.MapClaims)

	// var user model.User
	client, err  := db.ConnectDB()

	if err != nil {
		log.Fatal(err)
	}
	
	ctx := context.Background()
	defer client.Disconnect(ctx)
	collection := ConnectDBUsers(client)

	// convert to Hex
	idHex, _ := primitive.ObjectIDFromHex((*claims)["sub"].(string))

	// Define a filter to find the user
	filter := bson.M{"_id": idHex}

	var user model.User
	result := collection.FindOne(ctx, filter).Decode(&user)

	if result != nil {
		c.JSON(http.StatusUnauthorized)
		return c.JSON(fiber.Map{"status": "failed", "message": "Not logged in"})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Logged in", "user": user})
}

/*
	* POST /api/user/logout
	* Logout a user
*/
func LogoutUser(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name: "jwt",
		Value: "",
		Expires: time.Now().Add(-time.Hour),
		HTTPOnly: true,
	})
	return c.JSON(fiber.Map{"status": "success", "message": "Logged out"})
}

/*
	* POST /api/user/register
	* Register a new user
*/
func RegisterUser(c *fiber.Ctx) error {

	req := new(RegisterUserRequest)
	if err := c.BodyParser(req); err != nil {
		log.Println("Error parsing JSON request body:", err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	firstName:= req.FirstName
	lastName := req.LastName
	email := req.Email
	password := req.Password

	// hash the password in sha256
	password = fmt.Sprintf("%x", sha256.Sum256([]byte(password)))

	client, _  := db.ConnectDB()

	collection := ConnectDBUsers(client)

	// Define a filter to find the user with the given username and email, if any exist return error message
	filterEmail := bson.M{"email": email}

	// Count the number of documents that match the filter
	ctx := context.Background()
	countEmail, _ := collection.CountDocuments(ctx, filterEmail)

	// Return error if a user with the given username was found
	if (countEmail > 0) {
		return c.JSON(fiber.Map{"status": "failed email", "message": "Email already exists"})
	} else {

		// Create a new user
		user := model.User{
			FirstName: firstName,
			LastName: lastName,
			Password: password,
			Email: email,
			CreatedAt: util.GetCurrentTime(),
			UpdatedAt: util.GetCurrentTime(),
		}

		// Insert the new user into the database
		ctx := context.Background()
		result, err := collection.InsertOne(ctx, user)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Inserted ID: ", result.InsertedID)

		return c.JSON(fiber.Map{"status": "success", "message": "User created successfully"})
	}
}

/*
	* GET /api/user/:id
	* Get a user by id
*/
func GetUserById(c *fiber.Ctx) error {
	id := c.Params("id")

	client, _  := db.ConnectDB()
	ctx := context.Background()
	defer client.Disconnect(ctx)

	collection := ConnectDBUsers(client)

	// Define a filter to find the user with the given id
	filter := bson.M{"_id": id}

	// Count the number of documents that match the filter
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	count, _ := collection.CountDocuments(ctx, filter)

	// Return true if a user with the given id was found, false otherwise
	if (count > 0) {
		var user model.User
		collection.FindOne(ctx, filter).Decode(&user)
		return c.JSON(user)
	}
	return c.JSON(fiber.Map{"status": "failed", "message": "User not found"})
}

/*
	* PUT /api/user/:id
	* Update a user by id
*/
func UpdateUserById(c *fiber.Ctx) error {
	// Email is unique, so we can use it to find the user
	email := c.Params("email")

	first_name := c.FormValue("first_name")
	last_name := c.FormValue("last_name")
	username := c.FormValue("username")

	client, _  := db.ConnectDB()
	ctx := context.Background()
	defer client.Disconnect(ctx)

	collection := ConnectDBUsers(client)

	// Define a filter to find the user with the given email
	filter := bson.M{"email": email}

	// Count the number of documents that match the filter
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	count, _ := collection.CountDocuments(ctx, filter)

	// Return true if a user with the given id was found, false otherwise
	if (count > 0) {
		// Update the user with the given id
		update := bson.M{"$set": bson.M{
			"first_name": first_name,
			"last_name": last_name,
			"username": username,
			"updated_at": util.GetCurrentTime(),
		}}

		// Update the user in the database
		_, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Fatal(err)
		}
		return c.JSON(fiber.Map{"status": "success", "message": "User updated successfully"})
	}
	return c.JSON(fiber.Map{"status": "failed", "message": "User not found"})
}

/*
	* PUT /api/user/:id/password
	* Update a user password by id
*/
func UpdateUserPasswordById(c *fiber.Ctx) error {
	// Email is unique, so we can use it to find the user
	email := c.Params("email")

	password := c.FormValue("password")

	// hash the password in sha256
	password = fmt.Sprintf("%x", sha256.Sum256([]byte(password)))

	client, _  := db.ConnectDB()
	ctx := context.Background()
	defer client.Disconnect(ctx)

	collection := ConnectDBUsers(client)

	// Define a filter to find the user with the given id
	filter := bson.M{"email": email}

	// Count the number of documents that match the filter
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	count, _ := collection.CountDocuments(ctx, filter)

	// Return true if a user with the given id was found, false otherwise
	if (count > 0) {
		update := bson.M{"$set": bson.M{
			"password": password,
			"updated_at": util.GetCurrentTime(),
		}}

		// Update the user in the database
		_, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Fatal(err)
		}

		return c.JSON(fiber.Map{"status": "success", "message": "User password updated successfully"})
	}
	return c.JSON(fiber.Map{"status": "failed", "message": "User not found"})
}

/*

	* DELETE /api/user/:id
	* Delete a user by id
*/
func DeleteUserById(c *fiber.Ctx) error {
	id := c.Params("id")

	client, _  := db.ConnectDB()
	ctx := context.Background()
	defer client.Disconnect(ctx)

	collection := ConnectDBUsers(client)

	// Define a filter to find the user with the given id
	filter := bson.M{"_id": id}

	// Count the number of documents that match the filter
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	count, _ := collection.CountDocuments(ctx, filter)

	// Return true if a user with the given id was found, false otherwise
	if (count > 0) {
		// Delete the user with the given id
		_, err := collection.DeleteOne(ctx, filter)
		if err != nil {
			log.Fatal(err)
		}
		return c.JSON(fiber.Map{"status": "success", "message": "User deleted successfully"})
	}
	return c.JSON(fiber.Map{"status": "failed", "message": "User not found"})
}

