package main

import (
	"encoding/json"
	"fmt"
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

func getTopics() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retValue, err := json.Marshal(topics.GetAllTopics())
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Fprintf(w, "%s", retValue)
	})
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
	http.Handle("/get-topics", getTopics())
	http.Handle("/", handler())
	http.ListenAndServe(":8080", nil)
}
