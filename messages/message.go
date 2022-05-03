package messages

type MessageType int

const (
	COORDINATOR_WRITE MessageType = iota
	COORDINATOR_READ
	COORDINATOR_CREATE
	READ_REPAIR
	WRITE_ACK
	CREATE_ACK
	READ_OK
	WRITE_OK
)

//PeerMessage means message from other SandDB nodes
type PeerMessage struct {
	Type     MessageType `json:"type"`
	Content  string      `json:"content"`
	SourceID int         `json:"node_id"`
	Version  int         `json:"version"`
}
