package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

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
		if websocket.IsWebSocketUpgrade(r) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Println(err)
				return
			}

			log.Println("New connection from: ", conn.RemoteAddr().String())

			topics.AddClient(conn, r.URL.Path)
		} else {
			http.Error(w, "Server requires connection to be a websocket, use format '/{topic name}'", http.StatusUpgradeRequired)
		}
	})
}

func main() {
	address := flag.String("address", "localhost", "address to run server on")
	port := flag.Int("port", 8080, "port the server will listen for connections on")
	flag.Parse()
	serverAddress := fmt.Sprintf("%s:%d", *address, *port)
	fmt.Printf("%s", serverAddress)

	// Register our handler.
	http.Handle("/get-topics", getTopics())
	http.Handle("/", handler())
	http.ListenAndServe(serverAddress, nil)
}
