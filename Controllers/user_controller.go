package api

import (
	model "example/be-eventeq/Models"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

var users = []model.User{
	{ID: 1, Username: "admin", Password: "admin", Role: 0, Email: "admin@gmail.com"},
	{ID: 2, Username: "user", Password: "user", Role: 1, Email: "user@gmail.com"},
}

func LoginUser(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "admin" && password == "admin" {
		fmt.Println("Username: ", username, "Login Success")
		return c.JSON(fiber.Map{"message": "Login Success"})
	}
	return c.JSON(fiber.Map{"message": "Login Failed " + username})
}

func GetUsers(c *fiber.Ctx) error {
	return c.JSON(users)
}