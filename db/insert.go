package db

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"sanddb/messages"
	"sanddb/utils"
	"time"
)

func (h *Handler) HandleDBInsert(c *fiber.Ctx) error {
	var (
		reqBody messages.WriteRequest
	)
	filename := fmt.Sprintf("%d/table.json", h.Node.Id)
	localData, err := ReadJSON(filename)
	if err := c.BodyParser(&reqBody); err != nil {
		return err
	}
	table := GetTable(reqBody.TableName, localData)
	if table == nil {
		errMsg := fmt.Sprintf("Table %s does not exist.", reqBody.TableName)
		err = fiber.NewError(http.StatusBadRequest, errMsg)
		errBody, _ := json.Marshal(err)
		_ = c.Status(http.StatusBadRequest).Send(errBody)
		return err
	}
	clusteringKeyConcat := ""
	for _, clusteringKey := range reqBody.ClusteringKeyValues {
		clusteringKeyConcat += clusteringKey
	}
	clusteringKeyHash := utils.GetHash(clusteringKeyConcat)
	updatedPartition := GetPartition(table, reqBody.HashedPK)
	if updatedPartition == nil {
		if err = createNewPartition(reqBody, table, clusteringKeyHash); err != nil {
			return err
		}
	} else {
		if err = updatePartition(updatedPartition, reqBody, clusteringKeyHash); err != nil {
			return err
		}
	}
	if err = PersistTable(localData, filename, table); err != nil {
		return err
	}
	reply := &messages.PeerMessage{
		Type:     messages.WRITE_ACK,
		Content:  "1",
		SourceID: h.Node.Id,
	}
	resp, err := json.Marshal(reply)
	if err != nil {
		_ = c.SendStatus(http.StatusInternalServerError)
		return err
	}
	_ = c.Send(resp)
	return nil
}

func createNewPartition(req messages.WriteRequest, table *Table, clusteringKeyHash int64) error {
	metadata := &PartitionMetadata{
		PartitionKey:       req.HashedPK,
		PartitionKeyValues: req.PartitionKeyValues,
	}
	rows := make([]*Row, 0)
	cells := make([]*Cell, 0)
	for i := range req.CellNames {
		cell := &Cell{
			Name:  req.CellNames[i],
			Value: req.CellValues[i],
		}
		cells = append(cells, cell)
	}
	row := &Row{
		CreatedAt:           EpochTime(time.Now()),
		ClusteringKeyHash:   clusteringKeyHash,
		ClusteringKeyValues: req.ClusteringKeyValues,
		Cells:               cells,
	}
	rows = append(rows, row)
	partition := &Partition{
		Metadata: metadata,
		Rows:     rows,
	}
	table.Partitions = append(table.Partitions, partition)
	return nil
}

func updatePartition(partition *Partition, req messages.WriteRequest, clusteringKeyHash int64) error {
	cells := make([]*Cell, 0)
	for i := range req.CellNames {
		cell := &Cell{
			Name:  req.CellNames[i],
			Value: req.CellValues[i],
		}
		cells = append(cells, cell)
	}
	for _, row := range partition.Rows {
		if row.ClusteringKeyHash == clusteringKeyHash {
			row.Cells = cells
			row.UpdatedAt = EpochTime(time.Now())
			return nil
		}
	}
	newRow := &Row{
		CreatedAt:           EpochTime(time.Now()),
		ClusteringKeyHash:   clusteringKeyHash,
		ClusteringKeyValues: req.ClusteringKeyValues,
		Cells:               cells,
	}
	partition.Rows = append(partition.Rows, newRow)
	return nil
}
