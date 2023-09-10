import (
	    "net/http"
		    "net/http/httptest"
			    "testing"
				    "github.com/gofiber/fiber/v2"
					
					
)

func TestYourEndpoint(t *testing.T) {
	    // Create a new Fiber app
		    app := fiber.New()

			    // Define your test route
				    app.Get("/your-endpoint", func(c *fiber.Ctx) error {
					        // Your endpoint logic here
							        return c.SendString("Hello, World!")
									    })

										    // Create a new HTTP request
											    req := httptest.NewRequest(http.MethodGet, "/your-endpoint", nil)
												    resp := httptest.NewRecorder()

													    // Serve the request using your Fiber app
														    app.ServeHTTP(resp, req)

															    // Check the response
																    if resp.Code != http.StatusOK {
																	        t.Errorf("Expected status code %d but got %d", http.StatusOK, resp.Code)
																			    }

																				    expectedResponse := "Hello, World!"
																					    if resp.Body.String() != expectedResponse {
																						        t.Errorf("Expected response body %s but got %s", expectedResponse, resp.Body.String())
																								    }
																									}

}