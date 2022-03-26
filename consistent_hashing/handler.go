package consistent_hashing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) HandleRequest(c *fiber.Ctx) error {
	var request Request
	err := c.BodyParser(&request)
	if err != nil {
		fmt.Printf("Error in parsing request: %s", err.Error())
		return err
	}
	h.Request = &request
	if request.Type == REQUEST_WRITE {
		if err = h.handleWriteRequest(); err != nil {
			_ = c.SendString(err.Error())
			return err
		}
		_ = c.SendStatus(http.StatusOK)
	}
	return nil
}

func (h *Handler) handleWriteRequest() error {
	fmt.Printf("Request received from client by node %d.\n", h.Node.Id)
	// Hash partition key sent by client
	partitionKey := h.Request.Content
	hashedPK := GetHash(partitionKey)
	fmt.Printf("Partition key %s hashed to %d\n", partitionKey, hashedPK)

	fmt.Println("Node positions (hashes) in the ring:")
	fmt.Println(h.Ring.NodeHashes)

	// Look for the node
	node := h.Ring.GetNode(partitionKey)
	fmt.Printf("Routing request to node %d at position %d...\n", node.Id, node.Hash)
	h.sendCoordinatorRequest(&node)

	if h.Ring.ReplicationFactor > 0 {
		fmt.Printf("Ring replication factor is %d.\n", h.Ring.ReplicationFactor)

		nodesToReplicateTo := h.Ring.Replicate(partitionKey)
		for _, replNode := range nodesToReplicateTo {
			fmt.Printf("Replicating to node with hash %d\n", replNode.Hash)
			h.sendCoordinatorRequest(replNode)
		}
	}

	return nil

}

func (h *Handler) sendCoordinatorRequest(node *Node) error {
	var (
		responseMsg PeerMessage
	)
	coordinatorRequest := PeerMessage{
		Type:     MessageType(COORDINATOR_REQUEST),
		Content:  h.Request.Content,
		SourceID: h.Node.Id,
	}
	fmt.Printf("Sending coordinator request to node with hash %d.\n", node.Hash)
	body, err := json.Marshal(coordinatorRequest)
	if err != nil {
		fmt.Printf("Error in marshalling coordinator request: %s", err.Error())
		return err
	}
	postBody := bytes.NewBuffer(body)
	response, err := http.Post(node.IPAddress+node.Port+"/chash/coordinate", "application/json", postBody)

	if err != nil {
		fmt.Printf("Error in posting coordinator request: %s", err.Error())
		return err
	}
	defer response.Body.Close()
	jsonResponse, err := ioutil.ReadAll(response.Body)

	err = json.Unmarshal([]byte(jsonResponse), &responseMsg)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully routed request: %s\n\n", string(jsonResponse))

	return nil
}
