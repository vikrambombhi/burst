package messages

type Message struct {
	Message     string
	MessageType int
	Flushed     bool
}

func New(message []byte, messageType int) Message {
	return Message{
		Message:     string(message[:]),
		MessageType: messageType,
	}
}

func (message *Message) ToString() string {
	return message.Message
}

func (message *Message) GetType() int {
	return message.MessageType
}
