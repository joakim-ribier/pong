package drawer

import (
	"image/color"
	"time"
)

// networkMessageLevel represents type of a log level
type networkMessageLevel int

const (
	info networkMessageLevel = iota
	logg
	warning
)

// networkMessage represents a message to display to the user
type networkMessage struct {
	dateTime string
	text     string
	level    networkMessageLevel
}

// asColor maps the {level} field to a {color}
func (t networkMessage) asColor(colors map[string]color.Color) color.Color {
	switch t.level {
	case info:
		return colors["#atari3"]
	case warning:
		return colors["#atari1"]
	default:
		return colors["black"]
	}
}

type readyToPlay struct {
	ready        bool
	nbTpsLaps    int
	nbSeconds    int
	nbSecondsMax int
}

// networkClient represents a network client (subscriber)
type networkClient struct {
	networkAddr       string
	lastPing          time.Time
	lastPong          time.Time
	nbPingAttempts    int
	nbPingMaxAttempts int
	version           string
}

// newRemoteClient builds a new {networkClient} type
func newRemoteClient(networkAddr string) *networkClient {
	return &networkClient{
		networkAddr:       networkAddr,
		nbPingAttempts:    0,
		nbPingMaxAttempts: 3,
	}
}

// networkData represents data for the network game
type networkData struct {
	clients     map[string]*networkClient
	messages    []networkMessage
	readyToPlay readyToPlay
}

// newNetworkData builds a new {networkData} type
func newNetworkData() *networkData {
	return &networkData{
		clients:     make(map[string]*networkClient),
		readyToPlay: readyToPlay{ready: false, nbSeconds: 0, nbTpsLaps: 0, nbSecondsMax: 3},
	}
}

// ping computes the time elapsed between ping and pong in ms
func (t networkClient) ping() time.Duration {
	if t.lastPong.After(t.lastPing) {
		return t.lastPong.Sub(t.lastPing)
	}
	return 0 * time.Millisecond
}
