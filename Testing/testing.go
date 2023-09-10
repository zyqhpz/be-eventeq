package testing

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func YourEndpointHandler(c *fiber.Ctx) error {
	// Your endpoint logic here
	return c.SendString("Hello, World!")
}

func TestYourEndpoint(t *testing.T) {
	// Create a new Fiber app
	app := fiber.New()

	// Define your test route
	app.Get("/your-endpoint", YourEndpointHandler)

	// Create a new HTTP request
	req := httptest.NewRequest(http.MethodGet, "/your-endpoint", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Failed to perform the request: %v", err)
	}

	// Check the response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d but got %d", http.StatusOK, resp.StatusCode)
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expectedResponse := "Hello, World!"
	if string(body) != expectedResponse {
		t.Errorf("Expected response body %s but got %s", expectedResponse, string(body))
	}
}
