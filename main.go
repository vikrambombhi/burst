package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/topics"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// status function
func printTopics() {
	log.Println(topics.GetAllTopics())
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

	})
}

func main() {
	go printTopics()

	// Register our handler.
	http.Handle("/ws", handler())
	http.ListenAndServe(":8080", nil)
}
