package io

import (
	"encoding/gob"
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
	encoder := gob.NewEncoder(file.file)
	for message := range file.toFile {
		err := encoder.Encode(message)
		if err != nil {
			log.Fatal("encode error:", err)
		}
	}
}

func (file *file) GetStatus() int {
	return STATUS_OPEN
}
