package consistent_hashing

type RequestType int
type MessageType int

const (
	REQUEST_WRITE RequestType = iota
	REQUEST_READ
)

const (
	COORDINATOR_REQUEST RequestType = iota
	NODE_ACK
)

func (r RequestType) String() string {
	return [...]string{"Write", "Read"}[r]
}

type Handler struct {
	Request *Request
	Node    *Node
	Ring    *Ring
}

type Node struct {
	Id        int    `json:"id"`
	IPAddress string `json:"ip_address"`
	Port      string `json:"port"`
	Hash      int64  `json:"hash"`
	// Status    string `json:"status"`
}

// Ring consists of multiple Nodes
type Ring struct {
	Nodes      []*Node         `json:"nodes" yaml:"nodes"`
	MinVotes   int             `json:"min_votes" yaml:"min_votes"`
	NodeMap    map[int64]*Node `json:"nodeMap"`
	NodeHashes []int64         `json:"nodeHashes"`
}

// Request means message from client
type Request struct {
	Type    RequestType `json:"type"`
	Content string      `json:"content"`
}

// PeerMessage means message from other SandDB nodes
type PeerMessage struct {
	Type     MessageType `json:"type"`
	Content  string      `json:"content"`
	SourceID int         `json:"node_id"`
}

// Data is the information written in / fetched from DB
type Data struct {
}
