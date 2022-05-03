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
	Data   db.Row `json:"data"`
	Hash   int64  `json:"hash"`
	NodeID int    `json:"node_id"`
}

type RepairWriteRequest struct {
	TableName          string         `json:"table_name"`
	PartitionKeyNames  []string       `json:"partition_key_names"`
	ClusteringKeyNames []string       `json:"clustering_key_names"`
	Partitions         []db.Partition `json:"partitions"`
	NodeID             int            `json:"node_id"`
}

type SubrepairRequest struct {
	ExistingData []RepairGetRequest `json:"existing_data"`
	NodeID       int                `json:"node_id"`
}

type RepairDeleteRequest struct {
	NodeID int `json:"node_id"`
}

//type Ring struct {
//	Nodes             []*Node         `json:"nodes" yaml:"nodes"`
//	MinVotes          int             `json:"min_votes" yaml:"min_votes"`
//	CurrentNode       *Node           `json:"current_node"`
//	NodeMap           map[int64]*Node `json:"nodeMap"`
//	NodeHashes        []int64         `json:"nodeHashes"`
//	ReplicationFactor int             `json:"replication_factor"` // replicates at N-1 nodes
//}
//
//type Node struct {
//	//DataStore is for coordinator to store responses from other nodes before being sent back to the client
//	DataStore map[int]PeerMessage
//	Id        int    `json:"id"`
//	IPAddress string `json:"ip_address"`
//	Port      string `json:"port"`
//	Hash      int64  `json:"hash"`
//}
//
//type PeerMessage struct {
//	Type     MessageType `json:"type"`
//	Version  int         `json:"version"`
//	Content  string      `json:"content"`
//	SourceID int         `json:"node_id"`
//}

//type LocalData []Table
//
//type Table struct {
//	TableName          string      `json:"table_name"`
//	PartitionKeyNames  []string    `json:"partition_key_names"`
//	ClusteringKeyNames []string    `json:"clustering_key_names"`
//	Partitions         []Partition `json:"partitions"`
//}
//
//type Partition struct {
//	Metadata PartitionMetadata `json:"partition_metadata"`
//	Rows     []Row             `json:"rows"`
//}
//
//type PartitionMetadata struct {
//	PartitionKey       int64    `json:"partition_key"`
//	PartitionKeyValues []string `json:"partition_key_values"`
//}
//
//type Row struct {
//	CreatedAt           EpochTime `json:"created_at"`
//	UpdatedAt           EpochTime `json:"updated_at"`
//	DeletedAt           EpochTime `json:"deleted_at"`
//	ClusteringKeyHash   int64     `json:"clustering_key_hash"`
//	ClusteringKeyValues []string  `json:"clustering_key_values"`
//	Cells               []Cell    `json:"cells"`
//}
//
//type Cell struct {
//	Name  string `json:"name"`
//	Value string `json:"value"`
//}
//
//// EpochTime defines a timestamp encoded as epoch nanoseconds in JSON
//type EpochTime time.Time
