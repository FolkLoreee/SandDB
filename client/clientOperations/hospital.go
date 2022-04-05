package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

type RequestType int

const (
	REQUEST_WRITE RequestType = iota
	REQUEST_READ
)

func (r RequestType) String() string {
	return [...]string{"Write", "Read"}[r]
}

type Request struct {
	Type     RequestType `json:"type"`
	Content  string      `json:"content"`
	SourceID int         `json:"node_id"`
}

func GetAllHospitals(c *fiber.Ctx) error {
	var reply interface{}

	fmt.Println("Requesting for all hospital resources")

	jsonResponse, err := sendReadRequest(c)
	err = json.Unmarshal([]byte(jsonResponse), &reply)
	if err != nil {
		fmt.Printf("Error unmarshalling JSON at client: %s \n", err.Error())
		return c.Status(404).JSON(fiber.Map{"status": "error"})
	}

	return c.JSON(fiber.Map{"status": "success", "message": reply})
}

func GetHospital(c *fiber.Ctx) error {
	var hospitalData []map[string]string
	return c.JSON(fiber.Map{"status": "success", "message": hospitalData})
}

func GetResource(c *fiber.Ctx) error {
	var resourceData map[string]string
	return c.JSON(fiber.Map{"status": "success", "message": resourceData})
}

// helper function
func sendReadRequest(c *fiber.Ctx) ([]byte, error) {
	request := Request{
		Type:    REQUEST_READ,
		Content: "200",
	}

	nodePorts := []string{":8000", ":8001", ":8002", ":8003"}
	body, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("Error marshalling read request: %s", err.Error())
		return []byte{}, err
	}

	postBody := bytes.NewBuffer(body)
	rand.Seed(time.Now().UnixNano())
	response, err := http.Post("http://127.0.0.1"+nodePorts[rand.Intn(3)]+"/request", "application/json", postBody)
	if err != nil {
		fmt.Printf("Error sending client GET request: %s \n", err.Error())
		return []byte{}, c.Status(404).JSON(fiber.Map{"status": "error"})
	}
	defer response.Body.Close()

	jsonResponse, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error reading JSON response at client: %s \n", err.Error())
		return []byte{}, c.Status(404).JSON(fiber.Map{"status": "error"})
	}

	return jsonResponse, nil
}
