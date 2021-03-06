package read_write

import (
	"fmt"
	"sanddb/messages"
	"sanddb/utils"
	"strconv"
)

func (h *Handler) UpdateRing(reqMsg *messages.PeerMessage) {

	if reqMsg.Type == messages.KILL {
		fmt.Printf("Removing node %d from ring (len=%d).\n", reqMsg.SourceID, len(h.Ring.NodeHashes))

		// get node
		hash := utils.GetHash(strconv.Itoa(reqMsg.SourceID))
		nodeToRemove := h.Ring.NodeMap[hash]
		// nodeToRemove := h.Ring.GetNode(strconv.Itoa(reqMsg.SourceID))
		// update node status to dead
		nodeToRemove.Status = utils.DEAD
		fmt.Printf("Node %d status is %s. hash: %d\n", nodeToRemove.Id, nodeToRemove.Status.String(), nodeToRemove.Hash)

		// update ring: []NodeHashes, NodeMap, []Nodes
		// remove from node hashes
		h.Ring.NodeHashes = utils.RemoveNodeHash(h.Ring.NodeHashes, nodeToRemove.Hash)
		// update node status in NodeMap
		h.Ring.NodeMap[nodeToRemove.Hash] = nodeToRemove
		// update node in array of nodes

		for _, node := range h.Ring.NodeMap {
			fmt.Printf("Node %d: %s\n", node.Id, node.Status.String())
		}
	} else if reqMsg.Type == messages.REVIVED {
		fmt.Printf("Adding node %d to ring (len=%d).\n", reqMsg.SourceID, len(h.Ring.NodeHashes))

		// get node
		hash := utils.GetHash(strconv.Itoa(reqMsg.SourceID))
		nodeToAdd := h.Ring.NodeMap[hash]

		// update node status to alive
		nodeToAdd.Status = utils.ALIVE

		// update ring: []NodeHashes, NodeMap, []Nodes
		// add to node hashes
		h.Ring.NodeHashes = utils.AddNodeHash(h.Ring.NodeHashes, nodeToAdd.Hash)
		// update node status in NodeMap
		h.Ring.NodeMap[nodeToAdd.Hash] = nodeToAdd

		// update node in array of nodes
		for _, node := range h.Ring.NodeMap {
			fmt.Printf("Node %d: %s\n", node.Id, node.Status.String())
		}
	}
}
