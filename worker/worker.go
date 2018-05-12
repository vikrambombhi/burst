package worker

import (
	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/client"
	"github.com/vikrambombhi/burst/messages"
)

type Worker struct {
	queue   chan messages.Message
	clients []chan<- messages.Message // TODO: Change name and/or make this more elegant
}

func New() *Worker {
	worker := &Worker{}
	worker.start()
	return worker
}

// TODO: ensure to worker hasnt already been started
func (worker *Worker) start() {
	worker.queue = make(chan messages.Message, 10)
	go func() {
		for {
			select {
			case message := <-worker.queue:
				for _, client := range worker.clients {
					client <- message
				}
			}
		}
	}()
}

func (worker *Worker) AddClient(conn *websocket.Conn) {
	clientInput := client.New(conn, worker.queue)
	worker.clients = append(worker.clients, clientInput)
}
