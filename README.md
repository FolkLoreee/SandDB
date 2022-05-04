# SandDB

_"I'm gonna put some sand in Cassandra's eye."_

<p align="center">
  <img src="./apache_cassandra_logo.png" width="400px" alt="Apache Cassandra">
</p>

Custom Implementation of [Apache Cassandra](https://cassandra.apache.org/) DB for SUTD ISTD 2022 50.041: Distributed Systems and Computing Group Project.

## Project Structure

```
sanddb/
├─ anti_entropy/
├─ client/
├─ config/
├─ data/
├─ db/
├─ messages/
├─ read_write/
├─ ring-visualiser/
├─ utils/
├─ main.go
```

## API Endpoints 🔚

> Note: both partition keys and clustering keys are required since they form the primary key.

### Create Table

**HTTP Method**

```
POST
```

**URL**

```
http://localhost:<port>/create/
```

**Request Body:**

```json
{
  "table_name": "hospitals",
  "partition_key_names": ["HOSPITAL_ID", "DEPARTMENT"],
  "clustering_key_names": ["ROOM_ID"]
}
```

- table_name: name of the table to be inserted/updated
- partition_key_names: headers of the partition keys
- clustering_key_names: headers of clustering keys

### Insert/Update

**HTTP Method**

```
POST
```

**URL**

```
http://localhost:<port>/insert/
```

**Request Body:**

```json
{
  "table_name": "hospitals",
  "partition_keys": ["1", "GENERAL"],
  "clustering_keys": ["AA-1"],
  "cell_names": ["Bed", "Oxygen Tank"],
  "cell_values": ["3", "10"]
}
```

- table_name: name of the table to be inserted/updated
- partition_keys: values of the partition keys
- clustering_keys: values of the clustering keys
- cell_names: column headers to be added into the row
- cell_values: values of the columns to be added into the row

### Read

**HTTP Method**

```
POST
```

**URL**

```
http://localhost:<port>/read/
```

**Request Body**

```json
{
  "table_name": "hospitals",
  "partition_keys": ["1", "GENERAL"],
  "clustering_keys": ["AA-1"]
}
```

Params:

- table_name: name of the table to be queried from
- partition_keys: values of the partition keys
- clustering_keys: values of the clustering keys (optional)

### Delete

**HTTP Method**

```
POST
```

**URL**

```
http://localhost:<port>/delete/
```

**Request Body**

```json
{
  "table_name": "hospitals",
  "partition_keys": ["1", "GENERAL"],
  "clustering_keys": ["AA-1"]
}
```

Params:

- table_name: name of the table to remove an entry from
- partition_keys: values of the partition keys of the row to be deleted
- clustering_keys: values of the clustering keys of the row to be deleted

## Database Structs

`table_name.json`:

```go
type Table []Partition

type Partition struct {
	Metadata PartitionMetadata `json:"partition_metadata"`
	Rows     []Row             `json:"rows"`
}

type PartitionMetadata struct {
	TableName          string   `json:"table_name"`
	PartitionKey       int64    `json:"partition_key"`
	PartitionKeyNames  []string `json:"partition_key_names"`
	PartitionKeyValues []string `json:"partition_key_values"`
	ClusteringKeyNames []string `json:"clustering_key_names"`
}

type Row struct {
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	DeletedAt           time.Time `json:"deleted_at"`
	ClusteringKeyHash   int64     `json:"clustering_key_hash"`
	ClusteringKeyValues []string  `json:"clustering_key_values"`
	Cells               []Cell    `json:"cells"`
}

type Cell struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
```

## Anti-Entropy

For future work in implementing the entire full Merkle Tree, as well as its comparisons, these repositories might be useful:

- [https://github.com/cbergoon/merkletree](https://github.com/cbergoon/merkletree)
- [https://github.com/aprimadi/merkle-tree](https://github.com/aprimadi/merkle-tree)

## Acknowledgements

Credits and thanks to:

- Group 5 Team Members:
  - [Filbert Cia](https://github.com/FolkLoreee)
  - [James Raphael Tiovalen](https://github.com/jamestiotio)
  - [Ong Zhi Yi](https://github.com/gzyon)
  - [Yu Nicole Frances Cabansay](https://github.com/nicolefranc)
- 50.041 Course Instructor: [Professor Sudipta Chattopadhyay](https://istd.sutd.edu.sg/people/faculty/sudipta-chattopadhyay)
