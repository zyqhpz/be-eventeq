package main

import (
	"log"
	"os"
	controller "github.com/zyqhpz/be-eventeq/Controllers"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/contrib/websocket"
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
	app.Put("/api/user/update/:id", controller.UpdateUserById).Name("user.update")
	app.Put("/api/user/update/:email/password", controller.UpdateUserPasswordById).Name("user.updatePassword")

	/* Item */
	app.Get("/api/item", controller.GetItems).Name("item.get")
	app.Get("/api/itemActive", controller.GetActiveItems).Name("item.getActive")
	app.Post("/api/item/create", controller.AddItem).Name("item.create")
	app.Put("/api/item/update/:id", controller.UpdateItem).Name("item.update")
	app.Get("/api/item/:id", controller.GetItemById).Name("item.getItemById")
	app.Get("/api/item/user/:id", controller.GetItemsByUserId).Name("item.getItemByUserId")
	app.Get("/api/item/image/:id", controller.GetItemImageById).Name("item.getImage")
	
	app.Get("/api/itemsWithUser", controller.GetItemsWithUser).Name("item.getWithUser")
	app.Get("/api/itemsActiveWithUser", controller.GetItemsActiveWithUser).Name("item.getActiveWithUser")

	app.Get("/api/item/:itemId/booking", controller.GetBookingsByItemId).Name("item.getBookingsByItemId")

	/* Booking */
	app.Get("/api/itemsForBooking/:ownerId", controller.GetItemDetailsForBooking).Name("booking.getItemDetailsForBooking")
	app.Post("/api/booking/create", controller.CreateNewBooking).Name("booking.create")
	app.Put("/api/booking/cancel/:bookingId", controller.UpdateBookingStatusToCancelled).Name("booking.cancel")
	app.Get("/api/booking/:userId/upcoming/listing", controller.GetUpcomingBookingListByUserID).Name("booking.getUpcomingBookingListByUserID")
	app.Get("/api/booking/:userId/active/listing", controller.GetActiveBookingListByUserID).Name("booking.getActiveBookingListByUserID")
	app.Get("/api/booking/:userId/ended/listing", controller.GetEndedBookingListByUserID).Name("booking.getEndedBookingListByUserID")
	app.Get("/api/booking/:userId/inBooking/listing", controller.GetItemInBookingListByUserID).Name("booking.getInBookingListByUserID")
	app.Get("/api/booking/active/:bookingId", controller.GetActiveBookingByBookingID).Name("booking.getActiveBookingByBookingID")
	app.Put("/api/booking/active/:bookingId/retrieve", controller.UpdateBookingStatusAfterItemRetrieved).Name("booking.updateBookingStatusAfterItemRetrieved")
	app.Put("/api/booking/active/:bookingId/return", controller.UpdateBookingStatusAfterItemReturned).Name("booking.updateBookingStatusAfterItemReturned")
	app.Put("/api/booking/giveRating", controller.UpdateFeedbackBooking).Name("booking.updateFeedbackBooking")

	/* Event */
	app.Get("/api/event", controller.GetEvents).Name("event.get")
	app.Get("/api/eventActive", controller.GetActiveEvents).Name("event.getActive")
	app.Post("/api/event/create", controller.AddEvent).Name("event.create")
	app.Put("/api/event/update/:id", controller.UpdateEvent).Name("event.update")
	app.Get("/api/event/:id", controller.GetEventById).Name("event.getEventById")
	app.Get("/api/event/user/:id", controller.GetEventsByUserId).Name("event.getEventByUserId")
	app.Get("/api/eventsWithUser", controller.GetEventsWithUser).Name("event.getWithUser")
	app.Get("/api/eventsActiveWithUser", controller.GetEventsActiveWithUser).Name("event.getActiveWithUser")

	/* Chat */
	app.Get("/api/chat/getUsers/:id", controller.GetChatUsers).Name("chat.getUsers")
	app.Post("/api/chat/messages/", controller.FetchMessages).Name("chat.fetchMessages")
	app.Post("/api/chat/messages/send", controller.SendMessage).Name("chat.sendMessage")

	/* Payment */
	app.Get("/api/payment/redirect", controller.HandleRedirectUrl).Name("payment.redirect")
	app.Post("/api/payment/callback", controller.HandleCallbackUrl).Name("payment.callback")

	// WebSocket route
	app.Get("/ws", websocket.New(controller.WebSocketChat))

	// SSE route
	app.Get("/sse", adaptor.HTTPHandler(controller.Handler(controller.DashboardHandler)))

	app.Post("/user/test/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Login Success"})
	})
}

func main() {
    app := fiber.New(fiber.Config{
		// EnablePrintRoutes: true,
		// DisableStartupMessage: true,
	})
	
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins: "http://localhost:5173, https://fe-eventeq.vercel.app",
		AllowMethods: "GET, POST, PUT, DELETE",
		AllowHeaders: "Origin, Content-Type, Accept, Cache-Control, Connection",
	}))
	
	port := os.Getenv("PORT")
	
	if port == "" {
		port = "8080"
	}
	
	setupRoutes(app)
	log.Fatal(app.Listen("0.0.0.0:" + port))
}