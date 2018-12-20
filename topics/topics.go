package topics

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/worker"
)

type topics struct {
	workerPools map[string]*worker.WorkerPool
	sync.RWMutex
}

var t *topics

func init() {
	t = &topics{
		workerPools: make(map[string]*worker.WorkerPool),
	}
}

func AddClient(conn *websocket.Conn, topicName string, offset int64) {
	t.RLock()
	_, exists := t.workerPools[topicName]
	t.RUnlock()

	if !exists {
		t.Lock()
		t.workerPools[topicName] = worker.CreateWorkerPool(topicName)
		t.Unlock()
	}

	t.workerPools[topicName].AllocateClient(conn, offset)
}

func GetAllTopics() []string {
	t.RLock()
	defer t.RUnlock()
	topicNames := make([]string, 0, len(t.workerPools))
	for topicName := range t.workerPools {
		topicNames = append(topicNames, topicName)
	}
	return topicNames
}
