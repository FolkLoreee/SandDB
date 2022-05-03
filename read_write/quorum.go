package read_write

import (
	"errors"
	"fmt"
	"math"
	. "sanddb/messages"
	"time"
)

func (h *Handler) createQuorum(requestType RequestType) error {
	h.Responses = 0
	//Create Request has to be replicated to all nodes, not just replicas
	if requestType == REQUEST_CREATE {
		h.Ring.MinVotes = int(math.Ceil(float64(len(h.Ring.Nodes) / 2)))
	} else {
		h.Ring.MinVotes = int(math.Ceil(float64(h.Ring.ReplicationFactor / 2)))
	}
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
		return errors.New("insufficient ACKs")
	}
}

func (h *Handler) collectReplies() {
	//TODO: channel should close and node should move on once quorum is reached
	for {
		select {
		case _ = <-h.QuorumChannel:
			h.Responses++
			fmt.Printf("Data received. Current ACKs: %d\n", h.Responses)
		case <-time.After(h.Timeout):
			fmt.Printf("\nTimeout\n")
			return
		}
	}
}
