package worker

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/messages"
)

type workerAndChan struct {
	worker   *worker
	toWorker chan<- messages.Message
}

// TODO: use env variable to get number of workers per topic, currently set to 1
// TODO/maybe: make worker pools recersive
type WorkerPool struct {
	workersAndChans     [1]*workerAndChan
	lastAllocatedWorker int
	fromClient          chan messages.Message
}

func CreateWorkerPool() *WorkerPool {
	fromClient := make(chan messages.Message, 5)

	workersAndChans := [1]*workerAndChan{}

	for i, _ := range workersAndChans {
		worker, toWorker := createWorker(fromClient)
		workersAndChans[i] = &workerAndChan{
			worker:   worker,
			toWorker: toWorker,
		}
	}

	workerPool := &WorkerPool{
		workersAndChans:     workersAndChans,
		lastAllocatedWorker: 0,
		fromClient:          fromClient,
	}

	workerPool.broadcastMessages()
	return workerPool
}

// need to lock workerPool for safety
func (workerPool *WorkerPool) AllocateClient(conn *websocket.Conn) {
	log.Println("allocating client")
	workerPool.workersAndChans[workerPool.lastAllocatedWorker%1].worker.addClient(conn)
	workerPool.lastAllocatedWorker = workerPool.lastAllocatedWorker + 1
}

func (workerPool *WorkerPool) broadcastMessages() {
	go func(workerPool *WorkerPool) {
		log.Println("listening for messages on channel ", workerPool.fromClient)
		for message := range workerPool.fromClient {
			for _, workerAndChan := range workerPool.workersAndChans {
				workerAndChan.toWorker <- message
			}
		}
	}(workerPool)
}
