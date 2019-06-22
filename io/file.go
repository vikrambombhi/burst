package io

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"log"
	"os"

	"github.com/vikrambombhi/burst/messages"
)

type file struct {
	file   *os.File
	toFile <-chan *messages.Message
}

func createFileIO(filename string, fromIO chan<- *messages.Message) (*file, chan<- *messages.Message) {
	toFile := make(chan *messages.Message)
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Println("err creating file")
		panic(err)
	}

	file := &file{
		file:   f,
		toFile: toFile,
	}
	go file.writeMessages()
	return file, toFile
}

func (file *file) readMessages() {
}

func (file *file) writeMessages() {
	writer := bufio.NewWriter(file.file)
	for message := range file.toFile {
		// Convert Message.Message to binary
		messageBody := new(bytes.Buffer)
		binary.Write(messageBody, binary.LittleEndian, message.Message)

		// Convert Message.MessageType to binary, we know it will always be < 1 byte
		messageType := make([]byte, 1)
		binary.PutUvarint(messageType, uint64(message.MessageType))

		// Calculate length of message being writen to file
		length := make([]byte, 8)
		l := uint64(messageBody.Len())
		binary.PutUvarint(length, l)

		// Write length of message to file followed by the message itself
		writer.Write(length)
		writer.Write(messageBody.Bytes())
		writer.Write(messageType)

		// TODO: Check if flushing after every message is a bottleneck
		writer.Flush()
		message.Flushed = true
	}
}

func (file *file) GetStatus() int {
	return STATUS_OPEN
}
