package worker

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/client"
	"github.com/vikrambombhi/burst/messages"
)

type worker struct {
	toClient   chan messages.Message
	fromClient chan messages.Message
	clients    []chan<- messages.Message // TODO: Change name and/or make this more elegant
}

func createWorker(fromClient chan messages.Message) (*worker, chan<- messages.Message) {
	toClient := make(chan messages.Message, 10)
	worker := &worker{
		toClient:   toClient,
		fromClient: fromClient,
	}
	worker.start()
	return worker, toClient
}

// TODO: ensure to worker hasnt already been started
func (worker *worker) start() {
	go func() {
		for {
			select {
			case message := <-worker.toClient:
				for _, client := range worker.clients {
					client <- message
				}
			}
		}
	}()
}

// need to lock worker for safety
func (worker *worker) addClient(conn *websocket.Conn) {
	toClient := client.New(conn, worker.fromClient)
	worker.clients = append(worker.clients, toClient)
	fmt.Printf("worker now has %d clients\n", len(worker.clients))
}
