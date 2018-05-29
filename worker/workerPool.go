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

type w struct {
	worker   *worker
	toWorker chan<- messages.Message
}

type WorkerPool struct {
	workers             []*w
	lastAllocatedWorker int
	fromClient          chan messages.Message
}

func getWorkersPerPool() int {
	WorkersPerPool_string, ok := os.LookupEnv("WORKERS_PER_POOL")
	if !ok {
		return DEFAULT_WORKERS_PER_POOL
	}

	workersPerPool, err := strconv.Atoi(WorkersPerPool_string)
	if err != nil {
		return DEFAULT_WORKERS_PER_POOL
	}
	return workersPerPool
}

func CreateWorkerPool() *WorkerPool {
	WorkersPerPool = getWorkersPerPool()
	fmt.Printf("Creating worker-pool with %d workers\n", WorkersPerPool)

	fromClient := make(chan messages.Message, 20)

	// TODO: FIX TO USE NEW CHANNEL STRUCTURE
	workers := make([]*w, WorkersPerPool)

	for i, _ := range workers {
		worker, toWorker := createWorker(fromClient)
		workers[i] = &w{
			worker:   worker,
			toWorker: toWorker,
		}
	}

	workerPool := &WorkerPool{
		workers:             workers,
		lastAllocatedWorker: 0,
		fromClient:          fromClient,
	}

	go workerPool.broadcastMessages()
	return workerPool
}

// TODO: currently using round robin to allocate clients to workers, use some other form of load balancing later
//TODO: should lock workerPool for safety
func (workerPool *WorkerPool) AllocateClient(conn *websocket.Conn) {
	fmt.Printf("allocating client to worker #%d\n", workerPool.lastAllocatedWorker%WorkersPerPool)
	lastAllocatedWorker := workerPool.lastAllocatedWorker

	workerPool.workers[lastAllocatedWorker%WorkersPerPool].worker.addClient(conn)
	workerPool.lastAllocatedWorker = lastAllocatedWorker + 1
}

func (workerPool *WorkerPool) broadcastMessages() {
	for message := range workerPool.fromClient {
		for _, worker := range workerPool.workers {
			worker.toWorker <- message
		}
	}
}
