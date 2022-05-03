package utils

import (
	"fmt"
)

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
