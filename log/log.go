package log

import (
	"sync"

	"github.com/vikrambombhi/burst/io"
	"github.com/vikrambombhi/burst/messages"
)

type Log struct {
	*log
}

type log struct {
	messages []*messages.Message
	size     int64
	toFile   chan<- *messages.Message
	sync.RWMutex
}

func CreateLog(name string, size int64) Log {
	filename := "/tmp/" + name
	fileIOBuilder := io.FileIOBuilder{}
	fileIOBuilder.SetFilename(filename)
	fileIOBuilder.SetReadChannel(make(chan messages.Message))
	_, toFile, _ := fileIOBuilder.BuildIO()

	log := &log{
		messages: make([]*messages.Message, size),
		size:     size,
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

	return log.messages[position]
}

func (log Log) Write(message *messages.Message) {
	log.Lock()
	log.messages = append(log.messages, message)
	log.size++
	log.Unlock()

	log.toFile <- message
}
