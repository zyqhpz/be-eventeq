package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     int 	`json:"role"`
	Email	 string `json:"email"`
}

var users = []User{
	{ID: "1", Username: "admin", Password: "admin", Role: 0, Email: "admin@gmail.com"},
	{ID: "2", Username: "user", Password: "user", Role: 1, Email: "user.1@gmail.com"}, // 1 is owner
	{ID: "3", Username: "user2", Password: "user", Role: 2, Email: "user.2@gmail.com"}, // 2 is rentee
}

type Book struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Quantity int    `json:"quantity"`
}

// create a function for login user that accepts username and password parameter from POST request and return user if found
func loginUser(c *gin.Context) {
	username, ok := c.GetQuery("username")
	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Missing username query parameter"})
	}
	password, ok := c.GetQuery("password")
	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Missing password query parameter"})
	}

	for _, u := range users {
		if u.Username == username && u.Password == password {
			c.IndentedJSON(http.StatusOK, u)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "User not found"})
}

// create a function for register user
func registerUser(c *gin.Context) {
	var newUser User
	if err := c.BindJSON(&newUser); err != nil {
		return
	}
	users = append(users, newUser)
	c.IndentedJSON(http.StatusCreated, newUser)
}

// create function to get all users
func getUsers(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, users)
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())
	router.POST("/user/login", loginUser)
	router.POST("/user/register", registerUser)
	router.GET("/user", getUsers)
	router.Run("localhost:8080")
}