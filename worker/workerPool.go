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
	// TODO: implement latest worker. Only return latest message
	// latestWorkers       []*worker
	fromClient       chan messages.Message
	newClientCurrent chan *websocket.Conn
	// newClienLatest    chan *websocket.Conn
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
	newClientCurrent := make(chan *websocket.Conn, 1)

	// TODO: FIX TO USE NEW CHANNEL STRUCTURE
	workers := make([]*worker, WorkersPerPool)

	for i := range workers {
		worker := createWorker(fromClient, newClientCurrent, &logs)
		worker.start()
		workers[i] = worker
	}

	workerPool := &WorkerPool{
		fromClient:       fromClient,
		newClientCurrent: newClientCurrent,
	}

	go workerPool.broadcastMessages()
	return workerPool
}

func (workerPool *WorkerPool) AllocateClient(conn *websocket.Conn, offset int) {
	if offset == -1 {
		workerPool.newClientCurrent <- conn
	} else {
		// Create new worker for all clients not starting at current message
		customClient := make(chan *websocket.Conn, 1)
		worker := createWorker(workerPool.fromClient, customClient, &logs)
		worker.setOffSet(offset)
		worker.start()

		customClient <- conn
	}
}

func (workerPool *WorkerPool) AllocateFile(filename string, offset int) {
	worker := createWorker(workerPool.fromClient, workerPool.newClientCurrent, &logs)
	worker.setOffSet(0)
	worker.start()

	worker.addFileIO(filename)
}

func (workerPool *WorkerPool) broadcastMessages() {
	for message := range workerPool.fromClient {
		logs.Lock()
		logs.messages = append(logs.messages, message)
		logs.Unlock()
	}
}
