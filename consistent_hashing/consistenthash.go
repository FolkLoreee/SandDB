package consistent_hashing

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"unsafe"

	"github.com/gofiber/fiber/v2"
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
	// index := sort.Search(len(r.NodeHashes), func(idx int) bool {
	// 	fmt.Println(r.NodeHashes[idx] >= hash)
	// 	fmt.Printf("%d >= %d\n", idx, hash)
	// 	return r.NodeHashes[idx] >= hash
	// })
	index := 0
	for idx, nodeHash := range r.NodeHashes {
		if hash <= nodeHash {
			index = idx
			break
		}
	}

	return index
}

func (r *Ring) GetNode(partitionKey string) Node {
	hash := GetHash(partitionKey)
	index := r.Search(hash)

	nodeHash := r.NodeHashes[index]
	// return *r.Nodes[nodeHash]
	return *r.NodeMap[nodeHash]
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

func (h *Handler) HandleCoordinatorRequest(c *fiber.Ctx) error {
	var (
		requestMsg PeerMessage
	)
	coordinatorResponse := &PeerMessage{
		Type:     MessageType(NODE_ACK),
		Content:  "1",
		SourceID: h.Node.Id,
	}
	err := c.BodyParser(&requestMsg)
	if err != nil {
		return err
	}
	resp, err := json.Marshal(coordinatorResponse)
	if err != nil {
		_ = c.SendStatus(http.StatusInternalServerError)
		return err
	}
	_ = c.Send(resp)
	fmt.Printf("Received request from node %d.\n", requestMsg.SourceID)
	fmt.Printf("Content received: %s.", requestMsg.Content)
	return nil
}
