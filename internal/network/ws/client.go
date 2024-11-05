package ws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/joakim-ribier/pong/internal/network"
)

type PClient struct {
	remoteAddr string

	conn             *websocket.Conn
	connectionClosed bool
	ticker           ticker
}

type ticker struct {
	ticker *time.Ticker
	done   chan bool
}

func NewPClient(addr string) *PClient {
	return &PClient{
		connectionClosed: false,
		remoteAddr:       addr,
		ticker: ticker{
			ticker: time.NewTicker(5 * time.Second),
			done:   make(chan bool),
		},
	}
}

func (c *PClient) Conn(messages chan<- network.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, fmt.Sprintf("ws://%s/subscribe", c.remoteAddr), nil)
	if err != nil {
		log.Printf("client failed to conn to ws://%s...: %v", c.remoteAddr, err)
		messages <- network.NewMessage("connectionClosed", "failed").CopyId(c.remoteAddr)
		return
	}
	c.conn = conn
	c.connectionClosed = false

	defer c.closeSubscriberConnection()

	messages <- network.NewMessage("subscribe", nil).CopyId(c.remoteAddr)

	go c.ping(messages)
	c.read(messages)
}

func (c *PClient) read(messages chan<- network.Message) {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var message network.Message
		err := wsjson.Read(ctx, c.conn, &message)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				messages <- network.NewMessage("connectionClosed", "normal").CopyId(c.remoteAddr)
			} else {
				log.Printf("client failed to read to ws://%s...: %v", c.remoteAddr, err)
				messages <- network.NewMessage("connectionClosed", "error").CopyId(c.remoteAddr)
			}
			return
		} else {
			messages <- message
		}
	}
}

func (c *PClient) ping(messages chan<- network.Message) {
	ping := func() {
		c.Send(network.NewMessage("ping", nil).CopyId(c.remoteAddr))
		messages <- network.NewMessage("pingServer", nil).CopyId(c.remoteAddr)
	}

	// send the first ping before waiting the ticker (5 s)
	ping()

	for {
		select {
		case <-c.ticker.done:
			c.ticker.ticker.Stop()
			return
		case <-c.ticker.ticker.C:
			ping()
		}
	}
}

func (c *PClient) closeSubscriberConnection() {
	if c.conn != nil && !c.connectionClosed {
		log.Printf("ws://%s connection closed", c.remoteAddr)

		c.connectionClosed = true
		c.ticker.done <- true

		err := c.conn.Close(websocket.StatusNormalClosure, "user exits the game")
		if err != nil {
			log.Printf("failed to close connection of ws://%s...%v", c.remoteAddr, err)
		}
	}
}

// Shutdown closed the connection to the web sockets
func (c *PClient) Shutdown() {
	c.closeSubscriberConnection()
}

// Send sends a command to the web socket
func (c *PClient) Send(msg network.Message) {
	if c.conn != nil {
		//log.Printf("sends new message to ws://%s: %v", c.remoteAddr, msg)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
		defer cancel()

		err := wsjson.Write(ctx, c.conn, msg.CopyId(c.remoteAddr))
		if err != nil {
			log.Printf("ws://%s failed to send message: %v", c.remoteAddr, err)
		}
	}
}
