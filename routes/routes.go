package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/henryosei/async-go/handlers"
)

func Setup(app *fiber.App) {
	app.Post("/process/file", handlers.FileUpload)

}
