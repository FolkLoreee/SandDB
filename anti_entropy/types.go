package anti_entropy

import (
	"sanddb/db"
	"sanddb/utils"
	"time"
)

type RequestType int
type MessageType int

type RepairStatus int

const (
	FAILED RepairStatus = iota
	SUCCESSFUL
	NOTHING_CHANGED
)

type AntiEntropyHandler struct {
	Request                *Request
	Node                   *utils.Node
	Ring                   *utils.Ring
	RepairTimeout          time.Duration
	InternalRequestTimeout time.Duration
	GCGraceSeconds         int
}

type Request struct {
	Type     RequestType `json:"type"`
	Content  string      `json:"content"`
	SourceID int         `json:"node_id"`
}

type RepairGetRequest struct {
	TableName         string `json:"table_name"`
	PartitionKey      int64  `json:"partition_key"`
	ClusteringKeyHash int64  `json:"clustering_key_hash"`
	NodeID            int    `json:"node_id"`
}

type RepairGetResponse struct {
	Data   *db.Row `json:"data"`
	Hash   int64   `json:"hash"`
	NodeID int     `json:"node_id"`
}

type RepairWriteRequest struct {
	TableName          string          `json:"table_name"`
	PartitionKeyNames  []string        `json:"partition_key_names"`
	ClusteringKeyNames []string        `json:"clustering_key_names"`
	Partitions         []*db.Partition `json:"partitions"`
	NodeID             int             `json:"node_id"`
}

type SubrepairRequest struct {
	ExistingData []RepairGetRequest `json:"existing_data"`
	NodeID       int                `json:"node_id"`
}

type SubrepairResponse struct {
	DataToAdd []RepairGetRequest `json:"data_to_add"`
	NodeID    int                `json:"node_id"`
}

type RepairDeleteRequest struct {
	NodeID int `json:"node_id"`
}
