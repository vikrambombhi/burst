package messages

type Message struct {
	message     []byte
	messageType int
	flushed     bool
}

func New(message []byte, messageType int) Message {
	return Message{
		message:     message,
		messageType: messageType,
	}
}

func (message *Message) GetMessage() []byte {
	return message.message
}

func (message *Message) GetType() int {
	return message.messageType
}

func (message *Message) IsFlushed() bool {
	return message.flushed
}

func (message *Message) MarkAsFlushed() {
	message.flushed = true
}
