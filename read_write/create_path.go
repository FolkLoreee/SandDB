package read_write

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io/ioutil"
	"net/http"
	"sanddb/messages"
	"sanddb/utils"
)

func (h *Handler) HandleClientCreateRequest(c *fiber.Ctx) error {
	var (
		request messages.CreateRequest
	)
	err := c.BodyParser(&request)
	if err != nil {
		return err
	}
	fmt.Printf("Request received from client by receiverNode %d.\n", h.Node.Id)
	err = h.createQuorum(messages.REQUEST_CREATE)
	if err != nil {
		return err
	}
	for _, receiverNode := range h.Ring.Nodes {
		err = h.sendCreateRequest(receiverNode, request)
		if err != nil {
			fmt.Printf("Error in sending coordinator request to %d: %s\n", receiverNode.Id, err.Error())
			return err
		}
	}
	err = h.closeQuorum()
	if err != nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, err.Error())
	}
	successMsg := fmt.Sprintf("Table %s has been successfully created!", request.TableName)
	_ = c.Status(http.StatusCreated).SendString(successMsg)
	return nil
}

func (h *Handler) sendCreateRequest(node *utils.Node, data messages.CreateRequest) error {
	var (
		responseMsg messages.PeerMessage
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
	if response.StatusCode != http.StatusCreated {
		return errors.New(string(jsonResponse))
	}
	err = json.Unmarshal([]byte(jsonResponse), &responseMsg)
	if err != nil {
		fmt.Printf("Error in unmarshalling: %s\n", err.Error())
		return err
	}
	h.QuorumChannel <- responseMsg
	return nil
}
