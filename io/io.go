package io

import (
	"github.com/gorilla/websocket"
	"github.com/vikrambombhi/burst/messages"
)

const STATUS_CLOSED = 0
const STATUS_OPEN = 1

type IO interface {
	readMessages()
	writeMessages()
	GetStatus() int
}

// TODO: use builder pattern to create io interface for both web connections and files.
type ioBuilder interface {
	setConn(conn *websocket.Conn)
	setReadChannel(fromIO chan<- messages.Message)
	buildIO() *IO
}
