package topics

import (
	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/worker"
)

var Topics map[string][]*worker.Worker

func AddClient(conn *websocket.Conn, topicName string) {
	topic := Topics[topicName]
	worker := worker.New()
	worker.AddClient(conn)
	topic = append(topic, worker)
}

// status function
func GetAllTopics() []string {
	topicNames := make([]string, 0, len(Topics))
	for topicName := range Topics {
		topicNames = append(topicNames, topicName)
	}
	return topicNames
}
