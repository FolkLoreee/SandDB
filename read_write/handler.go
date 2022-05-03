package read_write

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sanddb/messages"
	"sanddb/utils"

	"github.com/gofiber/fiber/v2"
)

// Send a PeerMessage to other nodes and inform that node is killed
func (h *Handler) SendKillRequest(node *utils.Node) error {
	var responseMsg messages.PeerMessage
	killRequest := messages.PeerMessage{
		Type:     messages.KILL,
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

	var requestMsg messages.PeerMessage
	killResponse := &messages.PeerMessage{
		Type:     messages.KILL_ACK,
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
	fmt.Print("NODE HASHES: ")
	fmt.Println(h.Ring.NodeHashes)
	return nil
}

func (h *Handler) HandleClientKillRequest(c *fiber.Ctx) error {
	fmt.Println("Received kill node request from client")
	//
	//var requestMsg messages.KillRequest
	//err := c.BodyParser(&requestMsg)
	//if err != nil {
	//	fmt.Println("Error parsing client kill request")
	//	return err
	//}
	//
	//fmt.Print(requestMsg)

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

	return nil
}

func (h *Handler) SendReviveRequest(node *utils.Node) error {
	var responseMsg messages.PeerMessage
	reviveRequest := messages.PeerMessage{
		Type:     messages.REVIVED,
		Content:  "Happy Easter!",
		SourceID: h.Node.Id,
	}
	fmt.Printf("Sending revive message to node %d at port %s from node %d.\n", node.Id, node.Port, h.Node.Id)

	body, err := json.Marshal(reviveRequest)
	if err != nil {
		fmt.Printf("Error in marshalling revive request: %s\n", err.Error())
		return err
	}
	postBody := bytes.NewBuffer(body)

	response, err := http.Post(node.IPAddress+node.Port+"/revive", "application/json", postBody)

	if err != nil {
		fmt.Printf("Error in posting revive request: %s\n", err.Error())
		return err
	}
	defer response.Body.Close()
	jsonResponse, err := ioutil.ReadAll(response.Body)

	err = json.Unmarshal([]byte(jsonResponse), &responseMsg)
	if err != nil {
		return err
	}
	fmt.Println("Successfully informed revive.")
	return nil
}

func (h *Handler) HandleReviveNode(c *fiber.Ctx) error {
	fmt.Println("Received revive request.")

	var requestMsg messages.PeerMessage
	reviveResponse := &messages.PeerMessage{
		Type:     messages.REVIVED_ACK,
		Content:  "Acknowledged revived node. Ring updated.",
		SourceID: h.Node.Id,
	}
	err := c.BodyParser(&requestMsg)
	if err != nil {
		return err
	}

	fmt.Println(requestMsg)

	fmt.Println(requestMsg.Type)
	// update ring
	h.UpdateRing(&requestMsg)

	resp, err := json.Marshal(reviveResponse)
	if err != nil {
		_ = c.SendStatus(http.StatusInternalServerError)
		return err
	}
	_ = c.Send(resp)
	fmt.Printf("Received request from node %d.\n", requestMsg.SourceID)
	fmt.Printf("Content received: %s.\n", requestMsg.Content)
	fmt.Print("NODE HASHES: ")
	fmt.Println(h.Ring.NodeHashes)
	return nil
}
