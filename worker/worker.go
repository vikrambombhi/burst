package worker

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/client"
	"github.com/vikrambombhi/burst/messages"
)

type c struct {
	client   *client.Client
	toClient chan<- messages.Message
}

type worker struct {
	clients    []*c
	fromClient chan messages.Message
	toWorker   <-chan messages.Message
	sync.RWMutex
}

func createWorker(fromClient chan messages.Message) (*worker, chan<- messages.Message) {
	toWorker := make(chan messages.Message, 10)

	worker := &worker{
		fromClient: fromClient,
		toWorker:   toWorker,
	}

	go worker.start()
	return worker, toWorker
}

// TODO: ensure to worker hasnt already been started
func (worker *worker) start() {
	for message := range worker.toWorker {
		worker.RLock()
		for _, c := range worker.clients {
			if c.client.GetStatus() == client.STATUS_OPEN {
				go func(message messages.Message) {
					c.toClient <- message
				}(message)
			}
		}
		worker.RUnlock()
	}
}

// need to lock worker for safety
func (worker *worker) addClient(conn *websocket.Conn) {
	client, toClient := client.New(conn, worker.fromClient)
	c := &c{
		client:   client,
		toClient: toClient,
	}
	worker.Lock()
	worker.clients = append(worker.clients, c)
	fmt.Printf("worker now has %d clients\n", len(worker.clients))
	worker.Unlock()
}
