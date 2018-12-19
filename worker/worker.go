package worker

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/io"
	"github.com/vikrambombhi/burst/messages"
)

type c struct {
	client   io.IO
	toClient chan<- messages.Message
}

type worker struct {
	sync.RWMutex
	clients    []*c
	fromClient chan messages.Message
	newClients chan *websocket.Conn
	logs       *Logs
	offset     int
}

func createWorker(fromClient chan messages.Message, newClients chan *websocket.Conn, logs *Logs) *worker {
	logs.RLock()
	logTail := len(logs.messages)
	logs.RUnlock()

	worker := &worker{
		fromClient: fromClient,
		newClients: newClients,
		logs:       logs,
		offset:     logTail,
	}

	return worker
}

func (worker *worker) setOffSet(offset int) {
	logs.RLock()
	logTail := len(logs.messages) + 1
	logs.RUnlock()

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
	go func(i *int) {
		for {
			for {
				worker.logs.RLock()
				length := len(worker.logs.messages)
				worker.logs.RUnlock()
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
					break
				}
			}

			worker.logs.RLock()
			message := worker.logs.messages[*i]
			worker.logs.RUnlock()

			var wg sync.WaitGroup
			for y := 0; y < len(worker.clients); y++ {
				cl := worker.clients[y]
				if cl.client.GetStatus() == io.STATUS_OPEN {
					wg.Add(1)
					go func(client *c, message messages.Message, wg *sync.WaitGroup) {
						client.toClient <- message
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

func (worker *worker) addFileIO(filename string) {
	fileIOBuilder := io.FileIOBuilder{}
	fileIOBuilder.SetFilename(filename)
	fileIOBuilder.SetReadChannel(worker.fromClient)
	file, toFile, _ := fileIOBuilder.BuildIO()
	c := &c{
		client:   file,
		toClient: toFile,
	}
	worker.Lock()
	worker.clients = append(worker.clients, c)
	fmt.Printf("worker now has %d clients\n", len(worker.clients))
	worker.Unlock()
}
