package strict_quorum

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func (h *RequestHandler) HandleRequest(c *fiber.Ctx) error {
	var (
		request Request
	)
	err := c.BodyParser(request)
	if err != nil {
		fmt.Printf("Error in parsing request: %s", err.Error())
		return err
	}
	h.Request = &request
	if request.Type == REQUEST_WRITE {
		if err = h.handleWriteRequest(); err != nil {
			return err
			//TODO: send back an error response
		}
		//TODO: If no error, send back an OK response
	}
	return nil
}
func (h *RequestHandler) handleWriteRequest() error {
	//TODO: Send quorum request
	//TODO: Wait for quorum
	//TODO: Write data to local DB
	return nil
}
func (h *RequestHandler) handleReadRequest() (error, Data) {
	//TODO: Find the "closest replica"
	//TODO: return data
	return nil, Data{}
}

func (h *RequestHandler) sendQuorumRequest(nodeID int) error {
	//TODO: Util tool to send quorum request
	return nil
}
