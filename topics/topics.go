package topics

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/worker"
)

var Topics map[string]*worker.WorkerPool

func init() {
	log.Println("initalizing topics map")
	Topics = map[string]*worker.WorkerPool{}
}

// TODO: currently using round robin to allocate clients to workers, use some other form of load balancing later
func AddClient(conn *websocket.Conn, topicName string) {
	_, exists := Topics[topicName]
	if !exists {
		Topics[topicName] = worker.CreateWorkerPool()
	}
	Topics[topicName].AllocateClient(conn)
}

// status function
func GetAllTopics() []string {
	topicNames := make([]string, 0, len(Topics))
	for topicName := range Topics {
		topicNames = append(topicNames, topicName)
	}
	return topicNames
}
