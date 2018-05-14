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

func handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("URL: ", r.URL.Path)
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		log.Println("New connection from: ", conn.RemoteAddr().String())

		topics.AddClient(conn, r.URL.Path)
	})
}

func main() {
	// Register our handler.
	http.Handle("/", handler())
	http.ListenAndServe(":8080", nil)
}
