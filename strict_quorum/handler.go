package strict_quorum

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io/ioutil"
	"math"
	"net/http"
)

func (h *Handler) HandleRequest(c *fiber.Ctx) error {
	var (
		request Request
	)
	err := c.BodyParser(request)
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
	var (
		quorumVotes int
		minVotes    int
	)
	minVotes = int(math.Ceil(float64(len(h.Cluster.Nodes) / 2)))
	voteChannel := make(chan string)
	h.VoteChannel = voteChannel
	go h.countVotes()
	for i := range h.Cluster.Nodes {
		err := h.sendQuorumRequest(h.Cluster.Nodes[i])
		if err != nil {
			fmt.Printf("Error in sending quorum requests: %s", err.Error())
			return err
		}
	}
	if quorumVotes >= minVotes {
		//TODO: Write data to local DB
	}
	return nil
}
func (h *Handler) handleReadRequest() (error, Data) {
	//TODO: Find the "closest replica"
	//TODO: return data
	return nil, Data{}
}

func (h *Handler) sendQuorumRequest(node *Node) error {
	var (
		responseMsg PeerMessage
	)
	h.Votes = 0
	quorumRequest := PeerMessage{
		Type:     QUORUM_REQUEST,
		Content:  "",
		SourceID: h.Node.Id,
	}
	body, err := json.Marshal(quorumRequest)
	if err != nil {
		fmt.Printf("Error in marshalling quorum request: %s", err.Error())
		return err
	}
	postBody := bytes.NewBuffer(body)
	response, err := http.Post(node.IPAddress, "application/json", postBody)

	if err != nil {
		fmt.Printf("Error in posting quorum request: %s", err.Error())
		return err
	}
	defer response.Body.Close()
	jsonResponse, err := ioutil.ReadAll(response.Body)

	err = json.Unmarshal([]byte(jsonResponse), &responseMsg)
	if err != nil {
		return err
	}
	h.VoteChannel <- responseMsg.Content
	defer close(h.VoteChannel)
	return nil
}
func (h *Handler) countVotes() {
	for {
		select {
		case msg := <-h.VoteChannel:
			if msg == "1" {
				h.Votes++
			}
		}
	}
}
