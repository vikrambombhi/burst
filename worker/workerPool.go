package worker

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/messages"
)

const DEFAULT_WORKERS_PER_POOL int = 8

var WorkersPerPool int

type WorkerPool struct {
	sync.RWMutex
	workers             []*worker
	lastAllocatedWorker int
	fromClient          chan messages.Message
}

type Logs struct {
	messages []messages.Message
	sync.RWMutex
}

var logs Logs

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
	workers := make([]*worker, WorkersPerPool)

	for i, _ := range workers {
		worker := createWorker(fromClient, &logs)
		workers[i] = worker
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
func (workerPool *WorkerPool) AllocateClient(conn *websocket.Conn) {
	workerPool.RLock()
	fmt.Printf("allocating client to worker #%d\n", workerPool.lastAllocatedWorker%WorkersPerPool)
	lastAllocatedWorker := workerPool.lastAllocatedWorker
	workerPool.workers[lastAllocatedWorker%WorkersPerPool].addClient(conn)
	workerPool.RUnlock()

	workerPool.Lock()
	workerPool.lastAllocatedWorker = lastAllocatedWorker + 1
	workerPool.Unlock()
}

func (workerPool *WorkerPool) broadcastMessages() {
	for message := range workerPool.fromClient {
		logs.Lock()
		logs.messages = append(logs.messages, message)
		logs.Unlock()
	}
}
