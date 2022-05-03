package read_write

import (
	"sanddb/messages"
	"sanddb/utils"
	"time"
)

type Handler struct {
	//Request       *Request
	Node          *utils.Node
	Ring          *utils.Ring
	Timeout       time.Duration
	QuorumChannel chan messages.PeerMessage
	Responses     int
}

//Request means message from client
//type Request struct {
//	Type     RequestType `json:"type"`
//	Content  string      `json:"content"`
//	SourceID int         `json:"node_id"`
//}

//TODO: Remove version and content from PeerMessage
