package io

import (
	"errors"

	"github.com/vikrambombhi/burst/messages"
)

type FileIOBuilder struct {
	filename    string
	readChannel chan<- *messages.Message
}

func (f *FileIOBuilder) SetFilename(filename string) {
	f.filename = filename
}

func (f *FileIOBuilder) SetReadChannel(fromIO chan<- *messages.Message) {
	f.readChannel = fromIO
}

func (f *FileIOBuilder) BuildIO() (IO, chan<- *messages.Message, error) {
	if f.filename == "" || f.readChannel == nil {
		return nil, nil, errors.New("filename and/or read channel needs to be set")
	}
	file, toClient := createFileIO(f.filename, f.readChannel)
	return file, toClient, nil
}
