package messages

type Message struct {
	message string
}

func New(message []byte) Message {
	return Message{
		message: string(message[:]),
	}
}

func (message *Message) ToString() string {
	return message.message
}
