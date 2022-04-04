package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	// "reflect"

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

type Hospitals struct {
	Hospitals []Hospital `json:"hospitals"`
}

type Hospital struct {
	Partition Partition `json:"partition"`
	Rows      []Row     `json:"rows"`
}

type Partition struct {
	Key      []string `json:"key"`
	Position int      `json:"position"`
}

type Row struct {
	Type          string              `json:"type"`
	Position      int                 `json:"position"`
	Clustering    []string            `json:"clustering"`
	LivelinesInfo Liveliness          `json:"liveliness_info"`
	DeletionInfo  map[string]int      `json:"deletion_info"`
	Cells         []map[string]string `json:"cells"`
}

type Liveliness struct {
	Tstamp  int  `json:"tstamp"`
	TTL     int  `json:"ttl"`
	Expires int  `json:"expires_at"`
	Expired bool `json:"expired"`
}

func GetAllHospitals(c *fiber.Ctx) error {
	var reply interface{}

	fmt.Println("Received request")

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

	// Open our jsonFile
	hospitals, err := readJSON("data.json")
	if err != nil {
		fmt.Printf("Error reading JSON file: %s", err.Error())
	}

	hospitalId, _ := strconv.Atoi(c.Params("hospitalID"))
	fmt.Printf("Getting information for hospital %d...", hospitalId)
	hospital := hospitals.Hospitals[hospitalId]

	for _, row := range hospital.Rows {
		data := row.Cells
		for _, cell := range data {
			hospitalData = append(hospitalData, cell)
		}
	}

	return c.JSON(fiber.Map{"status": "success", "message": hospitalData})
}

func GetResource(c *fiber.Ctx) error {
	var resourceData map[string]string
	return c.JSON(fiber.Map{"status": "success", "message": resourceData})
}

// helper function
func readJSON(filename string) (Hospitals, error) {
	jsonFile, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		return Hospitals{}, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var hospitals Hospitals
	json.Unmarshal([]byte(byteValue), &hospitals)

	return hospitals, nil
}

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
