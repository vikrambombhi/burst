package topics

import (
	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/worker"
)

var Topics map[string]*worker.WorkerPool

func init() {
	Topics = map[string]*worker.WorkerPool{}
}

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
