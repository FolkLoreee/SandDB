package main

import (
	"fmt"
	"log"
	routes "sanddb/client/clientOperations"

	"github.com/gofiber/fiber/v2"
	// "github.com/gofiber/fiber/v2/middleware/logger"
)

func hello(c *fiber.Ctx) error {
	err := c.SendString("Henlo world with fiber")
	if err != nil {
		log.Fatalf("Error in hello world: %s", err)
	}
	return err
}

func setupRoutes(app *fiber.App) {
	fmt.Println("Starting client backend...")
	app.Get("/", hello)

	// api := app.Group("/api", logger.New())

	// routes: resources and hospitals
	// query parameters MUST include id & department (room optional)
	hospital := app.Group("/hospital")
	hospital.Get("/", routes.GetAllHospitals)
	hospital.Get("/:hospitalID", routes.GetHospital)
	hospital.Post("/:hospitalID", func(c *fiber.Ctx) error { return nil })

	resource := hospital.Group("/resources")
	resource.Get("/", func(c *fiber.Ctx) error { return nil })
	resource.Get("/:resourceName", func(c *fiber.Ctx) error { return nil })
	resource.Post("/:resourceName", func(c *fiber.Ctx) error { return nil })
}

func main() {
	// Start a new fiber app
	app := fiber.New()

	app.Get("/", hello)
	setupRoutes(app)
	// Listen on PORT 8000
	app.Listen(":8888")
}
