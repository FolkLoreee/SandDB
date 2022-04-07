package read_write

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func (h *Handler) handleClientReadRequest() (PeerMessage, error) {
	var latestVersion PeerMessage

	node := h.Node
	ring := h.Ring

	replyStore := map[int]PeerMessage{}
	node.DataStore = replyStore

	fmt.Println("=============================================")
	fmt.Println("Node", node.Id, "handling read request")
	fmt.Println("=============================================")

	//TODO (after storage system is done): instead of partition key, content of the request will be the client's query
	// Look for the receiverNode
	partitionKey := h.Request.Content

	receiverNode := h.Ring.GetNode(partitionKey)
	fmt.Printf("Routing request to receiverNode %d at position %d...\n", receiverNode.Id, receiverNode.Hash)
	//TODO: integrate with quorum to check if the number of responses received satisfies quorum
	err := h.createQuorum(REQUEST_READ)
	if err != nil {
		fmt.Printf("Error in creating quorum: %s", err.Error())
		return PeerMessage{}, err
	}
	err = h.sendReadRequest(receiverNode)
	if err != nil {
		fmt.Printf("Error in sending coordinator request: %s", err.Error())
		return PeerMessage{}, err
	}
	fmt.Printf("Ring replication factor is %d.\n", h.Ring.ReplicationFactor)

	replicas := h.Ring.Replicate(partitionKey)
	fmt.Printf("Replica content: %v\n", replicas)

	for _, receivingNode := range replicas {
		fmt.Println("Sending request to node", receivingNode.Id)
		err := h.sendReadRequest(receivingNode)
		if err != nil {
			fmt.Printf("Error sending read request: %s", err.Error())
			return PeerMessage{}, err
		}
	}
	err = h.closeQuorum()
	if err != nil {
		fmt.Printf("Closing quorum error: %s", err.Error())
		return PeerMessage{}, err
	}

	if len(node.DataStore) > 1 {
		// fmt.Println("Data collected from nodes:", node.DataStore)
		for _, data := range node.DataStore {
			if data.Version > latestVersion.Version {
				latestVersion = data
			}
		}
		fmt.Println("Data received:", latestVersion.Content, ", Version:", latestVersion.Version)

		//Change the message type to READ_REPAIR as it is handled differently
		latestVersion.Type = READ_REPAIR
		latestVersion.SourceID = node.Id
		for _, data := range node.DataStore {
			if data.Version < latestVersion.Version {
				fmt.Printf("Sending read repair to node %d\n", data.SourceID)
				err := h.sendWriteRequest(ring.Nodes[data.SourceID], latestVersion)
				if err != nil {
					return PeerMessage{}, err
				}
			}
		}
	} else {
		fmt.Println("No data received from nodes")
		return PeerMessage{}, errors.New("read fail: insufficient responses for quorum")
	}

	node.DataStore = map[int]PeerMessage{}

	return latestVersion, nil
}

//TODO: parameter will require the request query
func (h *Handler) sendReadRequest(receivingNode *Node) error {
	var reply PeerMessage
	node := h.Node
	//TODO: content will be populated with request query
	readRequest := PeerMessage{
		Type:     COORDINATOR_READ,
		Version:  0,
		Content:  "0",
		SourceID: node.Id,
	}

	body, err := json.Marshal(readRequest)
	if err != nil {
		fmt.Printf("Error marshalling read request: %s", err.Error())
		return err
	}

	postBody := bytes.NewBuffer(body)
	response, err := http.Post(receivingNode.IPAddress+receivingNode.Port+"/internal/read", "application/json", postBody)
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
	h.QuorumChannel <- reply
	return nil
}
func (h *Handler) HandleCoordinatorRead(c *fiber.Ctx) error {
	var (
		requestMsg PeerMessage
	)
	err := c.BodyParser(&requestMsg)
	if err != nil {
		return err
	}
	fmt.Println("Received read requestMsg from node", requestMsg.SourceID)
	//TODO: parse the actual read request here
	node := h.Node
	reply := &PeerMessage{
		Type:     READ_OK,
		SourceID: node.Id,
	}
	reply.Version, reply.Content = getData()

	body, err := json.Marshal(reply)
	_ = c.Send(body)
	fmt.Printf("Sent: %v\n", reply)
	return err
}

func getData() (int, string) {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(10), strconv.Itoa(rand.Intn(100))
}
