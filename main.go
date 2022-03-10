package main

import (
	"cassie/strict_quorum"
	"github.com/gofiber/fiber/v2"
	"log"
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
	requestHandler := &strict_quorum.RequestHandler{
		App: app,
	}
	app.Get("/", hello)
	app.Get("/request", requestHandler.HandleRequest)
	err := app.Listen(":8000")
	if err != nil {
		log.Fatalf("Error in starting up server: %s", err)
	}
}
