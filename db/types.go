package db

import (
	"sanddb/utils"
	"strconv"
	"time"
)

// Handler is for each individual node to handle local read/write to file
type Handler struct {
	Node *utils.Node
}
type EpochTime time.Time

type LocalData []*Table
type Table struct {
	TableName          string       `json:"table_name"`
	PartitionKeyNames  []string     `json:"partition_key_names"`
	ClusteringKeyNames []string     `json:"clustering_key_names"`
	Partitions         []*Partition `json:"partitions"`
}

type Partition struct {
	Metadata *PartitionMetadata `json:"partition_metadata"`
	Rows     []*Row             `json:"rows"`
}

/* PartitionMetadata
PartitionKey: hash value of the concatenated partition keys
PartitionKeyValues: values of the table's partition keys
*/
type PartitionMetadata struct {
	PartitionKey       int64    `json:"partition_key"`
	PartitionKeyValues []string `json:"partition_key_values"`
}

type Row struct {
	CreatedAt           EpochTime `json:"created_at"`
	UpdatedAt           EpochTime `json:"updated_at"`
	DeletedAt           EpochTime `json:"deleted_at"`
	ClusteringKeyHash   int64     `json:"clustering_key_hash"`
	ClusteringKeyValues []string  `json:"clustering_key_values"`
	Cells               []*Cell   `json:"cells"`
}

type Cell struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// MarshalJSON is used to convert the timestamp to JSON
func (t EpochTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(t).UnixNano(), 10)), nil
}

// UnmarshalJSON is used to convert the timestamp from JSON
func (t *EpochTime) UnmarshalJSON(s []byte) (err error) {
	r := string(s)
	q, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = time.Unix(0, q)
	return nil
}

// Unix returns t as a Unix time, the number of seconds elapsed
// since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
func (t EpochTime) Unix() int64 {
	return time.Time(t).Unix()
}

// This returns the Unix time in nanoseconds.
func (t EpochTime) UnixNano() int64 {
	return time.Time(t).UnixNano()
}

// Time returns the JSON time as a time.Time instance in UTC
func (t EpochTime) Time() time.Time {
	return time.Time(t).UTC()
}

// String returns t as a formatted string
func (t EpochTime) String() string {
	return t.Time().String()
}

type ReadResponse struct {
	SourceNode *utils.Node
	Row        Row
}
