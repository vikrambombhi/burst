package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/topics"
)

var topicName = "testTopic" // Single topic for now

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		log.Println("New connection from: ", conn.RemoteAddr().String())

		topics.AddClient(conn, topicName)
	})
}

func main() {
	// Register our handler.
	http.Handle("/ws", handler())
	http.ListenAndServe(":8080", nil)
}
