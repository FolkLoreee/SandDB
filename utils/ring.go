package utils

type Node struct {
	//DataStore is for coordinator to store responses from other nodes before being sent back to the client
	//DataStore map[int]PeerMessage
	Id        int    `json:"id"`
	IPAddress string `json:"ip_address"`
	Port      string `json:"port"`
	Hash      int64  `json:"hash"`
}

//Ring consists of multiple Nodes
type Ring struct {
	Nodes             []*Node         `json:"nodes" yaml:"nodes"`
	MinVotes          int             `json:"min_votes" yaml:"min_votes"`
	CurrentNode       *Node           `json:"current_node"`
	NodeMap           map[int64]*Node `json:"nodeMap"`
	NodeHashes        []int64         `json:"nodeHashes"`
	ReplicationFactor int             `json:"replication_factor"` // replicates at N-1 nodes
}
