package messages

type Message struct {
	message string
}

func New(message []byte) Message {
	return Message{
		message: string(message[:]),
	}
}
