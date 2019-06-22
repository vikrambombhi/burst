package log

import (
	"fmt"
	"sync"
	"time"

	"github.com/vikrambombhi/burst/io"
	"github.com/vikrambombhi/burst/messages"
)

type Log struct {
	*log
}

type log struct {
	messages []*messages.Message
	size     int64
	ringSize int64
	toFile   chan<- *messages.Message
	sync.RWMutex
}

func CreateLog(filename string, size int64) Log {
	fileIOBuilder := io.FileIOBuilder{}
	fileIOBuilder.SetFilename(filename)
	fileIOBuilder.SetReadChannel(make(chan *messages.Message))
	_, toFile, _ := fileIOBuilder.BuildIO()

	log := &log{
		messages: make([]*messages.Message, size),
		size:     0,
		ringSize: size,
		toFile:   toFile,
	}

	return Log{log}
}

func (log Log) Size() int64 {
	log.RLock()
	defer log.RUnlock()

	return log.size
}

func (log Log) Read(position int64) *messages.Message {
	log.RLock()
	defer log.RUnlock()

	return log.messages[position%log.ringSize]
}

// For debugging purposes only
func (log Log) printLog() {
	fmt.Printf("printing log of size: %v\n\n", log.Size())
	for i := 0; int64(i) < log.Size(); i++ {
		fmt.Printf("Message %v at position %v", log.Read(int64(i)).Message, i)
	}
	fmt.Printf("\n\n")
}

func (log Log) Write(message *messages.Message) {
	i := log.size % log.ringSize
	if log.messages[i] == nil {
		log.Lock()
		log.messages[i] = message
		log.size++
		log.Unlock()
	} else {
		for log.messages[i].Flushed == false {
			time.Sleep(5 * time.Second)
		}
		log.Lock()
		log.messages[i] = message
		log.size++
		log.Unlock()
	}
	log.toFile <- message
}
