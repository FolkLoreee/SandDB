package strict_quorum

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
)

func (h *Handler) HandleRequest(c *fiber.Ctx) error {
	var (
		clientRequest Request
	)
	err := c.BodyParser(&clientRequest)
	if err != nil {
		fmt.Printf("Error in parsing request: %s", err.Error())
		return err
	}
	h.Request = &clientRequest
	if clientRequest.Type == REQUEST_WRITE {
		if err = h.handleClientWriteRequest(); err != nil {
			_ = c.SendString(err.Error())
			return err
		} else {
			_ = c.SendStatus(http.StatusOK)
		}
	} else if clientRequest.Type == REQUEST_READ {
		data, err := h.handleClientReadRequest()
		if err != nil {
			_ = c.SendString(err.Error())
			return err
		} else {
			body, err := json.Marshal(data.Content)
			if err != nil {
				fmt.Printf("Error marshalling read request: %s", err.Error())
				return err
			}
			c.Status(http.StatusOK)
			_ = c.Send(body)
		}
	}
	return nil
}
