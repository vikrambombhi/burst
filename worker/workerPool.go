package worker

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/messages"
)

const DEFAULT_WORKERS_PER_POOL int = 5

var WorkersPerPool int

type workerAndChan struct {
	worker   *worker
	toWorker chan<- messages.Message
}

// TODO/maybe: make worker pools recersive
type WorkerPool struct {
	workersAndChans     []*workerAndChan
	lastAllocatedWorker int
	fromClient          chan messages.Message
}

func CreateWorkerPool() *WorkerPool {
	WorkersPerPool_string, ok := os.LookupEnv("WORKERS_PER_POOL")
	if !ok {
		WorkersPerPool = DEFAULT_WORKERS_PER_POOL
	} else {
		var err error
		WorkersPerPool, err = strconv.Atoi(WorkersPerPool_string)
		if err != nil {
			WorkersPerPool = DEFAULT_WORKERS_PER_POOL
		}
	}

	fmt.Printf("Creating worker-pool with %d workers\n", WorkersPerPool)

	fromClient := make(chan messages.Message, 5)
	workersAndChans := make([]*workerAndChan, WorkersPerPool)

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

// TODO: currently using round robin to allocate clients to workers, use some other form of load balancing later
//TODO: should lock workerPool for safety
func (workerPool *WorkerPool) AllocateClient(conn *websocket.Conn) {
	fmt.Printf("allocating client to worker #%d\n", workerPool.lastAllocatedWorker%WorkersPerPool)
	workerPool.workersAndChans[workerPool.lastAllocatedWorker%WorkersPerPool].worker.addClient(conn)
	workerPool.lastAllocatedWorker = workerPool.lastAllocatedWorker + 1
}

func (workerPool *WorkerPool) broadcastMessages() {
	go func(workerPool *WorkerPool) {
		for message := range workerPool.fromClient {
			for _, workerAndChan := range workerPool.workersAndChans {
				workerAndChan.toWorker <- message
			}
		}
	}(workerPool)
}
