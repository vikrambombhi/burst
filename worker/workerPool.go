package worker

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/log"
	"github.com/vikrambombhi/burst/messages"
)

type WorkerPool struct {
	// TODO: implement latest worker. Only return latest message
	// latestWorkers       []*worker
	fromClient       chan messages.Message
	newClientCurrent chan *websocket.Conn
	// newClienLatest    chan *websocket.Conn
}

var logs log.Log

func CreateWorkerPool(name string) *WorkerPool {
	logs = log.CreateLog(name, 100)

	fromClient := make(chan messages.Message, 20)
	newClientCurrent := make(chan *websocket.Conn, 0)

	workerPool := &WorkerPool{
		fromClient:       fromClient,
		newClientCurrent: newClientCurrent,
	}

	go workerPool.broadcastMessages()
	return workerPool
}

func (workerPool *WorkerPool) AllocateClient(conn *websocket.Conn, offset int64) {
	if offset == -1 {
		for {
			select {
			// Try having one of the existing workers accept the new client
			case workerPool.newClientCurrent <- conn:
				return
			// If no workers are available create a new one
			case <-time.After(time.Second):
				worker := createWorker(workerPool.fromClient, workerPool.newClientCurrent, logs)
				worker.setOffSet(offset)
				worker.start()
			}
		}
	} else {
		// Create new worker for all clients not starting at current message
		customClient := make(chan *websocket.Conn, 1)
		worker := createWorker(workerPool.fromClient, customClient, logs)
		worker.setOffSet(offset)
		worker.start()

		customClient <- conn
	}
}

func (workerPool *WorkerPool) broadcastMessages() {
	for message := range workerPool.fromClient {
		logs.Write(&message)
	}
}
