package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/henryosei/async-go/routes"
	"github.com/henryosei/async-go/tasks"
	"log"
)

const redisAddr = "127.0.0.1:6379"

func main() {
	app := fiber.New(fiber.Config{
		AppName: "Test", Concurrency: 10
	})

	routes.Setup(app)
	tasks.Init(redisAddr)
	defer tasks.Close()
	log.Fatal(app.Listen(":3001"))
}
