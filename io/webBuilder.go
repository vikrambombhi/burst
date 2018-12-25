package io

import (
	"errors"

	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/messages"
)

type WebIOBuilder struct {
	conn        *websocket.Conn
	readChannel chan<- *messages.Message
}

func (w *WebIOBuilder) SetConn(conn *websocket.Conn) {
	w.conn = conn
}

func (w *WebIOBuilder) SetReadChannel(fromIO chan<- *messages.Message) {
	w.readChannel = fromIO
}

func (w *WebIOBuilder) BuildIO() (IO, chan<- *messages.Message, error) {
	if w.conn == nil || w.readChannel == nil {
		return nil, nil, errors.New("connection and/or read channel needs to be set")
	}
	web, toClient := createWebIO(w.conn, w.readChannel)
	return web, toClient, nil
}
