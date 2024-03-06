package router

import (
	"github.com/gofiber/fiber/v2"
)

func hello(c *fiber.Ctx) error {
	return c.SendString("Hello ChinnoH!")
}

func SetupRoutes(app *fiber.App) {
	// homePage
	api := app.Group("/")
	api.Get("/", hello)

	SetupUserRoutes(app)
}
