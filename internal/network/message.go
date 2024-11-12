package network

type Message struct {
	Id   string `json:"id"`
	Data Data   `json:"data"`
}

type Data struct {
	Cmd   string      `json:"cmd"`
	Value interface{} `json:"value,omitempty"`
}

func NewMessage(cmd string, value interface{}) Message {
	return Message{Data: Data{Cmd: cmd, Value: value}}
}

func (m Message) CopyId(id string) Message {
	m.Id = id
	return m
}
