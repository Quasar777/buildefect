package main

import (
	"log"

	"github.com/Quasar777/buildefect/app/backend/database"
	"github.com/gofiber/fiber/v2"
)

func main() {
	database.ConnectDb()
	
	app := fiber.New()

	app.Get("/", func (c *fiber.Ctx) error {
        return c.SendString("Hello, World!")
    })

    log.Fatal(app.Listen(":8080"))
}