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
	"runtime"
	"runtime/pprof"
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
	f, err := os.Create("burst_cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Starting profile")
	pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()
}

func closeProfile() {
	fmt.Println("Closing profile")
	pprof.StopCPUProfile()
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
	cores := flag.Int("cores", -1, "Number of cores to use")
	address := flag.String("address", "0.0.0.0", "address to run server on")
	port := flag.Int("port", 8080, "port the server will listen for connections on")
	shouldProfile := flag.Bool("profile", false, "enable profiling")
	flag.Parse()

	fmt.Println("Cores: ", *cores)
	runtime.GOMAXPROCS(*cores)
	fmt.Printf("Using %d cores\n", runtime.NumCPU())

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

	if *shouldProfile {
		closeProfile()
		f, err := os.Create("burst_mem.prof")
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}

	defer os.Exit(1)

}
