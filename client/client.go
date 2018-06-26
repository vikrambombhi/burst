package client

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/messages"
)

type Client struct {
	addr       string
	conn       *websocket.Conn
	toClient   <-chan messages.Message
	fromClient chan<- messages.Message
	status     int
}

const STATUS_CLOSED = 0
const STATUS_OPEN = 1

func New(conn *websocket.Conn, fromClient chan<- messages.Message) (*Client, chan<- messages.Message) {
	toClient := make(chan messages.Message, 10)
	client := &Client{
		addr:       conn.RemoteAddr().String(),
		conn:       conn,
		fromClient: fromClient,
		toClient:   toClient,
		status:     STATUS_OPEN,
	}

	go client.readMessages()
	go client.writeMessages()
	return client, toClient
}

func (client *Client) GetStatus() int {
	return client.status
}

func (client *Client) readMessages() {
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

func (client *Client) writeMessages() {
	for message := range client.toClient {
		msg := []byte(message.ToString())
		//TODO: lock conn for safety
		if err := client.conn.WriteMessage(message.GetType(), msg); err != nil {
			log.Println(err)
			return
		}
	}
}
