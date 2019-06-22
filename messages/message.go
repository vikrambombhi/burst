package messages

type Message struct {
	Message     []byte
	MessageType int
	Flushed     bool
}

func New(message []byte, messageType int) Message {
	return Message{
		Message:     message,
		MessageType: messageType,
	}
}

func (message *Message) ToString() string {
	return string(message.Message)
}

func (message *Message) GetType() int {
	return message.MessageType
}
