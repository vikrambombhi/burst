package topics

import (
	"github.com/vikrambombhi/burst/client"
	"github.com/vikrambombhi/burst/messages"
)

type Topic struct {
	client   []client.Client
	msgQueue []*messages.Message
	name     string
}

var Topics map[string]Topic

func AddClient(topicName string, conn *websocket.Conn) {
	in := make(chan messages.Message)
	out := make(chan messages.Message)
	client := client.New(conn)

	topic, exists := Topics[topicName]
	if exists {
		topic.client = append(topic.client, client)
	} else {
		topic.client = append(topic.client, client)
		topic.msgQueue = make([]*messages.Message, 0)
		topic.name = topicName
	}
}

func BroadcastMessage(topicName string, message string) {
}

// status function
func GetAllTopics() []string {
	topicNames := make([]string, 0, len(Topics))
	for topicName := range Topics {
		topicNames = append(topicNames, topicName)
	}
	return topicNames
}
