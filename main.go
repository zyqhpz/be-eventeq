package main

import (
	// model "example/be-eventeq/Models"

	controller "github.com/zyqhpz/be-eventeq/Controllers"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var client *mongo.Client

func setupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	
	/* User */
	app.Get("/api/user", controller.GetUsers).Name("user.get")
	app.Get("/api/user/id/:id", controller.GetUserById).Name("user.getById")
	app.Post("/api/user/login/*", controller.LoginUser).Name("user.login")

	app.Get("/api/user/auth/", controller.LoginUserAuth).Name("user.login.auth")
	app.Post("/api/user/logout/", controller.LogoutUser).Name("user.logout")

	app.Post("/api/user/register/*", controller.RegisterUser).Name("user.register")
	app.Put("/api/user/update/:email", controller.UpdateUserById).Name("user.update")
	app.Put("/api/user/update/:email/password", controller.UpdateUserPasswordById).Name("user.updatePassword")

	/* Item */
	app.Get("/api/item", controller.GetItems).Name("item.get")
	app.Post("/api/item/create", controller.AddItem).Name("item.create")
	app.Get("/api/item/:id", controller.GetItemById).Name("item.getItemById")
	app.Get("/api/item/user/:id", controller.GetItemsByUserId).Name("item.getItemByUserId")
	app.Get("/api/item/image/:id", controller.GetItemImageById).Name("item.getImage")
	
	app.Get("/api/itemWithUser", controller.GetItemsWithUser).Name("item.getWithUser")

	/* Booking */
	app.Get("/api/itemsForBooking/:ownerId", controller.GetItemDetailsForBooking).Name("booking.getItemDetailsForBooking")

	/* Chat */
	app.Get("/api/chat/getUsers/:id", controller.GetChatUsers).Name("chat.getUsers")
	app.Post("/api/chat/messages/", controller.FetchMessages).Name("chat.fetchMessages")
	app.Post("/api/chat/messages/send", controller.SendMessage).Name("chat.sendMessage")

	app.Post("/user/test/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Login Success"})
	})
}


func main() {
    app := fiber.New(fiber.Config{
		// EnablePrintRoutes: true,
		// DisableStartupMessage: true,
	})
	
	app.Use(cors.New(
		cors.Config{
			AllowCredentials: true,
		},
	))
	setupRoutes(app)
    app.Listen(":8080")
}