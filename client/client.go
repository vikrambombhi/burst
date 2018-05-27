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
}

func New(conn *websocket.Conn, fromClient chan<- messages.Message) chan<- messages.Message {
	toClient := make(chan messages.Message, 10)
	client := Client{
		addr:       conn.RemoteAddr().String(),
		conn:       conn,
		fromClient: fromClient,
		toClient:   toClient,
	}

	go client.readMessages()
	go client.writeMessages()
	return toClient
}

func (client *Client) readMessages() {
	for {
		messageType, message, err := client.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("recived message sending in channel: ", client.fromClient)
		client.fromClient <- messages.New(message, messageType)
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
