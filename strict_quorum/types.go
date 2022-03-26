package strict_quorum

import (
	"time"
)

type RequestType int
type MessageType int

const (
	REQUEST_WRITE RequestType = iota
	REQUEST_READ
)
const (
	QUORUM_REQUEST MessageType = iota
	QUORUM_OK
)
const (
	READ_OK MessageType = iota
	WRITE_OK
)

func (r RequestType) String() string {
	return [...]string{"Write", "Read"}[r]
}

type Handler struct {
	Request     *Request
	Node        *Node
	Cluster     *Cluster
	Timeout     time.Duration
	VoteChannel chan string
	Votes       int
}
type Node struct {
	DataChannel chan Reply
	DataStore   map[int]Reply
	Id          int    `json:"id"`
	IPAddress   string `json:"ip_address"`
	Port        string `json:"port"`
}

//Cluster consists of multiple Nodes
type Cluster struct {
	Nodes       []*Node `json:"nodes" yaml:"nodes"`
	MinVotes    int     `json:"min_votes" yaml:"min_votes"`
	CurrentNode *Node   `json:"current_node"`
}

//Request means message from client
type Request struct {
	Type     RequestType `json:"type"`
	Content  int         `json:"content"`
	SourceID int         `json:"node_id"`
}

//PeerMessage means message from other SandDB nodes
type PeerMessage struct {
	Type     MessageType `json:"type"`
	Content  string      `json:"content"`
	SourceID int         `json:"node_id"`
}

//Data is the information written in / fetched from DB
type Data struct {
}

type Reply struct {
	Type     MessageType `json:"type"`
	Version  int         `json:"version"`
	Content  int         `json:"content"`
	SourceID int         `json:"node_id"`
}
