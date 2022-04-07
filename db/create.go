package db

import (
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func (h *Handler) HandleCreateTable(c *fiber.Ctx) error {
	var (
		reqBody CreateRequest
	)
	//TODO: check if table name already exists
	err := c.BodyParser(&reqBody)
	if err != nil {
		return err
	}
	//TODO: create the table struct
	partitions := make([]Partition, 0)
	table := Table{
		TableName:          reqBody.TableName,
		PartitionKeyNames:  reqBody.PartitionKeyNames,
		ClusteringKeyNames: reqBody.ClusteringKeyNames,
		Partitions:         partitions,
	}

	//TODO: persist this table metadata on json
	filename := strconv.Itoa(h.Node.Id) + ".json"
	err = PersistTable(filename, table)
	if err != nil {
		return err
	}
	return nil
}
