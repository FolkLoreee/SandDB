package leaderless_replication

import (
	"fmt"
	"time"
)

type Client struct {
	clientChannel chan Data
}

type Coordinator struct {
	dataChannel    chan Data
	requestChannel chan string
	nodeNetwork    []*Node
	messageStore   map[string][]Data
}

type Node struct {
	nodeID         int
	requestChannel chan string
	dataChannel    chan Data
	data           map[string]Data
	coordinator    *Coordinator
}

type Data struct {
	topic   string
	version int
	message int
}

/*
Client functions:
- send requests
- receive most updated write/read from coordinator

Coordinator functions:
- send client request to nodes in network and most updated write if any nodes are out of sync
- receive node replies and sends back the most updated reply back to the client

Node functions:
- depending on whether it receives a get or update request, either:
	- send its own data to the coordinator
	- update its data with the updated version
*/

// helper functions
func contains(dataStore []Data, d Data) bool {
	for _, data := range dataStore {
		if d == data {
			return true
		}
	}
	return false
}

// client send and receive

func (coordinator *Coordinator) clientSend(topic string) {
	coordinator.requestChannel <- topic
}

func (client *Client) clientGet(coordinator *Coordinator, getTopic string) {
	for {
		fmt.Println("=============================================")
		fmt.Println("New Client Request")
		fmt.Println("=============================================")

		// send a request to the coordinator
		fmt.Println("Client requesting for", getTopic)
		go coordinator.clientSend(getTopic)

		select {
		case receivedMessage := <-client.clientChannel:
			fmt.Println("Client received message", receivedMessage.message, "(version", receivedMessage.version, ")")
		case <-time.After(time.Second * 10):
			fmt.Println("No reply from coordinator")
		}
	}
}

// node functions
func (node *Node) nodeSend(message Data) {
	node.coordinator.dataChannel <- message
}

func (node *Node) handleCoordinatorRequest() {
	for {

		// receive request from coordinator
		select {

		// if coordinator sent get request
		case requestedTopic := <-node.requestChannel:
			reply := node.data[requestedTopic]
			go node.nodeSend(reply)

		// if coordinator sent update request
		case updateRequest := <-node.dataChannel:
			node.data[updateRequest.topic] = updateRequest
			fmt.Println(updateRequest.topic, "updated at Node", node.nodeID)
		}
	}
}

// coordinator functions
func (coordinator *Coordinator) broadcast(message Data) {
	for _, node := range coordinator.nodeNetwork {
		fmt.Println("Sending update to Node", node.nodeID)
		node.dataChannel <- message
	}
}

func (coordinator *Coordinator) request(topic string) {
	for _, node := range coordinator.nodeNetwork {
		fmt.Println("Requesting", topic, "from Node", node.nodeID)
		node.requestChannel <- topic
	}
}

func (coordinator *Coordinator) handleClientRequest(client *Client) {
	for {
		// receive requested topic from client
		requestMessage := <-coordinator.requestChannel
		go coordinator.request(requestMessage)

		// sleep while waiting for replies from nodes
		go coordinator.handleNodeReply()
		time.Sleep(time.Second * 5)

		// check for replies from nodes
		dataObtained := coordinator.messageStore[requestMessage]

		// if got data
		if len(dataObtained) > 0 {
			var latestVersion Data

			// get latest version
			for _, data := range dataObtained {
				if data.version > latestVersion.version {
					latestVersion = data
				}
			}

			// send latest version to client
			fmt.Println("Sending Client", latestVersion.message, "( version", latestVersion.version, ")")
			client.clientChannel <- latestVersion

			// clear logs
			dataObtained = []Data{}

		} else {
			fmt.Println("No data obtained for", requestMessage)
		}
	}
}

func (coordinator *Coordinator) handleNodeReply() {
	for {
		// receive reply from node
		nodeReply := <-coordinator.dataChannel
		topicStore := coordinator.messageStore[nodeReply.topic]

		// append reply to topicStore if coordinator does not already contain the reply
		if !contains(topicStore, nodeReply) {
			topicStore = append(topicStore, nodeReply)
		}

		coordinator.messageStore[nodeReply.topic] = topicStore
	}
}

func main() {
	fmt.Println("hello world")
	nodesList := []*Node{}

	// instantiate coordinator
	coordinator := Coordinator{
		dataChannel:    make(chan Data),
		requestChannel: make(chan string),
		nodeNetwork:    nodesList,
		messageStore:   map[string][]Data{},
	}

	// mockup data
	data := map[string]Data{
		"beds": {
			topic:   "beds",
			version: 2,
			message: 50,
		},
		"art kits": {
			topic:   "art kits",
			version: 3,
			message: 100,
		},
		"oxygen supply": {
			topic:   "oxygen supply",
			version: 1,
			message: 20,
		},
	}

	data2 := map[string]Data{
		"beds": {
			topic:   "beds",
			version: 1,
			message: 20,
		},
		"art kits": {
			topic:   "art kits",
			version: 4,
			message: 70,
		},
		"oxygen supply": {
			topic:   "oxygen supply",
			version: 1,
			message: 20,
		},
	}

	// dataList := []map[string]Data{data, data2}

	// instantiate nodes
	for i := 0; i < 5; i++ {
		node := Node{
			nodeID:         i,
			requestChannel: make(chan string),
			dataChannel:    make(chan Data),
			data:           map[string]Data{},
			coordinator:    &coordinator,
		}

		go node.handleCoordinatorRequest()

		nodesList = append(nodesList, &node)
	}

	nodesList[0].data = data
	nodesList[1].data = data
	nodesList[2].data = data
	nodesList[3].data = data2
	nodesList[4].data = data2

	coordinator.nodeNetwork = nodesList

	// instantiate client
	client := Client{
		clientChannel: make(chan Data),
	}

	go client.clientGet(&coordinator, "beds")
	go coordinator.handleClientRequest(&client)

	for {
	}
}
