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

//func (h *Handler) handleClientReadRequest() (error, Data) {
//	//TODO: Find the "closest replica"
//	//TODO: return data
//	return nil, Data{}
//}

//func (h *Handler) sendQuorumRequest(node *Node) error {
//	var (
//		responseMsg PeerMessage
//	)
//	quorumRequest := PeerMessage{
//		Type:     QUORUM_REQUEST,
//		Content:  "",
//		SourceID: h.Node.Id,
//	}
//	body, err := json.Marshal(quorumRequest)
//	if err != nil {
//		fmt.Printf("Error in marshalling quorum request: %s", err.Error())
//		return err
//	}
//	postBody := bytes.NewBuffer(body)
//	response, err := http.Post(node.IPAddress+node.Port+"/quorum/start", "application/json", postBody)
//
//	if err != nil {
//		fmt.Printf("Error in posting quorum request: %s", err.Error())
//		return err
//	}
//	defer response.Body.Close()
//	jsonResponse, err := ioutil.ReadAll(response.Body)
//
//	err = json.Unmarshal([]byte(jsonResponse), &responseMsg)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//func (h *Handler) countQuorum() {
//	for {
//		select {
//		case msg := <-h.QuorumChannel:
//			if msg == "1" {
//				h.Responses++
//				fmt.Printf("ACK received. Current ACKs: %d\n", h.Responses)
//			}
//		case <-time.After(h.Timeout):
//			fmt.Printf("Timeout\n")
//			return
//		}
//	}
//}
