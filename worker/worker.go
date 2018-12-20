package worker

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/io"
	"github.com/vikrambombhi/burst/log"
	"github.com/vikrambombhi/burst/messages"
)

type c struct {
	client   io.IO
	toClient chan<- *messages.Message
}

type worker struct {
	sync.RWMutex
	clients    []*c
	fromClient chan messages.Message
	newClients chan *websocket.Conn
	logs       log.Log
	offset     int64
}

func createWorker(fromClient chan messages.Message, newClients chan *websocket.Conn, logs log.Log) *worker {
	logTail := logs.Size()
	worker := &worker{
		fromClient: fromClient,
		newClients: newClients,
		logs:       logs,
		offset:     logTail,
	}

	return worker
}

func (worker *worker) setOffSet(offset int64) {
	logTail := worker.logs.Size()

	if offset < 0 {
		offset = logTail + offset
		if offset < 0 {
			offset = 0
		}
	} else {
		if offset > logTail {
			offset = logTail
		}
	}

	worker.offset = offset
}

func (worker *worker) start() {
	go func(i *int64) {
		for {
			for {
				length := worker.logs.Size()
				if *i >= length {
					// If all caught up to the messages log take time to accept new clients
					select {
					case conn := <-worker.newClients:
						webIOBuilder := io.WebIOBuilder{}
						webIOBuilder.SetConn(conn)
						webIOBuilder.SetReadChannel(worker.fromClient)
						client, toClient, _ := webIOBuilder.BuildIO()
						c := &c{
							client:   client,
							toClient: toClient,
						}
						worker.Lock()
						worker.clients = append(worker.clients, c)
						fmt.Printf("worker now has %d clients\n", len(worker.clients))
						worker.Unlock()
					default:
						time.Sleep(time.Second)
					}
				} else {
					fmt.Println("breaking from accepting new conns")
					break
				}
			}

			fmt.Println("reading message")
			message := worker.logs.Read(*i)
			fmt.Println("read message")

			var wg sync.WaitGroup
			for y := 0; y < len(worker.clients); y++ {
				cl := worker.clients[y]
				if cl.client.GetStatus() == io.STATUS_OPEN {
					wg.Add(1)
					go func(client *c, message messages.Message, wg *sync.WaitGroup) {
						client.toClient <- &message
						wg.Done()
					}(cl, *message, &wg)
				} else {
					worker.clients = append(worker.clients[:y], worker.clients[y+1:]...)
					y--
				}
			}
			wg.Wait()

			*i++
		}
	}(&worker.offset)
}
