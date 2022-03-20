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
func main() {
	var (
		config  c.Configurations
		cluster strict_quorum.Cluster
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
		IPAddress: config.Cluster.Nodes[nodeID].IPAddress,
		Port:      config.Cluster.Nodes[nodeID].Port,
	}
	//initialize a Cluster
	cluster = config.Cluster
	fmt.Printf("Printing cluster: %v\n", cluster.Nodes[1])
	requestHandler := &strict_quorum.Handler{
		Node:    node,
		Cluster: &cluster,
		Timeout: time.Duration(config.Timeout) * time.Second,
	}
	app.Get("/", hello)
	app.Post("/request", requestHandler.HandleRequest)

	quorumGroup := app.Group("/quorum")
	quorumGroup.Post("/start", requestHandler.HandleQuorumRequest)

	err = app.Listen(node.Port)
	if err != nil {
		log.Fatalf("Error in starting up server: %s", err)
	}
}
