package main

import (
	"fmt"
	"log"
	"os"
	"sanddb/anti_entropy"
	c "sanddb/config"
	"strconv"
	"time"

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

func setupRing(config c.Configurations) *anti_entropy.Ring {
	ring := &config.Ring
	ring.NodeMap = make(map[int64]*anti_entropy.Node)
	ring.ReplicationFactor = config.ReplicationFactor

	for _, node := range ring.Nodes {
		node.Hash = anti_entropy.GetHash(strconv.Itoa(node.Id))
		ring.NodeMap[node.Hash] = node
		ring.NodeHashes = append(ring.NodeHashes, node.Hash)
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
		fmt.Println("Please enter node ID as the first argument.")
		return
	}
	nodeID, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("Please enter a valid node ID.")
		return
	}
	node := &anti_entropy.Node{
		Id:        nodeID,
		IPAddress: config.Ring.Nodes[nodeID].IPAddress,
		Port:      config.Ring.Nodes[nodeID].Port,
		Hash:      anti_entropy.GetHash(strconv.Itoa(nodeID)),
	}
	fmt.Printf("Node #%d: Hash: %d", node.Id, node.Hash)
	// Initialize the Ring
	ring := setupRing(config)
	antiEntropyHandler := &anti_entropy.AntiEntropyHandler{
		Node: node,
		Ring: ring,
		// Repair timeout should be long enough, but not too long
		// In real life production systems with a large amount of data, this can take days or even weeks to fully complete
		// TODO: This should be used by the client
		RepairTimeout:          time.Duration(config.RepairTimeout) * time.Hour,
		InternalRequestTimeout: time.Duration(config.InternalRequestTimeout) * time.Second,
		GCGraceSeconds:         config.GCGraceSeconds,
	}
	ring.CurrentNode = node
	app.Get("/", hello)
	app.Post("/repair", antiEntropyHandler.HandleRepairRequest)
	app.Post("/full_repair", antiEntropyHandler.HandleFullRepairRequest)
	internalGroup := app.Group("/internal")
	internalGroup.Post("/repair/get_data", antiEntropyHandler.HandleRepairGetRequest)
	internalGroup.Post("/repair/write_data", antiEntropyHandler.HandleRepairWriteRequest)
	internalGroup.Post("/repair/trigger_delete", antiEntropyHandler.HandleRepairDeleteRequest)
	internalGroup.Post("/repair/missing_subrepair", antiEntropyHandler.HandleMissingSubrepairRequest)
	err = app.Listen(node.Port)
	if err != nil {
		log.Fatalf("Error in starting up server: %s", err)
	}
}
