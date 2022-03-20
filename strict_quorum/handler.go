package strict_quorum

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io/ioutil"
	"math"
	"net/http"
	"time"
)

func (h *Handler) HandleRequest(c *fiber.Ctx) error {
	var (
		request Request
	)
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
	h.Votes = 0
	h.Cluster.MinVotes = int(math.Ceil(float64(len(h.Cluster.Nodes) / 2)))
	voteChannel := make(chan string)
	h.VoteChannel = voteChannel
	go h.countVotes()
	for id := range h.Cluster.Nodes {
		if id != h.Node.Id {
			err := h.sendQuorumRequest(h.Cluster.Nodes[id])
			if err != nil {
				fmt.Printf("Error in sending quorum requests: %s", err.Error())
				//return err
			}
		}
	}
	time.Sleep(h.Timeout)
	if h.Votes >= h.Cluster.MinVotes {
		//TODO: Write data to local DB
		fmt.Println("Quorum passed!")
		fmt.Printf("Node %d: Number of votes received: %d\tNumber of votes required:%d\n", h.Node.Id, h.Votes, h.Cluster.MinVotes)
		fmt.Println("Writing to DB...")
	} else {
		fmt.Println("Insufficient Quorum")
		fmt.Printf("Node %d: Number of votes received: %d\tNumber of votes required:%d\n", h.Node.Id, h.Votes, h.Cluster.MinVotes)
	}
	defer close(h.VoteChannel)
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
	response, err := http.Post(node.IPAddress+node.Port+"/quorum/start", "application/json", postBody)

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
	return nil
}
func (h *Handler) countVotes() {
	for {
		select {
		case msg := <-h.VoteChannel:
			if msg == "1" {
				h.Votes++
				fmt.Printf("Vote received. Current votes: %d\n", h.Votes)
			}
		case <-time.After(h.Timeout):
			fmt.Printf("Timeout\n")
			return
		}
	}
}
