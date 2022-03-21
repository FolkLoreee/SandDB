package consistent_hashing

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) HandleRequest(c *fiber.Ctx) error {
	var request Request
	err := c.BodyParser(&request)
	if err != nil {
		fmt.Printf("Error in parsing request: %s", err.Error())
		return err
	}
	h.Request = &request
	if request.Type == REQUEST_WRITE {
		if err = h.handleWriteRequest(); err != nil {
			_ = c.SendString(err.Error())
			return err
		}
		_ = c.SendStatus(http.StatusOK)
	}
	return nil
}

func (h *Handler) handleWriteRequest() error {

	// Hash partition key sent by client
	partitionKey := h.Request.Content
	hashedPK := GetHash(partitionKey)
	fmt.Printf("%s hashed to %d\n", partitionKey, hashedPK)

	fmt.Println(h.Ring.NodeHashes)
	// index := h.Ring.Search(hashedPK)
	// fmt.Println(index)
	node := h.Ring.GetNode(partitionKey)
	fmt.Printf("%d with hash %d", node.Id, node.Hash)

	return nil

}
