package strict_quorum

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"net/http"
)

func (h *Handler) HandleQuorumRequest(c *fiber.Ctx) error {
	quorumResponse := &PeerMessage{
		Type:     QUORUM_OK,
		Content:  "1",
		SourceID: h.Node.Id,
	}
	resp, err := json.Marshal(quorumResponse)
	if err != nil {
		_ = c.SendStatus(http.StatusInternalServerError)
		return err
	}
	_ = c.Send(resp)
	return nil
}
