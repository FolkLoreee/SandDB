package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
	"log"
	"os"
	c "sanddb/config"
	"sanddb/db"
	"sanddb/read_write"
	"sanddb/utils"
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

func setupRing(config c.Configurations) *utils.Ring {
	ring := &config.Ring
	ring.NodeMap = make(map[int64]*utils.Node)
	ring.ReplicationFactor = config.ReplicationFactor

	for _, node := range ring.Nodes {
		node.Hash = utils.GetHash(strconv.Itoa(node.Id))
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
		fmt.Println("Please enter node ID")
		return
	}
	nodeID, err := strconv.Atoi(args[1])
	//initialize a Node
	node := &utils.Node{
		Id:        nodeID,
		IPAddress: config.Ring.Nodes[nodeID].IPAddress,
		Port:      config.Ring.Nodes[nodeID].Port,
		Hash:      utils.GetHash(strconv.Itoa(nodeID)),
	}
	fmt.Printf("Node #%d: Hash: %d", node.Id, node.Hash)
	// Initialize the Ring
	ring := setupRing(config)

	requestHandler := &read_write.Handler{
		Node:    node,
		Ring:    ring,
		Timeout: time.Duration(config.Timeout) * time.Second,
	}
	ring.CurrentNode = node
	app.Get("/", hello)
	//TODO: split between read and write request
	//app.Post("/request", requestHandler.HandleRequest)
	app.Post("/create", requestHandler.HandleClientCreateRequest)
	app.Post("/insert", requestHandler.HandleClientWriteRequest)
	app.Post("/read", requestHandler.HandleClientReadRequest)
	//internalGroup := app.Group("/internal")
	//internalGroup.Post("/read", requestHandler.HandleCoordinatorRead)
	//internalGroup.Post("/write", requestHandler.HandleCoordinatorWrite)

	dbHandler := &db.Handler{
		Node: node,
	}
	dbGroup := app.Group("/db")
	dbGroup.Post("/insert", dbHandler.HandleDBInsert)
	dbGroup.Post("/new", dbHandler.HandleCreateTable)
	dbGroup.Post("/read", dbHandler.HandleDBRead)
	err = app.Listen(node.Port)
	if err != nil {
		log.Fatalf("Error in starting up server: %s", err)
	}
}
