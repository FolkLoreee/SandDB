package main

import (
	"fmt"
	"log"
	"os"
	"sanddb/anti_entropy"
	c "sanddb/config"
	"strconv"
	"time"

	"os/signal"
	"sanddb/db"
	"sanddb/read_write"
	"sanddb/utils"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/spf13/viper"
)

func hello(c *fiber.Ctx) error {
	err := c.SendString("Henlo world with fiber")
	if err != nil {
		log.Fatalf("Error in hello world: %s", err)
	}
	return err
}
func gracefulShutdown(h *read_write.Handler) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	signal.Notify(s, syscall.SIGTERM)
	go func() {
		<-s
		fmt.Println("Shutting down gracefully.")
		// // update node status to dead
		// h.Node.Status = consistent_hashing.DEAD
		// h.Ring.NodeMap[h.Node.Hash] = h.Node

		// values := make([]*consistent_hashing.Node, 0, len(h.Ring.NodeMap))

		// for _, v := range h.Ring.NodeMap {
		// 	if v.Status == consistent_hashing.DEAD {
		// 		continue
		// 	}
		// 	values = append(values, v)
		// }
		// h.Ring.Nodes = values
		for _, node := range h.Ring.Nodes {
			if node.Id == h.Node.Id {
				fmt.Printf("Killed node %d.\n", node.Id)
				continue
			}
			fmt.Printf("Sending kill request to node %d.\n", node.Id)

			// inform nodes that this node is dead
			h.SendKillRequest(node)
		}
		os.Exit(0)
	}()
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

func informRevive(h *read_write.Handler) {
	ring := h.Ring

	for _, node := range ring.Nodes {
		if node.Id == h.Node.Id {
			fmt.Printf("Revive node %d.\n", node.Id)
			continue
		}
		fmt.Printf("Sending revive request to node %d.\n", node.Id)

		// Inform nodes that this node is revived
		h.SendReviveRequest(node)
	}

	fmt.Println("All nodes informed.")
}

func main() {
	var (
		config c.Configurations
	)
	app := fiber.New()

	app.Use(cors.New())

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
	// Inform of Node's existence
	informRevive(requestHandler)
	////Reading configuration files
	//viper.SetConfigFile("./config/config.yml")
	//if err := viper.ReadInConfig(); err != nil {
	//	fmt.Printf("Error reading config file: %s\n", err)
	//}
	//if err := viper.Unmarshal(&config); err != nil {
	//	fmt.Printf("Error decoding config file: %s\n", err)
	//}
	////Get Node ID from args
	//args := os.Args
	//if len(args) == 1 {
	//	fmt.Println("Please enter node ID as the first argument.")
	//	return
	//}
	//nodeID, err := strconv.Atoi(args[1])
	//if err != nil {
	//	fmt.Println("Please enter a valid node ID.")
	//	return
	//}
	//fmt.Printf("Node #%d: Hash: %d", node.Id, node.Hash)
	//// Initialize the Ring
	//ring := setupRing(config)
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
	//app.Post("/request", requestHandler.HandleRequest)
	app.Post("/create", requestHandler.HandleClientCreateRequest)
	app.Post("/insert", requestHandler.HandleClientWriteRequest)
	app.Post("/read", requestHandler.HandleClientReadRequest)
	//internalGroup := app.Group("/internal")
	//internalGroup.Post("/read", requestHandler.HandleCoordinatorRead)
	//internalGroup.Post("/write", requestHandler.HandleCoordinatorWrite)
	app.Post("/kill", requestHandler.HandleKillNode)
	app.Post("/killNode", requestHandler.HandleClientKillRequest)
	app.Post("/revive", requestHandler.HandleReviveNode)

	dbHandler := &db.Handler{
		Node: node,
	}
	dbGroup := app.Group("/db")
	dbGroup.Post("/insert", dbHandler.HandleDBInsert)
	dbGroup.Post("/new", dbHandler.HandleCreateTable)
	dbGroup.Post("/read", dbHandler.HandleDBRead)
	go gracefulShutdown(requestHandler)
	err = app.Listen(node.Port)
	if err != nil {
		log.Fatalf("Error in starting up server: %s", err)
	}
}
