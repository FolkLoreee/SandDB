package messages

type RequestType int

const (
	REQUEST_WRITE RequestType = iota
	REQUEST_READ
	REQUEST_CREATE
)

type CreateRequest struct {
	TableName          string   `json:"table_name"`
	PartitionKeyNames  []string `json:"partition_key_names"`
	ClusteringKeyNames []string `json:"clustering_key_names"`
}

type WriteRequest struct {
	TableName           string      `json:"table_name"`
	PartitionKeyValues  []string    `json:"partition_keys"`
	HashedPK            int64       `json:"pk_hash"`
	ClusteringKeyValues []string    `json:"clustering_keys"`
	CellNames           []string    `json:"cell_names"`
	CellValues          []string    `json:"cell_values"`
	Type                MessageType `json:"type"`
}

type ReadRequest struct {
	TableName           string      `json:"table_name"`
	PartitionKeyValues  []string    `json:"partition_keys"`
	HashedPK            int64       `json:"pk_hash"`
	ClusteringKeyValues []string    `json:"clustering_keys"`
	Type                MessageType `json:"type"`
}
