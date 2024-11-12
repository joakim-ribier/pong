package network

type Message struct {
	NetworkAddr string `json:"networkAddr"`
	Data        Data   `json:"data"`
}

type Data struct {
	Cmd   string      `json:"cmd"`
	Value interface{} `json:"value,omitempty"`
}

func NewMessage(cmd string, value interface{}) Message {
	return Message{Data: Data{Cmd: cmd, Value: value}}
}

func NewSimpleMessage(cmd string) Message {
	return NewMessage(cmd, nil)
}

func (m Message) AsCMD() CMD {
	return toCMD(m.Data.Cmd)
}

func (m Message) WithAddr(v string) Message {
	m.NetworkAddr = v
	return m
}

type CMD int

const (
	Notify CMD = iota
	Ping
	PingAll
	Pong
	Ready
	Shutdown
	Subscribe
	UpdateCurrentState
	UpdatePaddleY
)

func (c CMD) String() string {
	switch c {
	case Notify:
		return "Notify"
	case Ping:
		return "Ping"
	case PingAll:
		return "PingAll"
	case Pong:
		return "Pong"
	case Ready:
		return "Ready"
	case Shutdown:
		return "Shutdown"
	case Subscribe:
		return "Subscribe"
	case UpdateCurrentState:
		return "UpdateCurrentState"
	case UpdatePaddleY:
		return "UpdatePaddleY"
	default:
		return "Unknown"
	}
}

func toCMD(v string) CMD {
	switch v {
	case "Notify":
		return Notify
	case "Ping":
		return Ping
	case "PingAll":
		return PingAll
	case "Pong":
		return Pong
	case "Ready":
		return Ready
	case "Shutdown":
		return Shutdown
	case "Subscribe":
		return Subscribe
	case "UpdateCurrentState":
		return UpdateCurrentState
	case "UpdatePaddleY":
		return UpdatePaddleY
	default:
		return -1
	}
}
