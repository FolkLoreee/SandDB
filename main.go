package main

import (
	"fmt"
	"log"
	"os"
	c "sanddb/config"
	"sanddb/leaderless_replication_v3"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

func hello(c *fiber.Ctx) error {
	err := c.SendString("Henlo world with fiber")
	if err != nil {
		log.Fatalf("Error in hello world: %s", err)
	}
	return err
}

func main() {
	var (
		config  c.Configurations
		network leaderless_replication_v3.Network
	)
	app := fiber.New()

	viper.SetConfigFile("./config/config.yml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %s\n", err)
	}
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Error decoding config file %s\n", err)
	}

	args := os.Args
	if len(args) == 1 {
		fmt.Println("Please enter node ID")
		return
	}

	node_id, err := strconv.Atoi(args[1])
	// node := config.Cluster.Nodes[node_id]
	node := &leaderless_replication_v3.Node{
		NodeID:    node_id,
		IpAddress: config.Cluster.Nodes[node_id].IpAddress,
		Port:      config.Cluster.Nodes[node_id].Port,
	}
	config.Cluster.CurrentNode = node

	// fmt.Println(config)
	network = config.Cluster

	app.Get("/", hello)
	app.Post("/readRequest", network.HandleRequest)
	app.Post("/readNodeData", node.HandleNodeRequest)

	errorororor := app.Listen(node.Port)
	if err != nil {
		log.Fatalf("Error in starting up server: %s", errorororor)
	}
}
