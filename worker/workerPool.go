package worker

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/messages"
)

type WorkerPool struct {
	// TODO: implement latest worker. Only return latest message
	// latestWorkers       []*worker
	fromClient       chan messages.Message
	newClientCurrent chan *websocket.Conn
	// newClienLatest    chan *websocket.Conn
}

type Logs struct {
	messages []*messages.Message
	sync.RWMutex
}

var logs Logs

func CreateWorkerPool() *WorkerPool {
	fromClient := make(chan messages.Message, 20)
	newClientCurrent := make(chan *websocket.Conn, 1)

	workerPool := &WorkerPool{
		fromClient:       fromClient,
		newClientCurrent: newClientCurrent,
	}

	go workerPool.broadcastMessages()
	return workerPool
}

func (workerPool *WorkerPool) AllocateClient(conn *websocket.Conn, offset int) {
	if offset == -1 {
		for {
			select {
			// Try having one of the existing workers accept the new client
			case workerPool.newClientCurrent <- conn:
				return
			// If no workers are available create a new one
			case <-time.After(time.Second):
				worker := createWorker(workerPool.fromClient, workerPool.newClientCurrent, &logs)
				worker.setOffSet(offset)
				worker.start()
			}
		}
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
		logs.messages = append(logs.messages, &message)
		logs.Unlock()
	}
}
