package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var address = flag.String("address", "localhost:8080", "server address to connect too")
var topic = flag.String("topic", "stress", "topic to send/recieve messages on")
var amount = flag.Int("amount", 10000, "number of messages each node should send")
var rate = flag.Int("rate", 1, "rate of messages to send in milliseconds")
var nodes = flag.Int("nodes", 4, "number of nodes to send/recieve messages")

var stats struct {
	data []*recieveTimes
	sync.Mutex
}

type recieveTimes struct {
	times []time.Duration
	sync.Mutex
}

func runNode(done *sync.WaitGroup) {
	var wg sync.WaitGroup
	conn, err := connect(*address, *topic)
	if err != nil {
		log.Fatal("dial:", err)
	}
	wg.Add(1)
	go write(conn, &wg)
	wg.Add(1)
	go read(conn, &wg)
	wg.Wait()

	err = conn.Close()
	if err != nil {
		log.Println("close:", err)
	}
	done.Done()
}

func connect(address string, topic string) (*websocket.Conn, error) {
	t := "/" + topic
	u := url.URL{Scheme: "ws", Host: address, Path: t}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	return c, err
}

func write(conn *websocket.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(time.Millisecond / time.Duration(*rate))
	defer ticker.Stop()

	sentMessages := 0
	for range ticker.C {
		message := []byte(time.Now().Format(time.RFC3339Nano))
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("write:", err)
			return
		}
		sentMessages++
		if sentMessages >= *amount {
			fmt.Printf("worker sent %d messages giving readers time to catch up....\n", sentMessages)
			// Allow reader to read all the messages
			time.Sleep(time.Second * 5)
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
			}
			return
		}
	}
}

func read(conn *websocket.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	stat := &recieveTimes{}

	var timeAppender sync.WaitGroup
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		now := time.Now()
		timeAppender.Add(1)
		go appendTimeSub(stat, message, now, &timeAppender)
	}
	timeAppender.Wait()

	stats.Lock()
	stats.data = append(stats.data, stat)
	stats.Unlock()
}

func appendTimeSub(stat *recieveTimes, message []byte, now time.Time, wg *sync.WaitGroup) {
	defer wg.Done()
	m := string(message[:])

	tfs, err := time.Parse(time.RFC3339Nano, m)
	if err != nil {
		log.Println("Could not parse message", err)
		return
	}
	timeDiff := now.Sub(tfs)
	stat.Lock()
	stat.times = append(stat.times, timeDiff)
	stat.Unlock()
}

func main() {
	flag.Parse()

	var wg sync.WaitGroup
	for i := 0; i < *nodes; i++ {
		wg.Add(1)
		go runNode(&wg)
	}
	wg.Wait()

	var max time.Duration
	var min time.Duration
	var sum time.Duration
	var recieved int

	stats.Lock()
	fmt.Printf("number of nodes that read something: %d\n", len(stats.data))
	min = stats.data[0].times[1]
	for _, data := range stats.data {
		fmt.Printf("number of messages node read: %d\n", len(data.times))
		for _, t := range data.times {
			if t < min {
				min = t
			}
			if t > max {
				max = t
			}
			sum = sum + t
			recieved++
		}
	}
	stats.Unlock()

	avg := sum.Nanoseconds() / int64(recieved)
	fmt.Printf("\n\n\nmax: %v\tmin: %v\tsum: %v\tavg: %vμs\nrecieved %d messages\n", max.String(), min.String(), sum.String(), avg, recieved)
}
