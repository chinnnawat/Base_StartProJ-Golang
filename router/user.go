package router

import (
	"PROJ/controller"
	"PROJ/util"

	"github.com/gofiber/fiber/v2"
)

var UserRoutes fiber.Router

func SetupUserRoutes(app *fiber.App) {
	UserRoutes := app.Group("/user")
	UserRoutes.Post("/signup", controller.CreateUser)
	UserRoutes.Post("/login", controller.LoginUser)
	UserRoutes.Get("/get-access-token", controller.GetAccessToken)

	// Private Route
	privUserRoutes := UserRoutes.Group("/private")
	privUserRoutes.Use(util.SecureAuth())
	privUserRoutes.Get("/data", controller.GetUserData)
	// UserRoutes.Use(util.SecureAuth()).Get("/data", controller.GetUserData)
}
