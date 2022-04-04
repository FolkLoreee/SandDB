package read_write

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) HandleRequest(c *fiber.Ctx) error {
	var (
		clientRequest Request
	)
	err := c.BodyParser(&clientRequest)
	if err != nil {
		fmt.Printf("Error in parsing request: %s", err.Error())
		return err
	}
	h.Request = &clientRequest
	if clientRequest.Type == REQUEST_WRITE {
		if err = h.handleClientWriteRequest(); err != nil {
			_ = c.SendString(err.Error())
			return err
		} else {
			_ = c.SendStatus(http.StatusOK)
		}
	} else if clientRequest.Type == REQUEST_READ {
		data, err := h.handleClientReadRequest()
		if err != nil {
			_ = c.SendString(err.Error())
			return err
		} else {
			body, err := json.Marshal(data.Content)
			if err != nil {
				fmt.Printf("Error marshalling read request: %s", err.Error())
				return err
			}
			c.Status(http.StatusOK)
			_ = c.Send(body)
		}
	}
	return nil
}

// Send a PeerMessage to other nodes and inform that node is killed
func (h *Handler) SendKillRequest(node *Node) error {
	var responseMsg PeerMessage
	killRequest := PeerMessage{
		Type:     MessageType(KILL),
		Content:  "I am dead node",
		SourceID: h.Node.Id,
	}
	fmt.Printf("Sending kill message to node %d at port %s from node %d.\n", node.Id, node.Port, h.Node.Id)

	body, err := json.Marshal(killRequest)
	if err != nil {
		fmt.Printf("Error in marshalling kill request: %s\n", err.Error())
		return err
	}
	postBody := bytes.NewBuffer(body)

	response, err := http.Post(node.IPAddress+node.Port+"/kill", "application/json", postBody)

	if err != nil {
		fmt.Printf("Error in posting kill request: %s\n", err.Error())
		return err
	}
	defer response.Body.Close()
	jsonResponse, err := ioutil.ReadAll(response.Body)

	err = json.Unmarshal([]byte(jsonResponse), &responseMsg)
	if err != nil {
		return err
	}
	fmt.Println("Successfully informed kill.")
	return nil
}

func (h *Handler) HandleKillNode(c *fiber.Ctx) error {

	fmt.Println("Received kill request.")

	var requestMsg PeerMessage

	killResponse := &PeerMessage{
		Type:     MessageType(KILL_ACK),
		Content:  "Acknowledged dead node. Ring updated.",
		SourceID: h.Node.Id,
	}
	err := c.BodyParser(&requestMsg)
	if err != nil {
		return err
	}

	fmt.Print(requestMsg)

	// update ring
	h.UpdateRing(&requestMsg)

	resp, err := json.Marshal(killResponse)
	if err != nil {
		_ = c.SendStatus(http.StatusInternalServerError)
		return err
	}
	_ = c.Send(resp)
	fmt.Printf("Received request from node %d.\n", requestMsg.SourceID)
	fmt.Printf("Content received: %s.\n", requestMsg.Content)
	return nil
}
