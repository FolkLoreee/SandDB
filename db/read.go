package db

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"sanddb/messages"
	"sanddb/utils"
)

func (h *Handler) HandleDBRead(c *fiber.Ctx) error {
	var (
		reqBody messages.ReadRequest
		readRow *Row
	)
	err := c.BodyParser(&reqBody)
	if err != nil {
		return err
	}
	filename := fmt.Sprintf("data/%d.json", h.Node.Id)
	localData, err := ReadJSON(filename)

	table := GetTable(reqBody.TableName, localData)
	if table == nil {
		errMsg := fmt.Sprintf("Table %s does not exist.", reqBody.TableName)
		err = fiber.NewError(http.StatusBadRequest, errMsg)
		errBody, _ := json.Marshal(err)
		_ = c.Status(http.StatusBadRequest).Send(errBody)
		return err
	}
	readPartition := GetPartition(table, reqBody.HashedPK)
	clusteringKeyHash := utils.GetHashFromKeys(reqBody.ClusteringKeyValues)

	for _, row := range readPartition.Rows {
		if row.ClusteringKeyHash == clusteringKeyHash {
			readRow = row
		}
	}

	if readRow == nil {
		errMsg := fmt.Sprintf("Row not found.")
		err = fiber.NewError(http.StatusBadRequest, errMsg)
		errBody, _ := json.Marshal(err)
		_ = c.Status(http.StatusBadRequest).Send(errBody)
		return err
	}
	node := h.Node
	reply := ReadResponse{
		SourceNode: node,
		Row:        *readRow,
	}

	body, err := json.Marshal(reply)
	_ = c.Status(http.StatusOK).Send(body)
	fmt.Printf("Sent: %v\n", reply)
	return err
}
