package global

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func NewMessage(typ string, payload interface{}) Message {
	return Message{
		Type:    typ,
		Payload: payload,
	}
}
