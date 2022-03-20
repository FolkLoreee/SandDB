package strict_quorum

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
)

func (h *Handler) HandleQuorumRequest(c *fiber.Ctx) error {
	var (
		requestMsg PeerMessage
	)
	quorumResponse := &PeerMessage{
		Type:     QUORUM_OK,
		Content:  "1",
		SourceID: h.Node.Id,
	}
	err := c.BodyParser(&requestMsg)
	if err != nil {
		return err
	}
	resp, err := json.Marshal(quorumResponse)
	if err != nil {
		_ = c.SendStatus(http.StatusInternalServerError)
		return err
	}
	_ = c.Send(resp)
	fmt.Printf("Quorum OK sent to Node: %d\n", requestMsg.SourceID)
	return nil
}
