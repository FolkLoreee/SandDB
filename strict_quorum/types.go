package strict_quorum

import "github.com/gofiber/fiber/v2"

type RequestType int

const (
	REQUEST_WRITE RequestType = iota
	REQUEST_READ
)

func (r RequestType) String() string {
	return [...]string{"Write", "Read"}[r]
}

type RequestHandler struct {
	App     *fiber.App
	Request *Request
}
type Node struct {
	Id    int `json:"id"`
	Clock int `json:"clock"`
}

//Cluster consists of multiple
type Cluster struct {
	Votes   int   `json:"votes"`
	NodeIDs []int `json:"node_ids"`
}
type PeerHandler struct {
	Cluster Cluster
}

//Request means message from client
type Request struct {
	Type    RequestType `json:"type"`
	Content string      `json:"content"`
}

//PeerMessage means message from other SandDB nodes
type PeerMessage struct {
}

//Data is the information written in / fetched from DB
type Data struct {
}
