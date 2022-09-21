package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/topics"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func getOffset(r *http.Request) (int64, error) {
	offset := r.FormValue("offset")

	if offset == "" {
		return -1, nil
	}

	i, err := strconv.Atoi(offset)
	if err != nil {
		return -1, err
	}

	return int64(i), nil
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
		offset, err := getOffset(r)

		if err != nil {
			http.Error(w, "offset is not valid", http.StatusBadRequest)
			return
		}
		if websocket.IsWebSocketUpgrade(r) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				http.Error(w, "Connection upgrade failed", http.StatusInternalServerError)
				return
			}
			log.Println("New connection from: ", conn.RemoteAddr().String())

			topicName := strings.Replace(r.URL.Path, "/", "", -1)
			topics.AddClient(conn, topicName, offset)
		} else {
			http.Error(w, "Server requires connection to be a websocket, use format '/{topic name}'", http.StatusUpgradeRequired)
		}
	})
}

func profile() {
	pprofMux := http.DefaultServeMux
	http.DefaultServeMux = http.NewServeMux()

	// Pprof server.
	go func() {
		log.Println(http.ListenAndServe("localhost:8081", pprofMux))
	}()
}

func runServer(address string, port int, done chan bool, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	mux := http.NewServeMux()
	mux.Handle("/get-topics", getTopics())
	mux.Handle("/", handler())

	serverAddress := fmt.Sprintf("%s:%d", address, port)
	fmt.Printf("Starting server on %s\n", serverAddress)

	server := &http.Server{
		Addr:    serverAddress,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	<-done
	gracefullCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err := server.Shutdown(gracefullCtx); err != nil {
		log.Printf("shutdown error: %v\n", err)
	} else {
		log.Printf("gracefully stopped\n")
	}
}

func main() {
	address := flag.String("address", "0.0.0.0", "address to run server on")
	port := flag.Int("port", 8080, "port the server will listen for connections on")
	shouldProfile := flag.Bool("profile", false, "enable profiling")
	flag.Parse()

	if *shouldProfile {
		profile()
	}

	var wg sync.WaitGroup
	shutDownServer := make(chan bool)
	go runServer(*address, *port, shutDownServer, &wg)

	signalChan := make(chan os.Signal, 1)

	signal.Notify(
		signalChan,
		syscall.SIGHUP,  // kill -SIGHUP XXXX
		syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
	)

	<-signalChan
	log.Print("os.Interrupt - shutting down...\n")
	shutDownServer <- true
	wg.Wait()

	defer os.Exit(1)

}
