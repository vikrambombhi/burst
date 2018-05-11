package client

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/messages"
)

type Client struct {
	addr     string
	conn     *websocket.Conn
	msgQueue []*[]byte
}

func New(conn *websocket.Conn, in chan messages.Message, out chan messages.Message) Client {
	client := Client{
		addr:     conn.RemoteAddr().String(),
		conn:     conn,
		msgQueue: make([]*[]byte, 0),
	}
	go readMessages(client, out)
	go writeMessages(client, in)
	return client
}

func readMessages(client Client, out chan<- messages.Message) {
	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("received message: ", message)
		out <- messages.New(message)
	}
}

func writeMessages(client Client, in <-chan messages.Message) {
	for {
		if len(client.msgQueue) != 0 {
			// create copy to avoid locking?
			message := make([]byte, len(*client.msgQueue[0]))
			copy(message, *client.msgQueue[0])

			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println(err)
				return
			}
			client.msgQueue = client.msgQueue[1:]
		}
	}
}
