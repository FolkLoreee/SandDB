package main

import (
	"cassie/strict_quorum"
	"github.com/gofiber/fiber/v2"
	"log"
	"math/rand"
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
	nodeID := rand.Intn(10000)
	//initialize a Node
	node := &strict_quorum.Node{
		Id:    nodeID,
		Clock: 0,
	}

	requestHandler := &strict_quorum.Handler{
		Node: node,
	}
	app.Get("/", hello)
	app.Get("/request", requestHandler.HandleRequest)
	err := app.Listen(":8000")
	if err != nil {
		log.Fatalf("Error in starting up server: %s", err)
	}
}
