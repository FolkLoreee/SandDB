package strict_quorum

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io/ioutil"
	"net/http"
)

func (h *Handler) handleClientWriteRequest() error {
	fmt.Printf("Request received from client by receiverNode %d.\n", h.Node.Id)
	// Hash partition key sent by client
	partitionKey := h.Request.Content
	hashedPK := GetHash(partitionKey)
	fmt.Printf("Partition key %s hashed to %d\n", partitionKey, hashedPK)

	fmt.Println("Node positions (hashes) in the ring:")
	fmt.Println(h.Ring.NodeHashes)

	// Look for the receiverNode
	receiverNode := h.Ring.GetNode(partitionKey)
	fmt.Printf("Routing request to receiverNode %d at position %d...\n", receiverNode.Id, receiverNode.Hash)
	data := PeerMessage{
		Type:     COORDINATOR_WRITE,
		Version:  0,
		Content:  h.Request.Content,
		SourceID: h.Node.Id,
	}
	go h.collectReplies()

	err := h.createQuorum()
	if err != nil {
		return err
	}

	err = h.sendWriteRequest(receiverNode, data)
	if err != nil {
		fmt.Printf("Error in sending coordinator request: %s", err.Error())
		return err
	}
	fmt.Printf("Ring replication factor is %d.\n", h.Ring.ReplicationFactor)

	nodesToReplicateTo := h.Ring.Replicate(partitionKey)
	for _, replNode := range nodesToReplicateTo {
		fmt.Printf("Replicating to node with hash %d\n", replNode.Hash)
		err = h.sendWriteRequest(replNode, data)
		if err != nil {
			fmt.Printf("Error in sending coordinator replication message: %s", err.Error())
			return err
		}
	}
	err = h.closeQuorum()
	if err != nil {
		return err
	}

	return nil
}
func (h *Handler) sendWriteRequest(node *Node, data PeerMessage) error {
	var (
		responseMsg PeerMessage
	)
	fmt.Printf("Sending coordinator request to node with hash %d.\n", node.Hash)
	body, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error in marshalling coordinator request: %s", err.Error())
		return err
	}
	postBody := bytes.NewBuffer(body)
	response, err := http.Post(node.IPAddress+node.Port+"/request/writeNodeData", "application/json", postBody)

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

	//READ_REPAIR does not require quorum
	if data.Type == COORDINATOR_WRITE {
		h.QuorumChannel <- responseMsg
	}
	fmt.Printf("Successfully routed request: %s\n\n", string(jsonResponse))

	return nil
}

func (h *Handler) HandleCoordinatorWrite(c *fiber.Ctx) error {
	var (
		requestMsg PeerMessage
	)
	reply := &PeerMessage{
		Type:     WRITE_ACK,
		Content:  "1",
		SourceID: h.Node.Id,
	}
	err := c.BodyParser(&requestMsg)
	if err != nil {
		return err
	}
	resp, err := json.Marshal(reply)
	if err != nil {
		_ = c.SendStatus(http.StatusInternalServerError)
		return err
	}
	_ = c.Send(resp)
	fmt.Printf("Received request from node %d.\n", requestMsg.SourceID)
	fmt.Printf("Content received: %s.\n", requestMsg.Content)
	return nil
}
