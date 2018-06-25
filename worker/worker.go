package worker

import (
	"fmt"
	"sync"
	"time"

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
	logs       *Logs
	sync.RWMutex
}

func createWorker(fromClient chan messages.Message, logs *Logs) (*worker, chan<- messages.Message) {
	toWorker := make(chan messages.Message, 10)

	worker := &worker{
		fromClient: fromClient,
		toWorker:   toWorker,
		logs:       logs,
	}

	go worker.reader()
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

func (worker *worker) reader() {
	i := 0
	for {
		for {
			worker.RLock()
			length := len(worker.logs.messages)
			worker.RUnlock()
			if i >= length {
				time.Sleep(time.Millisecond)
			} else {
				break
			}
		}

		worker.RLock()
		message := worker.logs.messages[i]
		worker.RUnlock()

		var wg sync.WaitGroup
		for _, c := range worker.clients {
			if c.client.GetStatus() == client.STATUS_OPEN {
				wg.Add(1)
				go func(message messages.Message, wg *sync.WaitGroup) {
					c.toClient <- message
					wg.Done()
				}(message, &wg)
			}
		}
		wg.Wait()

		i++
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
