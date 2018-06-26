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
	sync.RWMutex
	clients    []*c
	fromClient chan messages.Message
	logs       *Logs
}

func createWorker(fromClient chan messages.Message, logs *Logs) *worker {
	worker := &worker{
		fromClient: fromClient,
		logs:       logs,
	}

	go worker.reader()
	return worker
}

func (worker *worker) reader() {
	i := 0
	for {
		for {
			worker.logs.RLock()
			length := len(worker.logs.messages)
			worker.logs.RUnlock()
			if i >= length {
				time.Sleep(time.Millisecond)
			} else {
				break
			}
		}

		worker.logs.RLock()
		message := worker.logs.messages[i]
		worker.logs.RUnlock()

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
