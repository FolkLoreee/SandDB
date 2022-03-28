package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
	"log"
	"os"
	c "sanddb/config"
	"sanddb/strict_quorum"
	"strconv"
	"time"
)

func hello(c *fiber.Ctx) error {
	err := c.SendString("Henlo world with fiber")
	if err != nil {
		log.Fatalf("Error in hello world: %s", err)
	}
	return err
}

func setupRing(config c.Configurations) *strict_quorum.Ring {
	ring := &config.Ring
	ring.NodeMap = make(map[int64]*strict_quorum.Node)

	for _, node := range ring.Nodes {
		node.Hash = strict_quorum.GetHash(strconv.Itoa(node.Id))
		ring.NodeMap[node.Hash] = node
		ring.NodeHashes = append(ring.NodeHashes, node.Hash)
		ring.ReplicationFactor = 2
	}
	return ring
}

func main() {
	var (
		config c.Configurations
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
	//Get Node ID from args
	args := os.Args
	if len(args) == 1 {
		fmt.Println("Please enter node ID")
		return
	}
	nodeID, err := strconv.Atoi(args[1])
	//initialize a Node
	node := &strict_quorum.Node{
		Id:        nodeID,
		IPAddress: config.Ring.Nodes[nodeID].IPAddress,
		Port:      config.Ring.Nodes[nodeID].Port,
		Hash:      strict_quorum.GetHash(strconv.Itoa(nodeID)),
	}
	fmt.Printf("Node #%d: Hash: %d", node.Id, node.Hash)
	// Initialize the Ring
	ring := setupRing(config)

	requestHandler := &strict_quorum.Handler{
		Node:    node,
		Ring:    ring,
		Timeout: time.Duration(config.Timeout) * time.Second,
	}
	ring.CurrentNode = node
	app.Get("/", hello)
	app.Post("/request", requestHandler.HandleRequest)

	//quorumGroup := app.Group("/quorum")
	//quorumGroup.Post("/start", requestHandler.HandleQuorumRequest)

	internalGroup := app.Group("/internal")
	internalGroup.Post("/read", requestHandler.HandleCoordinatorRead)
	internalGroup.Post("/write", requestHandler.HandleCoordinatorWrite)

	//chashGroup := app.Group("/chash")
	//chashGroup.Post("/coordinate", requestHandler.HandleCoordinatorWrite)

	err = app.Listen(node.Port)
	if err != nil {
		log.Fatalf("Error in starting up server: %s", err)
	}
}
