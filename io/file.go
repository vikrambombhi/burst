package io

import (
	"bufio"
	"log"
	"os"

	"github.com/vikrambombhi/burst/messages"
)

type file struct {
	file   *os.File
	toFile <-chan messages.Message
}

func createFileIO(filename string, fromIO chan<- messages.Message) (*file, chan<- messages.Message) {
	toFile := make(chan messages.Message, 10)
	f, err := os.Create(filename)
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
		writer.WriteString(message.ToString())
		writer.Flush()
	}
}

func (file *file) GetStatus() int {
	return STATUS_OPEN
}
