package anti_entropy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sanddb/db"
	"sanddb/utils"
	"sort"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spaolacci/murmur3"
)

// This anti-entropy module should be delegated to a daemon that runs in the background at each node.
// This function should then be called at regular intervals.
// Frequency of the anti-entropy repair operation is a configurable knob that can be tuned.
// We leave tuning out of the scope of this project, since that would take quite a significant amount of time.
// For testing and demonstration purposes, we can manually trigger the execution of this anti-entropy repair procedure.
// This manual trigger is similar to executing the "nodetool repair" command in the real Apache Cassandra.
// Note that this process is usually computationally expensive/intensive, and thus should be run sparingly during "peaceful" times only.
// Strong assumptions are being made, some of which are:
// 1. Client requests are deferred until repair is complete (or that client's requests are not frequent enough). A background thread could potentially handle this repair, but additional care needs to be taken when resolving conflicts between client requests and repair requests (such as by comparing timestamps).
// 2. No network partitions occur during the repair process. This also assumes that all messages eventually arrive at their designated destinations.
// 3. No non-Byzantine or Byzantine failures, such as node crashes or wrong computations, occur during the repair process.
// Usually, this module is also triggered during SSTable compaction process, but since we do not implement actual SSTables for this project, we do not need to worry about that.
// By right, this process would also delete all tombstones created more than GC_GRACE_SECONDS ago.
// This implementation repairs the entire data that resides (and is supposed to reside) in the current node.
// Future improvements can include a subrange repair feature as well to minimize the impact of the repair.
func (h *AntiEntropyHandler) HandleFullRepairRequest(c *fiber.Ctx) error {
	var netClient = &http.Client{
		Timeout: h.InternalRequestTimeout,
	}
	// LAPLUS represents the status of the anti-entropy repair process (defaults to NOTHING_CHANGED)
	LAPLUS := NOTHING_CHANGED
	nodeID := h.Node.Id
	file, err := ioutil.ReadFile("data/" + strconv.Itoa(nodeID) + ".json")
	if err != nil {
		log.Println("Error reading file:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}
	var data db.LocalData
	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		log.Println("Error unmarshalling data:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}

	// Get ring from AntiEntropyHandler
	ring := h.Ring

	// Initialize existingData to be empty
	// This is used for subrepair requests
	// Rationale: If data exists and belongs to the primary node, an attempt to get the data will be made to the replicas
	// Only one is needed since it will be reset after each phase/view of the repair process and since only at most one repair command will be executed in the entire ring at a time
	// Keeping track of this here (instead of at the replicas) ensures stateless operations between different phases of the anti-entropy repair process, ensuring simplicity
	var existingData []RepairGetRequest

	for _, table := range data {
		for _, partition := range table.Partitions {
			index := ring.Search(partition.Metadata.PartitionKey)
			// Perform anti-entropy repair for the Primary Range.
			// We should not initiate repair of data "owned" by other nodes in this node since that is not this node's responsibility!
			// This is to avoid redundant repairs for the sake of performance and to keep the entire protocol simple.
			if index == nodeID {
				for _, row := range partition.Rows {
					// Perform the normal operation here
					dataFromReplicas := make([]RepairGetResponse, h.Ring.ReplicationFactor)
					// We basically get a hash representation of the binary form of the row
					// This is super hacky, but nobody cares!
					cells, err := json.Marshal(row)
					if err != nil {
						log.Println("Error marshalling row:", err)
						return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
					}
					dataHash := murmur3.New64()
					dataHash.Write([]byte(cells))
					hash := utils.ByteArrayToInt(dataHash.Sum(nil))
					dataFromReplicas[0] = RepairGetResponse{
						Data:   row,
						Hash:   hash,
						NodeID: index,
					}
					requestData := RepairGetRequest{
						TableName:         table.TableName,
						PartitionKey:      partition.Metadata.PartitionKey,
						ClusteringKeyHash: row.ClusteringKeyHash,
						NodeID:            nodeID,
					}
					// Append to existingData
					existingData = append(existingData, requestData)
					for i := 1; i < h.Ring.ReplicationFactor; i++ {
						secondaryIndex := (index + i) % len(h.Ring.NodeHashes)
						// Prepare POST body
						requestBody, err := json.Marshal(requestData)
						if err != nil {
							log.Println("Error marshalling request data:", err)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
						}
						postBody := bytes.NewBuffer(requestBody)
						// Perform POST request
						response, err := netClient.Post(h.Ring.Nodes[secondaryIndex].IPAddress+h.Ring.Nodes[secondaryIndex].Port+"/internal/repair/get_data", "application/json", postBody)
						if err != nil {
							log.Println("Error performing POST request:", err)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
						}
						defer response.Body.Close()
						// Read response
						body, err := ioutil.ReadAll(response.Body)
						if err != nil {
							log.Println("Error reading response:", err)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
						}
						// Unmarshal response
						var responseData RepairGetResponse
						err = json.Unmarshal(body, &responseData)
						if err != nil {
							log.Println("Error unmarshalling response:", err)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
						}

						secondaryData, err := json.Marshal(responseData.Data)
						if err != nil {
							log.Println("Error marshalling row:", err)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
						}
						secondaryDataHash := murmur3.New64()
						secondaryDataHash.Write([]byte(secondaryData))
						secondaryHash := utils.ByteArrayToInt(secondaryDataHash.Sum(nil))

						if responseData.Hash != secondaryHash && responseData.Hash != -1 {
							log.Println("Hash mismatch due to potential Byzantine error! Hashes:", responseData.Hash, secondaryHash)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: Hash mismatch due to potential Byzantine error!")
						}

						dataFromReplicas[i] = responseData
					}

					if len(dataFromReplicas) != h.Ring.ReplicationFactor {
						log.Println("Not enough replicas to perform anti-entropy repair!")
						return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: Not enough replicas to perform anti-entropy repair!")
					}

					// Select the most updated data
					latestDataIndex := 0
					latestUpdatedAt := dataFromReplicas[0].Data.UpdatedAt.UnixNano()
					for j := 1; j < len(dataFromReplicas); j++ {
						replicaData := dataFromReplicas[j]
						// TODO: Hmm seems like -1 is actually a valid legal hash value for an int64 data type, might need another placeholder for this (idea: use null/empty string check for the "data" field?)
						if replicaData.Hash != -1 {
							if replicaData.Data.UpdatedAt.UnixNano() > latestUpdatedAt {
								latestUpdatedAt = replicaData.Data.UpdatedAt.UnixNano()
								latestDataIndex = j
							} else if replicaData.Data.UpdatedAt.UnixNano() == latestUpdatedAt {
								latestReplicaData, err := json.Marshal(dataFromReplicas[latestDataIndex].Data)
								if err != nil {
									log.Println("Error marshalling row:", err)
									return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
								}
								currentReplicaData, err := json.Marshal(replicaData.Data)
								if err != nil {
									log.Println("Error marshalling row:", err)
									return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
								}
								// You might complain about this tie-breaking mechanism to get the total order, but this is what Apache Cassandra actually does :)
								compareResult := bytes.Compare(currentReplicaData, latestReplicaData)
								// If the data itself is equal, it doesn't matter which version we choose
								// Additionally, if the data is the same, we won't enter this update path in the first place
								// Hence, we just need to update the latestDataIndex if the data itself is different
								if compareResult > 0 {
									latestDataIndex = j
								}
							}
						}
					}

					// Get all the indices from dataFromReplicas that need updating
					nodesToUpdate := make([]int, 0)
					for i, data := range dataFromReplicas {
						// TODO: Ditto about the -1 hash value
						if data.Hash == -1 {
							// This means that the replica does not have this data and thus we need to write
							nodesToUpdate = append(nodesToUpdate, i)
						} else if data.Data.UpdatedAt.UnixNano() < latestUpdatedAt {
							nodesToUpdate = append(nodesToUpdate, i)
						} else if data.Data.UpdatedAt.UnixNano() == latestUpdatedAt && i != latestDataIndex {
							latestReplicaData, err := json.Marshal(dataFromReplicas[latestDataIndex].Data)
							if err != nil {
								log.Println("Error marshalling row:", err)
								return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
							}
							currentReplicaData, err := json.Marshal(data.Data)
							if err != nil {
								log.Println("Error marshalling row:", err)
								return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
							}
							compareResult := bytes.Compare(currentReplicaData, latestReplicaData)
							if compareResult < 0 {
								nodesToUpdate = append(nodesToUpdate, i)
							}
						}
					}

					updateRequest := RepairWriteRequest{
						TableName:          table.TableName,
						PartitionKeyNames:  table.PartitionKeyNames,
						ClusteringKeyNames: table.ClusteringKeyNames,
						Partitions: []*db.Partition{
							{
								Metadata: partition.Metadata,
								Rows: []*db.Row{
									dataFromReplicas[latestDataIndex].Data,
								},
							},
						},
						NodeID: nodeID,
					}

					// Perform POST request to ALL replicas specified in nodesToUpdate (which might include this primary node itself)
					// This method ensures that there will be no further stale values/version conflicts between the data stored in the RAM and the data stored on disk for the current primary node, guaranteeing consistency
					for _, id := range nodesToUpdate {
						replicaID := (nodeID + id) % len(h.Ring.NodeHashes)

						requestBody, err := json.Marshal(updateRequest)
						if err != nil {
							log.Println("Error marshalling request data:", err)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
						}

						// Prepare POST body
						updateBody := bytes.NewBuffer(requestBody)

						updateResponse, err := netClient.Post(h.Ring.Nodes[replicaID].IPAddress+h.Ring.Nodes[replicaID].Port+"/internal/repair/write_data", "application/json", updateBody)
						if err != nil {
							log.Println("Error performing POST request:", err)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
						}
						defer updateResponse.Body.Close()
						if updateResponse.StatusCode != fiber.StatusOK {
							log.Println("Error performing POST request:", updateResponse.StatusCode)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + updateResponse.Status)
						}
						if LAPLUS == NOTHING_CHANGED {
							LAPLUS = SUCCESSFUL
						}
					}
				}
			}
		}
	}

	// Send missing subrepair requests (with existingData) to the replicas
	for i := 1; i < h.Ring.ReplicationFactor; i++ {
		subrepairRequest := SubrepairRequest{
			ExistingData: existingData,
			NodeID:       nodeID,
		}
		subrepairRequestBody, err := json.Marshal(subrepairRequest)
		if err != nil {
			log.Println("Error marshalling request data:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
		}
		subrepairBody := bytes.NewBuffer(subrepairRequestBody)
		replicaID := (nodeID + i) % len(h.Ring.NodeHashes)
		subrepairResponse, err := netClient.Post(h.Ring.Nodes[replicaID].IPAddress+h.Ring.Nodes[replicaID].Port+"/internal/repair/missing_subrepair", "application/json", subrepairBody)
		if err != nil {
			log.Println("Error performing POST request:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
		}
		defer subrepairResponse.Body.Close()
		if subrepairResponse.StatusCode != fiber.StatusOK {
			log.Println("Error performing POST request:", subrepairResponse.StatusCode)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + subrepairResponse.Status)
		}
		// Append subrepair response to existingData
		subrepairResponseBody, err := ioutil.ReadAll(subrepairResponse.Body)
		if err != nil {
			log.Println("Error reading response:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
		}
		var subrepairResponseData SubrepairResponse
		err = json.Unmarshal(subrepairResponseBody, &subrepairResponseData)
		if err != nil {
			log.Println("Error unmarshalling response:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
		}
		existingData = append(existingData, subrepairResponseData.DataToAdd...)

		if LAPLUS == NOTHING_CHANGED {
			LAPLUS = SUCCESSFUL
		}
	}

	// Send all delete triggers
	// Delete all tombstones that are older than GC_GRACE_SECONDS, only AFTER performing synchronization between replicas
	deleteRequest := RepairDeleteRequest{
		NodeID: nodeID,
	}
	for i := 0; i < h.Ring.ReplicationFactor; i++ {
		deleteRequestBody, err := json.Marshal(deleteRequest)
		if err != nil {
			log.Println("Error marshalling request data:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
		}
		deleteBody := bytes.NewBuffer(deleteRequestBody)
		replicaID := (nodeID + i) % len(h.Ring.NodeHashes)
		deleteResponse, err := netClient.Post(h.Ring.Nodes[replicaID].IPAddress+h.Ring.Nodes[replicaID].Port+"/internal/repair/trigger_delete", "application/json", deleteBody)
		if err != nil {
			log.Println("Error performing POST request:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
		}
		defer deleteResponse.Body.Close()
		if deleteResponse.StatusCode != fiber.StatusOK {
			log.Println("Error performing POST request:", deleteResponse.StatusCode)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + deleteResponse.Status)
		}
		if LAPLUS == NOTHING_CHANGED {
			LAPLUS = SUCCESSFUL
		}
	}

	// Request a primary range repair (not full repair) to all of the following nodes
	for i := 1; i < len(h.Ring.NodeHashes); i++ {
		replicaID := (nodeID + i) % len(h.Ring.NodeHashes)
		primaryRangeRepairResponse, err := netClient.Post(h.Ring.Nodes[replicaID].IPAddress+h.Ring.Nodes[replicaID].Port+"/repair", "application/json", bytes.NewBuffer([]byte{}))
		if err != nil {
			log.Println("Error performing POST request:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
		}
		defer primaryRangeRepairResponse.Body.Close()
		if primaryRangeRepairResponse.StatusCode != fiber.StatusOK {
			log.Println("Error performing POST request:", primaryRangeRepairResponse.StatusCode)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + primaryRangeRepairResponse.Status)
		}
		if LAPLUS == NOTHING_CHANGED {
			LAPLUS = SUCCESSFUL
		}
	}

	if LAPLUS == SUCCESSFUL {
		return c.Status(fiber.StatusOK).SendString("Successfully performed the anti-entropy repair.")
	} else if LAPLUS == FAILED {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair.")
	} else {
		return c.Status(fiber.StatusOK).SendString("Nothing was changed during repair since all replicas are currently consistent. Nice!")
	}
}

// Essentially, this is mostly the same as a full repair, but without passing the chain forward to all of the nodes.
// This function is used by a full repair call.
func (h *AntiEntropyHandler) HandleRepairRequest(c *fiber.Ctx) error {
	var netClient = &http.Client{
		Timeout: h.InternalRequestTimeout,
	}
	LAPLUS := NOTHING_CHANGED
	nodeID := h.Node.Id
	file, err := ioutil.ReadFile("data/" + strconv.Itoa(nodeID) + ".json")
	if err != nil {
		log.Println("Error reading file:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}
	var data db.LocalData
	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		log.Println("Error unmarshalling data:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}

	ring := h.Ring

	var existingData []RepairGetRequest

	for _, table := range data {
		for _, partition := range table.Partitions {
			index := ring.Search(partition.Metadata.PartitionKey)
			if index == nodeID {
				for _, row := range partition.Rows {
					dataFromReplicas := make([]RepairGetResponse, h.Ring.ReplicationFactor)
					cells, err := json.Marshal(row)
					if err != nil {
						log.Println("Error marshalling row:", err)
						return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
					}
					dataHash := murmur3.New64()
					dataHash.Write([]byte(cells))
					hash := utils.ByteArrayToInt(dataHash.Sum(nil))
					dataFromReplicas[0] = RepairGetResponse{
						Data:   row,
						Hash:   hash,
						NodeID: index,
					}
					requestData := RepairGetRequest{
						TableName:         table.TableName,
						PartitionKey:      partition.Metadata.PartitionKey,
						ClusteringKeyHash: row.ClusteringKeyHash,
						NodeID:            nodeID,
					}
					existingData = append(existingData, requestData)
					for i := 1; i < h.Ring.ReplicationFactor; i++ {
						secondaryIndex := (index + i) % len(h.Ring.NodeHashes)
						requestBody, err := json.Marshal(requestData)
						if err != nil {
							log.Println("Error marshalling request data:", err)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
						}
						postBody := bytes.NewBuffer(requestBody)
						response, err := netClient.Post(h.Ring.Nodes[secondaryIndex].IPAddress+h.Ring.Nodes[secondaryIndex].Port+"/internal/repair/get_data", "application/json", postBody)
						if err != nil {
							log.Println("Error performing POST request:", err)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
						}
						defer response.Body.Close()
						body, err := ioutil.ReadAll(response.Body)
						if err != nil {
							log.Println("Error reading response:", err)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
						}
						var responseData RepairGetResponse
						err = json.Unmarshal(body, &responseData)
						if err != nil {
							log.Println("Error unmarshalling response:", err)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
						}

						secondaryData, err := json.Marshal(responseData.Data)
						if err != nil {
							log.Println("Error marshalling row:", err)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
						}
						secondaryDataHash := murmur3.New64()
						secondaryDataHash.Write([]byte(secondaryData))
						secondaryHash := utils.ByteArrayToInt(secondaryDataHash.Sum(nil))

						if responseData.Hash != secondaryHash && responseData.Hash != -1 {
							log.Println("Hash mismatch due to potential Byzantine error! Hashes:", responseData.Hash, secondaryHash)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: Hash mismatch due to potential Byzantine error!")
						}

						dataFromReplicas[i] = responseData
					}

					if len(dataFromReplicas) != h.Ring.ReplicationFactor {
						log.Println("Not enough replicas to perform anti-entropy repair!")
						return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: Not enough replicas to perform anti-entropy repair!")
					}

					latestDataIndex := 0
					latestUpdatedAt := dataFromReplicas[0].Data.UpdatedAt.UnixNano()
					for j := 1; j < len(dataFromReplicas); j++ {
						replicaData := dataFromReplicas[j]
						// TODO: Hmm seems like -1 is actually a valid legal hash value for an int64 data type, might need another placeholder for this (idea: use null/empty string check for the "data" field?)
						if replicaData.Hash != -1 {
							if replicaData.Data.UpdatedAt.UnixNano() > latestUpdatedAt {
								latestUpdatedAt = replicaData.Data.UpdatedAt.UnixNano()
								latestDataIndex = j
							} else if replicaData.Data.UpdatedAt.UnixNano() == latestUpdatedAt {
								latestReplicaData, err := json.Marshal(dataFromReplicas[latestDataIndex].Data)
								if err != nil {
									log.Println("Error marshalling row:", err)
									return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
								}
								currentReplicaData, err := json.Marshal(replicaData.Data)
								if err != nil {
									log.Println("Error marshalling row:", err)
									return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
								}
								compareResult := bytes.Compare(currentReplicaData, latestReplicaData)
								if compareResult > 0 {
									latestDataIndex = j
								}
							}
						}
					}

					nodesToUpdate := make([]int, 0)
					for i, data := range dataFromReplicas {
						// TODO: Ditto about the -1 hash value
						if data.Hash == -1 {
							nodesToUpdate = append(nodesToUpdate, i)
						} else if data.Data.UpdatedAt.UnixNano() < latestUpdatedAt {
							nodesToUpdate = append(nodesToUpdate, i)
						} else if data.Data.UpdatedAt.UnixNano() == latestUpdatedAt && i != latestDataIndex {
							latestReplicaData, err := json.Marshal(dataFromReplicas[latestDataIndex].Data)
							if err != nil {
								log.Println("Error marshalling row:", err)
								return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
							}
							currentReplicaData, err := json.Marshal(data.Data)
							if err != nil {
								log.Println("Error marshalling row:", err)
								return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
							}
							compareResult := bytes.Compare(currentReplicaData, latestReplicaData)
							if compareResult < 0 {
								nodesToUpdate = append(nodesToUpdate, i)
							}
						}
					}

					updateRequest := RepairWriteRequest{
						TableName:          table.TableName,
						PartitionKeyNames:  table.PartitionKeyNames,
						ClusteringKeyNames: table.ClusteringKeyNames,
						Partitions: []*db.Partition{
							{
								Metadata: partition.Metadata,
								Rows: []*db.Row{
									dataFromReplicas[latestDataIndex].Data,
								},
							},
						},
						NodeID: nodeID,
					}

					for _, id := range nodesToUpdate {
						replicaID := (nodeID + id) % len(h.Ring.NodeHashes)

						requestBody, err := json.Marshal(updateRequest)
						if err != nil {
							log.Println("Error marshalling request data:", err)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
						}

						updateBody := bytes.NewBuffer(requestBody)

						updateResponse, err := netClient.Post(h.Ring.Nodes[replicaID].IPAddress+h.Ring.Nodes[replicaID].Port+"/internal/repair/write_data", "application/json", updateBody)
						if err != nil {
							log.Println("Error performing POST request:", err)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
						}
						defer updateResponse.Body.Close()
						if updateResponse.StatusCode != fiber.StatusOK {
							log.Println("Error performing POST request:", updateResponse.StatusCode)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + updateResponse.Status)
						}
						if LAPLUS == NOTHING_CHANGED {
							LAPLUS = SUCCESSFUL
						}
					}
				}
			}
		}
	}

	for i := 1; i < h.Ring.ReplicationFactor; i++ {
		subrepairRequest := SubrepairRequest{
			ExistingData: existingData,
			NodeID:       nodeID,
		}
		subrepairRequestBody, err := json.Marshal(subrepairRequest)
		if err != nil {
			log.Println("Error marshalling request data:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
		}
		subrepairBody := bytes.NewBuffer(subrepairRequestBody)
		replicaID := (nodeID + i) % len(h.Ring.NodeHashes)
		subrepairResponse, err := netClient.Post(h.Ring.Nodes[replicaID].IPAddress+h.Ring.Nodes[replicaID].Port+"/internal/repair/missing_subrepair", "application/json", subrepairBody)
		if err != nil {
			log.Println("Error performing POST request:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
		}
		defer subrepairResponse.Body.Close()
		if subrepairResponse.StatusCode != fiber.StatusOK {
			log.Println("Error performing POST request:", subrepairResponse.StatusCode)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + subrepairResponse.Status)
		}
		subrepairResponseBody, err := ioutil.ReadAll(subrepairResponse.Body)
		if err != nil {
			log.Println("Error reading response:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
		}
		var subrepairResponseData SubrepairResponse
		err = json.Unmarshal(subrepairResponseBody, &subrepairResponseData)
		if err != nil {
			log.Println("Error unmarshalling response:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
		}
		existingData = append(existingData, subrepairResponseData.DataToAdd...)

		if LAPLUS == NOTHING_CHANGED {
			LAPLUS = SUCCESSFUL
		}
	}

	deleteRequest := RepairDeleteRequest{
		NodeID: nodeID,
	}
	for i := 0; i < h.Ring.ReplicationFactor; i++ {
		deleteRequestBody, err := json.Marshal(deleteRequest)
		if err != nil {
			log.Println("Error marshalling request data:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
		}
		deleteBody := bytes.NewBuffer(deleteRequestBody)
		replicaID := (nodeID + i) % len(h.Ring.NodeHashes)
		deleteResponse, err := netClient.Post(h.Ring.Nodes[replicaID].IPAddress+h.Ring.Nodes[replicaID].Port+"/internal/repair/trigger_delete", "application/json", deleteBody)
		if err != nil {
			log.Println("Error performing POST request:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
		}
		defer deleteResponse.Body.Close()
		if deleteResponse.StatusCode != fiber.StatusOK {
			log.Println("Error performing POST request:", deleteResponse.StatusCode)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + deleteResponse.Status)
		}
		if LAPLUS == NOTHING_CHANGED {
			LAPLUS = SUCCESSFUL
		}
	}

	if LAPLUS == SUCCESSFUL {
		return c.Status(fiber.StatusOK).SendString("Successfully performed the anti-entropy repair.")
	} else if LAPLUS == FAILED {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair.")
	} else {
		return c.Status(fiber.StatusOK).SendString("Nothing was changed during repair since all replicas are currently consistent. Nice!")
	}
}

// Ask the other nodes to read and return data for repair.
func (h *AntiEntropyHandler) HandleRepairGetRequest(c *fiber.Ctx) error {
	// Parse JSON input
	var requestData RepairGetRequest
	if err := c.BodyParser(&requestData); err != nil {
		log.Println("Error parsing request body:", err)
		return c.Status(fiber.StatusBadRequest).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}
	nodeID := h.Node.Id
	file, err := ioutil.ReadFile("data/" + strconv.Itoa(nodeID) + ".json")
	if err != nil {
		log.Println("Error reading file:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}
	var data db.LocalData
	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		log.Println("Error unmarshalling data:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}

	// Default values if no corresponding row data is found
	responseData := RepairGetResponse{
		Data:   &db.Row{},
		Hash:   -1,
		NodeID: nodeID,
	}

	// Get the data with the details from requestData
	for _, table := range data {
		if table.TableName == requestData.TableName {
			for _, partition := range table.Partitions {
				if partition.Metadata.PartitionKey == requestData.PartitionKey {
					for _, row := range partition.Rows {
						if row.ClusteringKeyHash == requestData.ClusteringKeyHash {
							cells, err := json.Marshal(row)
							if err != nil {
								log.Println("Error marshalling row:", err)
								return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
							}
							dataHash := murmur3.New64()
							dataHash.Write([]byte(cells))
							hash := dataHash.Sum(nil)
							// Prepare response
							responseData.Data = row
							responseData.Hash = utils.ByteArrayToInt(hash[:])
							responseData.NodeID = nodeID
						}
					}
				}
			}
		}
	}

	resp, err := json.Marshal(responseData)

	if err != nil {
		log.Println("Error marshalling row:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}
	log.Println("Replying to get_data request for repair by", requestData.NodeID)
	return c.Status(fiber.StatusOK).Send(resp)
}

// Ask the other nodes to write data for repair.
// This write will be I/O heavy, since each call will grab the data from the file on disk, modify the data on RAM, and then write/flush it back to the file on disk.
// That said, since the purpose of this function is to ensure consistency via repair, which we will eventually achieve via this function, we do not really care about the performance.
// Future iterations might want to improve and optimize this implementation.
func (h *AntiEntropyHandler) HandleRepairWriteRequest(c *fiber.Ctx) error {
	// Parse JSON input
	var requestData RepairWriteRequest
	if err := c.BodyParser(&requestData); err != nil {
		log.Println("Error parsing request body:", err)
		return c.Status(fiber.StatusBadRequest).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}
	nodeID := h.Node.Id
	file, err := ioutil.ReadFile("data/" + strconv.Itoa(nodeID) + ".json")
	if err != nil {
		log.Println("Error reading file:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}
	var data db.LocalData
	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		log.Println("Error unmarshalling data:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}

	dataIsUpdated := false
	dataIsFound := false

	// Edge case where there is no data in local file yet
	if len(data) == 0 {
		newTable := &db.Table{
			TableName:          requestData.TableName,
			PartitionKeyNames:  requestData.PartitionKeyNames,
			ClusteringKeyNames: requestData.ClusteringKeyNames,
			Partitions:         requestData.Partitions,
		}
		data = append(data, newTable)
		if !dataIsUpdated {
			dataIsUpdated = true
		}
	} else {
	out:
		for i, table := range data {
			if table.TableName == requestData.TableName {
				for j, partition := range table.Partitions {
					if partition.Metadata.PartitionKey == requestData.Partitions[0].Metadata.PartitionKey {
						for k, row := range partition.Rows {
							if row.ClusteringKeyHash == requestData.Partitions[0].Rows[0].ClusteringKeyHash {
								dataIsFound = true
								incomingData := requestData.Partitions[0].Rows[0]
								// Additional check to only execute writing if the incoming data is actually newer than the existing data (which should always be the case if a write_data request is performed in the first place, but just in case)
								if row.UpdatedAt.UnixNano() < incomingData.UpdatedAt.UnixNano() {
									if !dataIsUpdated {
										dataIsUpdated = true
									}
									// Update the row data
									data[i].Partitions[j].Rows[k] = incomingData
									break out
								} else if row.UpdatedAt.UnixNano() == incomingData.UpdatedAt.UnixNano() {
									incomingDataByteArray, err := json.Marshal(incomingData)
									if err != nil {
										log.Println("Error marshalling row:", err)
										return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
									}
									currentDataByteArray, err := json.Marshal(row)
									if err != nil {
										log.Println("Error marshalling row:", err)
										return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
									}
									compareResult := bytes.Compare(currentDataByteArray, incomingDataByteArray)
									if compareResult < 0 {
										if !dataIsUpdated {
											dataIsUpdated = true
										}
										data[i].Partitions[j].Rows[k] = incomingData
										break out
									}
								}
							}
						}
						// Add missing data
						if !dataIsUpdated && !dataIsFound {
							dataIsUpdated = true
							data[i].Partitions[j].Rows = append(data[i].Partitions[j].Rows, requestData.Partitions[0].Rows[0])
							break out
						}
					}
				}
				// Add missing data
				if !dataIsUpdated && !dataIsFound {
					dataIsUpdated = true
					data[i].Partitions = append(data[i].Partitions, requestData.Partitions[0])
					break out
				}
			}
		}
		// Add missing data
		if !dataIsUpdated && !dataIsFound {
			dataIsUpdated = true
			newTable := &db.Table{
				TableName:          requestData.TableName,
				PartitionKeyNames:  requestData.PartitionKeyNames,
				ClusteringKeyNames: requestData.ClusteringKeyNames,
				Partitions:         requestData.Partitions,
			}
			data = append(data, newTable)
		}
	}

	// Sort the rows in each partition
	for _, table := range data {
		for _, partition := range table.Partitions {
			sort.SliceStable(partition.Rows, func(i, j int) bool {
				return partition.Rows[i].ClusteringKeyHash < partition.Rows[j].ClusteringKeyHash
			})
		}
	}

	// Sort the partitions in each table
	for _, table := range data {
		sort.SliceStable(table.Partitions, func(i, j int) bool {
			if table.Partitions[i].Metadata.PartitionKey == table.Partitions[j].Metadata.PartitionKey {
				return table.Partitions[i].Rows[0].ClusteringKeyHash < table.Partitions[j].Rows[0].ClusteringKeyHash
			}
			return table.Partitions[i].Metadata.PartitionKey < table.Partitions[j].Metadata.PartitionKey
		})
	}

	// Sort the tables based on table name, partition key, and clustering key hash before writing to file
	sort.SliceStable(data, func(i, j int) bool {
		if data[i].TableName == data[j].TableName {
			if data[i].Partitions[0].Metadata.PartitionKey == data[j].Partitions[0].Metadata.PartitionKey {
				return data[i].Partitions[0].Rows[0].ClusteringKeyHash < data[j].Partitions[0].Rows[0].ClusteringKeyHash
			}
			return data[i].Partitions[0].Metadata.PartitionKey < data[j].Partitions[0].Metadata.PartitionKey
		}
		return data[i].TableName < data[j].TableName
	})

	// Write the data back to disk
	if dataIsUpdated {
		file, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			log.Println("Error marshalling data:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
		}
		err = ioutil.WriteFile("data/"+strconv.Itoa(nodeID)+".json", file, 0644)
		if err != nil {
			log.Println("Error writing file:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
		}
	}

	log.Println("Replying to write_data request for repair by", requestData.NodeID)

	if dataIsUpdated {
		return c.Status(fiber.StatusOK).SendString("Successfully performed the requested write for repair.")
	} else {
		return c.Status(fiber.StatusOK).SendString("No data was updated for repair.")
	}
}

// Ask the other nodes to delete data for repair.
// Since time is always ever moving forward, there might be edge cases whereby some replicas have not deleted the tombstones (<= threshold), but some other replicas actually have deleted their local tombstones (> threshold).
// However, we don't really care since EVENTUALLY, every replica will delete its local tombstones in the future. Any tombstones that are still around might still be written back to the primary node or its replicas, but that's not a problem since it will not be read by/returned to the client. Thus, data consistency is not compromised.
func (h *AntiEntropyHandler) HandleRepairDeleteRequest(c *fiber.Ctx) error {
	// GC_GRACE_SECONDS is 10 days by default
	GC_GRACE_SECONDS := time.Duration(h.GCGraceSeconds*24) * time.Hour
	// Parse JSON input
	var requestData RepairDeleteRequest
	if err := c.BodyParser(&requestData); err != nil {
		log.Println("Error parsing request body:", err)
		return c.Status(fiber.StatusBadRequest).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}
	nodeID := h.Node.Id
	file, err := ioutil.ReadFile("data/" + strconv.Itoa(nodeID) + ".json")
	if err != nil {
		log.Println("Error reading file:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}
	var data db.LocalData
	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		log.Println("Error unmarshalling data:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}

	ring := h.Ring

	dataDeleted := false

	for i, table := range data {
		for j, partition := range table.Partitions {
			index := ring.Search(partition.Metadata.PartitionKey)
			// We perform only primary range deletion on behalf of the requestor node
			if index == requestData.NodeID {
				for _, row := range partition.Rows {
					// Technically, negative epoch time is actually valid (before January 1, 1970), but we use it in this middleware application as invalid (other placeholders could be considered in the future)
					if row.DeletedAt.UnixNano() >= 0 && time.Since(time.Unix(0, row.DeletedAt.UnixNano())) > GC_GRACE_SECONDS {
						data[i].Partitions[j].Rows = append(data[i].Partitions[j].Rows[:j], data[i].Partitions[j].Rows[j+1:]...)
						if !dataDeleted {
							dataDeleted = true
						}
					}
				}
			}
		}
	}

	// Write the data back to disk
	if dataDeleted {
		file, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			log.Println("Error marshalling data:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
		}
		err = ioutil.WriteFile("data/"+strconv.Itoa(nodeID)+".json", file, 0644)
		if err != nil {
			log.Println("Error writing file:", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
		}
	}

	log.Println("Replying to delete_data request for repair by", requestData.NodeID)

	if dataDeleted {
		log.Println("Deleted all tombstones that are older than GC_GRACE_SECONDS ago.")
		return c.Status(fiber.StatusOK).SendString("Successfully performed the requested delete for repair.")
	} else {
		log.Println("No tombstones to delete.")
		return c.Status(fiber.StatusOK).SendString("No data was deleted for repair.")
	}
}

// [EDGE CASE] Ask the other nodes to perform repair for data not present in this node.
// We use ExistingData to determine which data is missing from the primary node.
func (h *AntiEntropyHandler) HandleMissingSubrepairRequest(c *fiber.Ctx) error {
	var netClient = &http.Client{
		Timeout: h.InternalRequestTimeout,
	}
	// Parse JSON input
	var requestData SubrepairRequest
	if err := c.BodyParser(&requestData); err != nil {
		log.Println("Error parsing request body:", err)
		return c.Status(fiber.StatusBadRequest).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}
	nodeID := h.Node.Id
	file, err := ioutil.ReadFile("data/" + strconv.Itoa(nodeID) + ".json")
	if err != nil {
		log.Println("Error reading file:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}
	var data db.LocalData
	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		log.Println("Error unmarshalling data:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}

	ring := h.Ring

	subrepairResponse := SubrepairResponse{
		DataToAdd: []RepairGetRequest{},
		NodeID:    nodeID,
	}

	for _, table := range data {
		for _, partition := range table.Partitions {
			index := ring.Search(partition.Metadata.PartitionKey)
			// Do primary range repair for the requestor node
			if index == requestData.NodeID {
				for _, row := range partition.Rows {
					rowData := RepairGetRequest{
						TableName:         table.TableName,
						PartitionKey:      partition.Metadata.PartitionKey,
						ClusteringKeyHash: row.ClusteringKeyHash,
						NodeID:            nodeID,
					}
					// Execute repair only if data is not found in the attached existingData
					if !ExistingDataContains(requestData.ExistingData, rowData) {
						subrepairResponse.DataToAdd = append(subrepairResponse.DataToAdd, rowData)
						dataFromReplicas := make([]RepairGetResponse, h.Ring.ReplicationFactor)
						cells, err := json.Marshal(row)
						if err != nil {
							log.Println("Error marshalling row:", err)
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
						}
						dataHash := murmur3.New64()
						dataHash.Write([]byte(cells))
						hash := utils.ByteArrayToInt(dataHash.Sum(nil))
						dataFromReplicas[0] = RepairGetResponse{
							Data:   &db.Row{},
							Hash:   -1,
							NodeID: index,
						}

						for i := 1; i < h.Ring.ReplicationFactor; i++ {
							secondaryIndex := (index + i) % len(h.Ring.NodeHashes)
							if secondaryIndex == nodeID {
								dataFromReplicas[i] = RepairGetResponse{
									Data:   row,
									Hash:   hash,
									NodeID: nodeID,
								}
							} else {
								requestBody, err := json.Marshal(requestData)
								if err != nil {
									log.Println("Error marshalling request data:", err)
									return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
								}
								postBody := bytes.NewBuffer(requestBody)
								response, err := netClient.Post(h.Ring.Nodes[secondaryIndex].IPAddress+h.Ring.Nodes[secondaryIndex].Port+"/internal/repair/get_data", "application/json", postBody)
								if err != nil {
									log.Println("Error performing POST request:", err)
									return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
								}
								defer response.Body.Close()
								body, err := ioutil.ReadAll(response.Body)
								if err != nil {
									log.Println("Error reading response:", err)
									return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
								}
								var responseData RepairGetResponse
								err = json.Unmarshal(body, &responseData)
								if err != nil {
									log.Println("Error unmarshalling response:", err)
									return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
								}

								secondaryData, err := json.Marshal(responseData.Data)
								if err != nil {
									log.Println("Error marshalling row:", err)
									return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
								}
								secondaryDataHash := murmur3.New64()
								secondaryDataHash.Write([]byte(secondaryData))
								secondaryHash := utils.ByteArrayToInt(secondaryDataHash.Sum(nil))

								if responseData.Hash != secondaryHash && responseData.Hash != -1 {
									log.Println("Hash mismatch due to potential Byzantine error! Hashes:", responseData.Hash, secondaryHash)
									return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: Hash mismatch due to potential Byzantine error!")
								}

								dataFromReplicas[i] = responseData
							}
						}

						if len(dataFromReplicas) != h.Ring.ReplicationFactor {
							log.Println("Not enough replicas to perform anti-entropy repair!")
							return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: Not enough replicas to perform anti-entropy repair!")
						}

						latestDataIndex := ((nodeID-index)%len(h.Ring.NodeHashes) + len(h.Ring.NodeHashes)) % len(h.Ring.NodeHashes)
						latestUpdatedAt := dataFromReplicas[latestDataIndex].Data.UpdatedAt.UnixNano()
						for j := 1; j < len(dataFromReplicas); j++ {
							replicaData := dataFromReplicas[j]
							// TODO: Hmm seems like -1 is actually a valid legal hash value for an int64 data type, might need another placeholder for this (idea: use null/empty string check for the "data" field?)
							if replicaData.Hash != -1 {
								if replicaData.Data.UpdatedAt.UnixNano() > latestUpdatedAt {
									latestUpdatedAt = replicaData.Data.UpdatedAt.UnixNano()
									latestDataIndex = j
								} else if replicaData.Data.UpdatedAt.UnixNano() == latestUpdatedAt && j != latestDataIndex {
									latestReplicaData, err := json.Marshal(dataFromReplicas[latestDataIndex].Data)
									if err != nil {
										log.Println("Error marshalling row:", err)
										return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
									}
									currentReplicaData, err := json.Marshal(replicaData.Data)
									if err != nil {
										log.Println("Error marshalling row:", err)
										return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
									}
									compareResult := bytes.Compare(currentReplicaData, latestReplicaData)
									if compareResult > 0 {
										latestDataIndex = j
									}
								}
							}
						}

						nodesToUpdate := make([]int, 0)
						for i, data := range dataFromReplicas {
							// TODO: Ditto about the -1 hash value
							if data.Hash == -1 {
								nodesToUpdate = append(nodesToUpdate, i)
							} else if data.Data.UpdatedAt.UnixNano() < latestUpdatedAt {
								nodesToUpdate = append(nodesToUpdate, i)
							} else if data.Data.UpdatedAt.UnixNano() == latestUpdatedAt && i != latestDataIndex {
								latestReplicaData, err := json.Marshal(dataFromReplicas[latestDataIndex].Data)
								if err != nil {
									log.Println("Error marshalling row:", err)
									return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
								}
								currentReplicaData, err := json.Marshal(data.Data)
								if err != nil {
									log.Println("Error marshalling row:", err)
									return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
								}
								compareResult := bytes.Compare(currentReplicaData, latestReplicaData)
								if compareResult < 0 {
									nodesToUpdate = append(nodesToUpdate, i)
								}
							}
						}

						updateRequest := RepairWriteRequest{
							TableName:          table.TableName,
							PartitionKeyNames:  table.PartitionKeyNames,
							ClusteringKeyNames: table.ClusteringKeyNames,
							Partitions: []*db.Partition{
								{
									Metadata: partition.Metadata,
									Rows: []*db.Row{
										dataFromReplicas[latestDataIndex].Data,
									},
								},
							},
							NodeID: nodeID,
						}

						for _, id := range nodesToUpdate {
							replicaID := (index + id) % len(h.Ring.NodeHashes)

							requestBody, err := json.Marshal(updateRequest)
							if err != nil {
								log.Println("Error marshalling request data:", err)
								return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
							}

							updateBody := bytes.NewBuffer(requestBody)

							updateResponse, err := netClient.Post(h.Ring.Nodes[replicaID].IPAddress+h.Ring.Nodes[replicaID].Port+"/internal/repair/write_data", "application/json", updateBody)
							if err != nil {
								log.Println("Error performing POST request:", err)
								return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
							}
							defer updateResponse.Body.Close()
							if updateResponse.StatusCode != fiber.StatusOK {
								log.Println("Error performing POST request:", updateResponse.StatusCode)
								return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + updateResponse.Status)
							}
						}
					}
				}
			}
		}
	}

	log.Println("Replying to missing_subrepair request for repair by", requestData.NodeID)

	resp, err := json.Marshal(subrepairResponse)

	if err != nil {
		log.Println("Error marshalling row:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to perform the anti-entropy repair. Error: " + err.Error())
	}

	return c.Status(fiber.StatusOK).Send(resp)
}

// Naive O(N) algorithm to determine if row data is present in the existing data
// Optimizations might involve implementing a hashmap or binary search
func ExistingDataContains(existingData []RepairGetRequest, row RepairGetRequest) bool {
	for _, r := range existingData {
		// Take advantage of short circuit evaluation to avoid unnecessary nestings
		if r.TableName == row.TableName && r.PartitionKey == row.PartitionKey && r.ClusteringKeyHash == row.ClusteringKeyHash {
			return true
		}
	}
	return false
}
