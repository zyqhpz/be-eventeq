package model

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     int    `json:"role"`
	Email    string `json:"email"`
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