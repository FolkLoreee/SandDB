package read_write

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strconv"
	"unsafe"
)

func ByteArrayToInt(arr []byte) int64 {
	val := int64(0)
	size := len(arr)
	for i := 0; i < size; i++ {
		*(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&val)) + uintptr(i))) = arr[i]
	}
	return val
}

func GetHash(id string) int64 {
	data := []byte(id)
	hash := md5.Sum(data)
	return ByteArrayToInt(hash[:])
}

func Sort(int64Values []int64) []int64 {
	out := make([]int64, len(int64Values))
	int64AsIntValues := make([]int, len(int64Values))

	for i, val := range int64Values {
		int64AsIntValues[i] = int(val)
	}

	sort.Ints(int64AsIntValues)

	for i, val := range int64AsIntValues {
		out[i] = int64(val)
	}

	return out
}

func Int64ToInt(int64Values []int64) []int {
	int64AsIntValues := make([]int, len(int64Values))

	for i, val := range int64Values {
		int64AsIntValues[i] = int(val)
	}
	return int64AsIntValues
}

func (r *Ring) Search(hash int64) int {
	index := 0
	for idx, nodeHash := range r.NodeHashes {
		if hash <= nodeHash {
			index = idx
			break
		}
	}

	return index
}

func (r *Ring) GetNode(partitionKey string) *Node {
	hash := GetHash(partitionKey)
	index := r.Search(hash)

	nodeHash := r.NodeHashes[index]
	return r.NodeMap[nodeHash]
}

func (r *Ring) Replicate(partitionKey string) []*Node {
	nodesToReplicateTo := []*Node{}
	hash := GetHash(partitionKey)
	index := r.Search(hash)
	fmt.Printf("Replicating from node with hash %d\n", hash)

	// replicated nodes
	fmt.Println("Nodes to replicate to:")
	for i := 1; i < r.ReplicationFactor; i++ {
		replIdx := (index + i) % len(r.NodeHashes)
		fmt.Println(replIdx)
		nodeHash := r.NodeHashes[replIdx]
		node := r.NodeMap[nodeHash]
		nodesToReplicateTo = append(nodesToReplicateTo, node)
		fmt.Println(node.Hash)
	}

	return nodesToReplicateTo
}

// func (r *Ring) AddNode(node *Node) {
// 	r.Nodes[node.Hash] = node
// 	nodeHashes := append(r.NodeHashes, node.Hash)
// 	sortedNodeHashes := Sort(nodeHashes)
// 	r.NodeHashes = sortedNodeHashes
// }

// func (r *Ring) RemoveNode(node *Node) {
// 	index := r.Search(node.Hash)
// 	indexplus := index + 1
// 	nodeHashes := append(r.NodeHashes[:index], r.NodeHashes[index+1:])
// 	r.NodeHashes = nodeHashes
// 	delete(r.Nodes, nodeHashes[index])
// }

func (h *Handler) UpdateRing(reqMsg *PeerMessage) {
	fmt.Printf("Removing node %d from ring (len=%d).\n", reqMsg.SourceID, len(h.Ring.NodeHashes))

	// get node
	hash := GetHash(strconv.Itoa(reqMsg.SourceID))
	nodeToRemove := h.Ring.NodeMap[hash]
	// nodeToRemove := h.Ring.GetNode(strconv.Itoa(reqMsg.SourceID))
	// update node status to dead
	nodeToRemove.Status = DEAD
	fmt.Printf("Node %d status is %s. hash: %d\n", nodeToRemove.Id, nodeToRemove.Status.String(), nodeToRemove.Hash)

	// update ring: []NodeHashes, NodeMap, []Nodes
	// remove from node hashes
	h.Ring.NodeHashes = RemoveNodeHash(h.Ring.NodeHashes, nodeToRemove.Hash)
	// update node status in NodeMap
	h.Ring.NodeMap[nodeToRemove.Hash] = nodeToRemove
	// update node in array of nodes

	for _, node := range h.Ring.NodeMap {
		fmt.Printf("Node %d: %s\n", node.Id, node.Status.String())
	}
}

func RemoveNodeHash(nodeHashes []int64, hash int64) []int64 {
	newNodeHashes := make([]int64, len(nodeHashes))
	for _, nHash := range nodeHashes {
		if nHash == hash {
			continue
		}
		newNodeHashes = append(newNodeHashes, nHash)
	}
	return newNodeHashes
}
