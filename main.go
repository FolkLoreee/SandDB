package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func hello(c *fiber.Ctx) error {
	err := c.SendString("Henlo world with fiber")
	if err != nil {
		log.Fatalf("Error in hello world: %s", err)
	}
	return err
}
func main() {
	app := fiber.New()
	app.Get("/", hello)
	err := app.Listen(":8000")
	if err != nil {
		log.Fatalf("Error in starting up server: %s", err)
	}
}
