package read_write

import (
	"time"
)

type RequestType int
type MessageType int
type NodeStatus int

const (
	REQUEST_WRITE RequestType = iota
	REQUEST_READ
)
const (
	COORDINATOR_WRITE MessageType = iota
	COORDINATOR_READ
	READ_REPAIR
	WRITE_ACK
	READ_OK
	WRITE_OK
	KILL
	KILL_ACK
)

const (
	ALIVE NodeStatus = iota
	DEAD
)

func (r RequestType) String() string {
	return [...]string{"Write", "Read"}[r]
}

func (s NodeStatus) String() string {
	return [...]string{"Alive", "Dead"}[s]
}

type Handler struct {
	Request       *Request
	Node          *Node
	Ring          *Ring
	Timeout       time.Duration
	QuorumChannel chan PeerMessage
	Responses     int
}
type Node struct {
	//DataStore is for coordinator to store responses from other nodes before being sent back to the client
	DataStore map[int]PeerMessage
	Id        int        `json:"id"`
	IPAddress string     `json:"ip_address"`
	Port      string     `json:"port"`
	Hash      int64      `json:"hash"`
	Status    NodeStatus `json:"status"`
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

//Request means message from client
type Request struct {
	Type     RequestType `json:"type"`
	Content  string      `json:"content"`
	SourceID int         `json:"node_id"`
}

//PeerMessage means message from other SandDB nodes
type PeerMessage struct {
	Type     MessageType `json:"type"`
	Version  int         `json:"version"`
	Content  string      `json:"content"`
	SourceID int         `json:"node_id"`
}

//Data is the information written in / fetched from DB
type Data struct {
}
