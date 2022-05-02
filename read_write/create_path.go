package read_write

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io/ioutil"
	"net/http"
)

func (h *Handler) HandleClientCreateRequest(c *fiber.Ctx) error {
	var (
		request CreateRequest
	)
	err := c.BodyParser(&request)
	if err != nil {
		return err
	}
	fmt.Printf("Request received from client by receiverNode %d.\n", h.Node.Id)
	err = h.createQuorum(REQUEST_CREATE)
	if err != nil {
		return err
	}
	for _, receiverNode := range h.Ring.Nodes {
		//if receiverNode.Id != h.Node.Id {
		err = h.sendCreateRequest(receiverNode, request)
		if err != nil {
			fmt.Printf("Error in sending coordinator request: %s\n", err.Error())
		}
		//}
	}
	err = h.closeQuorum()
	if err != nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, err.Error())
	}
	_ = c.SendStatus(http.StatusOK)
	return nil
}

func (h *Handler) sendCreateRequest(node *Node, data CreateRequest) error {
	var (
		responseMsg PeerMessage
	)
	fmt.Printf("Sending coordinator request to node with has %d\n", node.Hash)
	body, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Erorr in marshalling coordinator request: %s", err.Error())
		return err
	}
	postBody := bytes.NewBuffer(body)
	response, err := http.Post(node.IPAddress+node.Port+"/db/new", "application/json", postBody)

	if err != nil {
		fmt.Printf("Error in posting coordinator request: %s", err.Error())
		return err
	}
	defer response.Body.Close()
	jsonResponse, err := ioutil.ReadAll(response.Body)
	err = json.Unmarshal([]byte(jsonResponse), &responseMsg)
	if err != nil {
		fmt.Printf("Error in unmarshalling: %s\n", err.Error())
		return err
	}
	h.QuorumChannel <- responseMsg
	return nil
}