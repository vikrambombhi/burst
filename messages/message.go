package messages

type Message struct {
	message     string
	messageType int
}

func New(message []byte, messageType int) Message {
	return Message{
		message:     string(message[:]),
		messageType: messageType,
	}
}

func (message *Message) ToString() string {
	return message.message
}

func (message *Message) GetType() int {
	return message.messageType
}
