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
	tailWorkers         []*worker
	offsetWorkers       []*worker
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
		worker.start()
		workers[i] = worker
	}

	workerPool := &WorkerPool{
		tailWorkers:         workers,
		lastAllocatedWorker: 0,
		fromClient:          fromClient,
	}

	go workerPool.broadcastMessages()
	return workerPool
}

// TODO: currently using round robin to allocate clients to workers, use some other form of load balancing later
func (workerPool *WorkerPool) AllocateClient(conn *websocket.Conn, offset int) {
	if offset == -1 {
		workerPool.Lock()
		defer workerPool.Unlock()

		lastAllocatedWorker := workerPool.lastAllocatedWorker
		workerPool.tailWorkers[lastAllocatedWorker%WorkersPerPool].addClient(conn)
		workerPool.lastAllocatedWorker = lastAllocatedWorker + 1
	} else {
		worker := createWorker(workerPool.fromClient, &logs)
		worker.setOffSet(offset)
		worker.start()

		worker.addClient(conn)

		workerPool.Lock()
		workerPool.offsetWorkers = append(workerPool.offsetWorkers, worker)
		workerPool.Unlock()
	}
}

func (workerPool *WorkerPool) broadcastMessages() {
	for message := range workerPool.fromClient {
		logs.Lock()
		logs.messages = append(logs.messages, message)
		logs.Unlock()
	}
}
