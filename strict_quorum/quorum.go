package strict_quorum

import (
	"errors"
	"fmt"
	"math"
	"time"
)

func (h *Handler) createQuorum() error {
	h.Responses = 0
	h.Ring.MinVotes = int(math.Ceil(float64(h.Ring.ReplicationFactor / 2)))
	quorumChannel := make(chan PeerMessage)
	h.QuorumChannel = quorumChannel
	go h.collectReplies()
	return nil
}

func (h *Handler) closeQuorum() error {
	defer close(h.QuorumChannel)
	time.Sleep(h.Timeout)
	if h.Responses >= h.Ring.MinVotes {
		fmt.Println("Quorum passed!")
		fmt.Printf("Node %d: Number of votes received: %d\tNumber of votes required:%d\n", h.Node.Id, h.Responses, h.Ring.MinVotes)
		return nil
	} else {
		fmt.Println("Insufficient Quorum")
		fmt.Printf("Node %d: Number of votes received: %d\tNumber of votes required:%d\n", h.Node.Id, h.Responses, h.Ring.MinVotes)
		return errors.New("write failed: insufficient ACKs")
	}
}

func (h *Handler) collectReplies() {
	//TODO: channel should close and node should move on once quorum is reached
	node := h.Node
	for {
		select {
		case reply := <-h.QuorumChannel:
			if reply.Type == READ_OK {
				node.DataStore[reply.SourceID] = reply
			}
			h.Responses++
			fmt.Printf("Data received. Current ACKs: %d\n", h.Responses)
		case <-time.After(h.Timeout):
			fmt.Printf("\nTimeout\n")
			return
		}
	}
}