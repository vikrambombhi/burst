package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

var address = flag.String("address", "localhost:8080", "server address to connect too")
var topic = flag.String("topic", "chat", "topic to send/recieve messages on")

func connect(address string, topic string) (*websocket.Conn, error) {
	t := "/" + topic
	u := url.URL{Scheme: "ws", Host: address, Path: t}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	return c, err
}

func write(conn *websocket.Conn, messages chan string) {
	for message := range messages {
		msg := []byte(message)
		err := conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("write:", err)
			return
		}
	}
}

func read(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("Recieved: %s", string(message))
	}
}

func main() {
	flag.Parse()
	conn, err := connect(*address, *topic)
	if err != nil {
		log.Fatal("connecting to server:", err)
	}

	messages := make(chan string)
	go write(conn, messages)
	go read(conn)

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Type message and hit enter to send")
	fmt.Println("Enter 'close' to close")
	fmt.Println("---------------------")

	for {
		text, _ := reader.ReadString('\n')
		fmt.Print("\r")

		textDelimed := strings.Replace(text, "\n", "", -1)

		if strings.Compare("close", textDelimed) == 0 {
			break
		}

		messages <- text
	}

	close(messages)

	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
	}

	err = conn.Close()
	if err != nil {
		log.Println("close:", err)
	}
}
