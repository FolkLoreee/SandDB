package strict_quorum

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

func (r RequestType) String() string {
	return [...]string{"Write", "Read"}[r]
}

type Handler struct {
	Request     *Request
	Node        *Node
	Cluster     *Cluster
	VoteChannel chan string
	Votes       int
}
type Node struct {
	Id        int    `json:"id"`
	Clock     int    `json:"clock"`
	IPAddress string `json:"ip_address"`
}

//Cluster consists of multiple Nodes
type Cluster struct {
	Votes int     `json:"votes"`
	Nodes []*Node `json:"node_ids"`
}

//Request means message from client
type Request struct {
	Type    RequestType `json:"type"`
	Content string      `json:"content"`
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
