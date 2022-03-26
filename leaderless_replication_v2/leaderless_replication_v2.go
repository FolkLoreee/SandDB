package leaderless_replication_v2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"

	// "time"

	"github.com/gofiber/fiber/v2"
)

type RequestType int
type MessageType int

const (
	REQUEST_WRITE RequestType = iota
	REQUEST_READ
)

const (
	READ_OK MessageType = iota
	WRITE_OK
)

type Client struct {
	clientID int
}

func (r RequestType) String() string {
	return [...]string{"Write", "Read"}[r]
}

type Request struct {
	Type    RequestType `json:"type"`
	Content string      `json:"content"`
}

type Reply struct {
	Type    MessageType `json:"type"`
	Content Data        `json:"content"`
}

type Node struct {
	NodeID        int      `json:"id"`
	NodeNetwork   *Network `json:"node_network"`
	IsCoordinator bool     `json:"is_coordinator"`
	IpAddress     string   `json:"node_ip"`
	Port          string   `json:"port"`
}

type Network struct {
	Nodes       []*Node `json:"nodes" yaml:"nodes"`
	Coordinator *Node   `json:"coordinator_node" yaml:"coordinator_node"`
}

type Data struct {
	topic   string
	version int
	message int
}

// helper functions
func contains(dataStore []Data, d Data) bool {
	for _, data := range dataStore {
		if d == data {
			return true
		}
	}
	return false
}

// helper functions
func (requestingNode *Node) request(reqType RequestType, content string, node *Node) (error, Reply) {

	var reply Reply

	// construct Request object
	request := Request{
		Type:    reqType,
		Content: content,
	}

	// marshall request into json
	body, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("Error marshalling request: %s", err.Error())
		return err, reply
	}

	// send requestjson to node
	postBody := bytes.NewBuffer(body)
	response, err := http.Post(node.IpAddress+node.Port+"/getLastWrite", "application/json", postBody)
	if err != nil {
		fmt.Printf("Error posting request: %s", err.Error())
		return err, reply
	}

	// read response from node
	defer response.Body.Close()
	jsonResponse, err := ioutil.ReadAll(response.Body)
	err = json.Unmarshal([]byte(jsonResponse), &reply)
	if err != nil {
		return err, reply
	}

	// return reply
	return nil, reply
}

// node functions

func (node *Node) handleRequest(c *fiber.Ctx) error {
	var request Request

	// parse request message
	err := c.BodyParser(&request)
	if err != nil {
		fmt.Printf("Error parsing node request: %s", err.Error())
		return err
	}

	// check request type
	switch request.Type {

	// if read request
	case REQUEST_READ:
		if node.IsCoordinator {
			err := node.getLastWrite()
			if err != nil {
				return err
			}
		} else {
			err = node.replyRequest(READ_OK,
				Data{
					topic:   "",
					version: rand.Intn(5),
					message: rand.Intn(100),
				}, c)
			if err != nil {
				return err
			}
		}

	// if write request
	case REQUEST_WRITE:
		err = node.replyRequest(WRITE_OK, Data{}, c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (coordinator *Node) getLastWrite() error {
	var (
		dataStore     map[int]Data
		latestVersion Data
	)

	for _, node := range coordinator.NodeNetwork.Nodes {
		err, reply := node.request(REQUEST_READ, "", node)
		if err != nil {
			return err
		}

		dataStore[node.NodeID] = reply.Content
	}

	// if replyStore is not empty
	if len(dataStore) > 0 {
		// get latest version of data
		for _, data := range dataStore {
			if data.version > latestVersion.version {
				latestVersion = data
			}
		}

		// send reply back to client
		fmt.Println("Latest version of data obtained:", latestVersion)
	}

	return nil
}

func (node *Node) replyRequest(msgType MessageType, content Data, c *fiber.Ctx) error {

	// construct Reply object2
	reply := Reply{
		Type:    msgType,
		Content: content,
	}

	// marshall reply to json
	body, err := json.Marshal(reply)
	if err != nil {
		fmt.Println("Error marshalling reply from node", node.NodeID, ":", err.Error())
		return err
	}

	// send replyjson
	postBody := bytes.NewBuffer(body)
	_ = c.Send(postBody.Bytes())

	return nil
}

func (node *Node) sendReadRequest(c *fiber.Ctx) error {

	// create request
	var coordinatorNode *Node
	for _, node := range node.NodeNetwork.Nodes {
		if node.IsCoordinator {
			coordinatorNode = node
		}
	}
	err, _ := node.request(REQUEST_READ, "", coordinatorNode)
	if err != nil {
		fmt.Println("Error sending read request from coordinator to node")
		return err
	}

	return nil
}

func (requestingNode *Node) sendWriteRequest(node *Node) error {

	// send WRITE request to node
	err, reply := requestingNode.request(REQUEST_WRITE, "", node)
	if err != nil {
		return err
	}

	fmt.Println("Successfully updated Node", node.NodeID, "with", reply.Content)

	return nil
}

func (node *Node) HandleClientRequest(c *fiber.Ctx) error {
	var request Request

	// parse request message
	err := c.BodyParser(&request)
	if err != nil {
		fmt.Printf("Error parsing request: %s", err.Error())
		return err
	}

	if request.Type == REQUEST_READ {
		err := node.sendReadRequest(c)
		if err != nil {
			fmt.Printf("Error sending read request from client: %s", err.Error())
			return err
		}
	}

	return nil
}
