package read_write

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sanddb/db"
	"sanddb/messages"
	"sanddb/utils"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) HandleClientReadRequest(c *fiber.Ctx) error {
	var (
		req messages.ReadRequest
	)

	node := h.Node

	fmt.Println("=============================================")
	fmt.Println("Node", node.Id, "handling read request")
	fmt.Println("=============================================")

	if err := c.BodyParser(&req); err != nil {
		return err
	}
	partitionKeyConcat := ""
	// Look for the receiverNode
	for _, partitionKey := range req.PartitionKeyValues {
		partitionKeyConcat += partitionKey
	}
	req.HashedPK = utils.GetHash(partitionKeyConcat)

	receiverNode := h.Ring.GetNode(partitionKeyConcat)
	fmt.Printf("Routing request to receiverNode %d at position %d...\n", receiverNode.Id, receiverNode.Hash)
	err := h.createQuorum(messages.REQUEST_READ)
	if err != nil {
		fmt.Printf("Error in creating quorum: %s", err.Error())
		return err
	}
	responses := make([]db.ReadResponse, 0)
	response, err := h.sendReadRequest(receiverNode, req)
	if err != nil {
		fmt.Printf("Error in sending coordinator request: %s", err.Error())
		return err
	}
	responses = append(responses, response)
	fmt.Printf("Ring replication factor is %d.\n", h.Ring.ReplicationFactor)

	replicas := h.Ring.Replicate(partitionKeyConcat)
	fmt.Printf("Replica content: %v\n", replicas)

	for _, receivingNode := range replicas {
		fmt.Println("Sending request to node", receivingNode.Id)
		response, err = h.sendReadRequest(receivingNode, req)
		if err != nil {
			fmt.Printf("Error sending read request: %s", err.Error())
			return err
		}
		responses = append(responses, response)
	}
	err = h.closeQuorum()
	if err != nil {
		fmt.Printf("Closing quorum error: %s", err.Error())
		return err
	}
	latestVersion := &db.Row{}
	if len(responses) > 1 {
		// fmt.Println("Data collected from nodes:", node.DataStore)
		for _, resp := range responses {
			if latestVersion.ClusteringKeyHash == 0 || resp.Row.UpdatedAt.UnixNano() > latestVersion.UpdatedAt.UnixNano() {
				latestVersion.CreatedAt = resp.Row.CreatedAt
				latestVersion.UpdatedAt = resp.Row.UpdatedAt
				latestVersion.DeletedAt = resp.Row.DeletedAt
				latestVersion.ClusteringKeyHash = resp.Row.ClusteringKeyHash
				latestVersion.ClusteringKeyValues = resp.Row.ClusteringKeyValues
				latestVersion.Cells = resp.Row.Cells
			}
		}
		cellNames := make([]string, 0)
		cellValues := make([]string, 0)
		for _, cell := range latestVersion.Cells {
			cellNames = append(cellNames, cell.Name)
			cellValues = append(cellValues, cell.Value)
		}
		writeReq := messages.WriteRequest{
			TableName:           req.TableName,
			PartitionKeyValues:  req.PartitionKeyValues,
			ClusteringKeyValues: req.ClusteringKeyValues,
			CellNames:           cellNames,
			CellValues:          cellValues,
			Type:                messages.READ_REPAIR,
		}
		for _, resp := range responses {
			if resp.Row.UpdatedAt.UnixNano() < latestVersion.UpdatedAt.UnixNano() {
				fmt.Printf("Sending read repair to node %d\n", resp.SourceNode.Id)
				if err = h.sendWriteRequest(resp.SourceNode, writeReq); err != nil {
					return err
				}
			}
		}
	} else {
		fmt.Println("No data received from nodes")
		return fiber.NewError(http.StatusInternalServerError, "Read fail: Insufficient responses for Quorum")
	}

	body, err := json.Marshal(latestVersion)
	if err != nil {
		fmt.Printf("Error in marshalling response: %s", err.Error())
		return err
	}
	_ = c.Status(http.StatusOK).Send(body)
	return nil
}

func (h *Handler) sendReadRequest(receivingNode *utils.Node, req messages.ReadRequest) (db.ReadResponse, error) {
	readResponse := db.ReadResponse{}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error marshalling read request: %s", err.Error())
		return readResponse, err
	}
	postBody := bytes.NewBuffer(body)

	response, err := http.Post(receivingNode.IPAddress+receivingNode.Port+"/db/read", "application/json", postBody)
	if err != nil {
		fmt.Printf("Error posting read request: %s", err.Error())
		return readResponse, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		jsonResponse, err := ioutil.ReadAll(response.Body)
		err = json.Unmarshal([]byte(jsonResponse), &readResponse)
		if err != nil {
			return db.ReadResponse{}, err
		}
		reply := messages.PeerMessage{
			Type:     messages.READ_OK,
			Content:  "1",
			SourceID: readResponse.SourceNode.Id,
		}
		h.QuorumChannel <- reply
		return readResponse, nil
	} else {
		errMsg := fmt.Sprintf("Row not found.")
		err = fiber.NewError(http.StatusBadRequest, errMsg)
		return db.ReadResponse{}, err
	}
}
