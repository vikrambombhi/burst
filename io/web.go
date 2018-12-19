package io

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/messages"
)

type web struct {
	conn       *websocket.Conn
	fromClient chan<- messages.Message
	toClient   <-chan *messages.Message
	status     int
}

func createWebIO(conn *websocket.Conn, fromIO chan<- messages.Message) (*web, chan<- *messages.Message) {
	toClient := make(chan *messages.Message, 10)
	web := &web{
		conn:       conn,
		fromClient: fromIO,
		toClient:   toClient,
		status:     STATUS_OPEN,
	}

	go web.readMessages()
	go web.writeMessages()
	return web, toClient
}

func (client *web) GetStatus() int {
	return client.status
}

func (client *web) readMessages() {
	for {
		messageType, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				client.status = STATUS_CLOSED
			}
			log.Println(err)
			client.conn.Close()
			return
		}
		if messageType == websocket.CloseMessage {
			client.status = STATUS_CLOSED
			client.conn.Close()
			return
		}

		go func(message messages.Message) {
			client.fromClient <- message
		}(messages.New(message, messageType))
	}
}

func (client *web) writeMessages() {
	for message := range client.toClient {
		msg := []byte(message.ToString())
		if err := client.conn.WriteMessage(message.GetType(), msg); err != nil {
			log.Println(err)
			return
		}
	}
}
