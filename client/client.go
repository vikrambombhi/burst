package client

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/messages"
)

type Client struct {
	addr string
	conn *websocket.Conn
}

func New(conn *websocket.Conn, queue chan<- messages.Message) chan<- messages.Message {
	in := make(chan messages.Message, 10)
	client := Client{
		addr: conn.RemoteAddr().String(),
		conn: conn,
	}

	go client.readMessages(queue)
	go client.writeMessages(in)
	return in
}

func (client *Client) readMessages(queue chan<- messages.Message) {
	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		queue <- messages.New(message)
	}
}

func (client *Client) writeMessages(in <-chan messages.Message) {
	for message := range in {
		msgal := []byte(message.ToString())
		//TODO: lock conn for safety
		if err := client.conn.WriteMessage(websocket.TextMessage, msgal); err != nil {
			log.Println(err)
			return
		}
	}
}
