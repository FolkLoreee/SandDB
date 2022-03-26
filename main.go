package main

import (
	"fmt"
	"log"
	"os"
	c "sanddb/config"
	"sanddb/consistent_hashing"
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

func setupRing(config c.Configurations) consistent_hashing.Ring {
	ring := config.Ring
	ring.NodeMap = make(map[int64]*consistent_hashing.Node)

	for _, node := range ring.Nodes {
		node.Hash = consistent_hashing.GetHash(strconv.Itoa(node.Id))
		ring.NodeMap[node.Hash] = node
		ring.NodeHashes = append(ring.NodeHashes, node.Hash)
		ring.ReplicationFactor = 2
	}
	return ring
}

func main() {
	var (
		config c.Configurations
		ring   consistent_hashing.Ring
	)
	app := fiber.New()
	//Reading configuration files
	viper.SetConfigFile("./config/config.yml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %s\n", err)
	}
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Error decoding config file: %s\n", err)
	}
	// Get Node ID from args
	args := os.Args
	if len(args) == 1 {
		fmt.Println("Please enter node ID")
		return
	}

	nodeID, err := strconv.Atoi(args[1])

	// Initialize a Node based on the args
	node := &consistent_hashing.Node{
		Id:        nodeID,
		IPAddress: config.Ring.Nodes[nodeID].IPAddress,
		Port:      config.Ring.Nodes[nodeID].Port,
		Hash:      consistent_hashing.GetHash(strconv.Itoa(nodeID)),
	}

	fmt.Printf("Node #%d: Hash: %d", node.Id, node.Hash)
	// Initialize the Ring
	ring = setupRing(config)

	fmt.Printf("Printing cluster: %v\n", ring.Nodes[1])
	requestHandler := &consistent_hashing.Handler{
		Node: node,
		Ring: &ring,
	}

	app.Get("/", hello)
	app.Post("/request", requestHandler.HandleRequest)

	chashGroup := app.Group("/chash")
	chashGroup.Post("/coordinate", requestHandler.HandleCoordinatorRequest)

	err = app.Listen(node.Port)
	if err != nil {
		log.Fatalf("Error in starting up server: %s", err)
	}
}
