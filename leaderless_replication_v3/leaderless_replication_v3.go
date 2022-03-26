package leaderless_replication_v3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

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

func (r RequestType) String() string {
	return [...]string{"Write", "Read"}[r]
}

type Request struct {
	Type     RequestType `json:"type"`
	Content  int         `json:"content"`
	SourceID int         `json:"node_id"`
}

type Reply struct {
	Type     MessageType `json:"type"`
	Version  int         `json:"version"`
	Content  int         `json:"content"`
	SourceID int         `json:"node_id"`
}

type Node struct {
	DataChannel chan Reply
	DataStore   map[int]Reply
	NodeID      int    `json:"id"`
	IpAddress   string `json:"node_ip"`
	Port        string `json:"port"`
}

type Network struct {
	Nodes       []*Node `json:"nodes"`
	CurrentNode *Node   `json:"current_node"`
}

func (network *Network) HandleRequest(c *fiber.Ctx) error {
	var request Request

	err := c.BodyParser(&request)
	if err != nil {
		fmt.Println("Error parsing request at node", network.CurrentNode.NodeID, err.Error())
		return err
	}

	if request.Type == REQUEST_READ {
		if err := network.CurrentNode.handleReadRequest(network); err != nil {
			_ = c.SendString(err.Error())
			return err
		}
		_ = c.SendStatus(http.StatusOK)
	}

	return nil
}

func (node *Node) handleReadRequest(network *Network) error {
	var latestVersion Reply

	replyChannel := make(chan Reply)
	node.DataChannel = replyChannel
	replyStore := map[int]Reply{}
	node.DataStore = replyStore

	fmt.Println("=============================================")
	fmt.Println("Node", node.NodeID, "handling read request")
	fmt.Println("=============================================")

	go node.collectReplies()
	for _, receivingNode := range network.Nodes {
		if receivingNode.NodeID == node.NodeID {
			continue
		}
		fmt.Println("Sending request to node", receivingNode.NodeID)
		err := node.sendReadRequest(receivingNode)
		if err != nil {
			fmt.Printf("Error sending read request: %s", err.Error())
		}
	}
	time.Sleep(time.Second)

	if len(node.DataStore) > 0 {
		// fmt.Println("Data collected from nodes:", node.DataStore)
		for _, data := range node.DataStore {
			if data.Version > latestVersion.Version {
				latestVersion = data
			}
		}

		for _, data := range node.DataStore {
			if data.Version < latestVersion.Version {
				err := node.sendWriteRequest(network.Nodes[data.SourceID], data.Content)
				if err != nil {
					return err
				}
			}
		}
		fmt.Println("Data received:", latestVersion.Content, ", Version:", latestVersion.Version)
	} else {
		fmt.Println("No data received from nodes")
	}

	node.DataStore = map[int]Reply{}

	defer close(node.DataChannel)
	return nil
}

func (node *Node) sendReadRequest(receivingNode *Node) error {
	var reply Reply

	readRequest := Request{
		Type:     REQUEST_READ,
		Content:  0,
		SourceID: node.NodeID,
	}

	body, err := json.Marshal(readRequest)
	if err != nil {
		fmt.Printf("Error marshalling read request: %s", err.Error())
		return err
	}

	postBody := bytes.NewBuffer(body)
	response, err := http.Post(receivingNode.IpAddress+receivingNode.Port+"/readNodeData", "application/json", postBody)
	if err != nil {
		fmt.Printf("Error posting read request: %s", err.Error())
		return err
	}
	defer response.Body.Close()

	jsonResponse, err := ioutil.ReadAll(response.Body)
	err = json.Unmarshal([]byte(jsonResponse), &reply)
	if err != nil {
		return nil
	}
	fmt.Println("Response received:", &reply)

	node.DataChannel <- reply
	return nil
}

func (node *Node) collectReplies() {
	for {
		select {
		case reply := <-node.DataChannel:
			node.DataStore[reply.SourceID] = reply
		case <-time.After(time.Second):
			fmt.Printf("\nTimeout\n")
			return
		}
	}
}

func (node *Node) HandleNodeRequest(c *fiber.Ctx) error {
	var request Request

	err := c.BodyParser(&request)
	if err != nil {
		return err
	}

	if request.Type == REQUEST_READ {
		fmt.Println("Received read request from node", request.SourceID)
		reply := &Reply{
			Type:     READ_OK,
			SourceID: node.NodeID,
		}
		fmt.Println(getData())
		reply.Version, reply.Content = getData()

		body, err := json.Marshal(reply)
		if err != nil {
			_ = c.SendStatus(http.StatusInternalServerError)
			return err
		}

		_ = c.Send(body)
		fmt.Println("Reply sent from", node.NodeID, ":", reply)
	} else if request.Type == REQUEST_WRITE {
		fmt.Println("Received write request from node", request.SourceID)
		if err := node.handleWriteRequest(); err != nil {
			_ = c.SendString(err.Error())
			return err
		}
		_ = c.SendStatus(http.StatusOK)
	}

	return nil
}

func getData() (int, int) {
	return rand.Intn(10), rand.Intn(100)
}

func (node *Node) sendWriteRequest(receivingNode *Node, content int) error {
	writeRequest := Request{
		Type:     REQUEST_WRITE,
		Content:  content,
		SourceID: node.NodeID,
	}

	body, err := json.Marshal(writeRequest)
	if err != nil {
		fmt.Printf("Error marshalling write request: %s", err.Error())
		return err
	}

	postBody := bytes.NewBuffer(body)
	response, err := http.Post(receivingNode.IpAddress+receivingNode.Port+"writeRequest", "application/json", postBody)
	if err != nil {
		fmt.Printf("Error posting write request: %s", err.Error())
		return err
	}
	defer response.Body.Close()

	// TODO: figure out how to get http response

	return nil
}

func (node *Node) handleWriteRequest() error {
	// TODO: write request to local DB

	return nil
}
