package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Client struct {
	msgQueue []*[]byte
	conn     *websocket.Conn
}

type Topic struct {
	client []Client
	name   string
}

var msgQueue = make([]*[]byte, 0)
var topics = map[string][]Client{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// status function
func printTopics() {
	var i int = 0
	for {
		if len(topics) > 0 && len(topics) != i {
			i++
			for t, _ := range topics {
				log.Println("topic name: ", t)
			}
		}
	}
}

func readMessages(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		msgQueue = append(msgQueue, &message)
	}
}

func writeMessages(conn *websocket.Conn) {
	for {
		if len(msgQueue) != 0 {
			// create copy to avoid locking?
			message := make([]byte, len(*msgQueue[0]))
			copy(message, *msgQueue[0])
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println(err)
				return
			}
			msgQueue = msgQueue[1:]
			log.Println("There are now %i messages remaining in queue", len(msgQueue))
		}
	}
}

func handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		log.Println("New connection from: ", conn.RemoteAddr().String())

		topics["testRoom"] = append(topics["testRoom"], Client{
			msgQueue: make([]*[]byte, 0),
			conn:     conn,
		})

		go readMessages(conn)
		go writeMessages(conn)
	})
}

func main() {
	go printTopics()

	// Register our handler.
	http.Handle("/ws", handler())
	http.ListenAndServe(":8080", nil)
}
