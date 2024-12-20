// Package ws implements a WebSocket connection
// to communicate between the server and client.
//
// Deprecated: The WebSocket is too slow, now we have to use udp package.
//
// This package is frozen and no new functionality will be added.
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
	ticker           network.Ticker
}

// Deprecated
func NewPClient(addr string) *PClient {
	return &PClient{
		connectionClosed: false,
		remoteAddr:       addr,
		ticker: network.Ticker{
			Ticker: time.NewTicker(5 * time.Second),
			Done:   make(chan bool),
		},
	}
}

func (c *PClient) ListenAndServe(messages chan<- network.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, fmt.Sprintf("ws://%s/subscribe", c.remoteAddr), nil)
	if err != nil {
		log.Printf("client failed to conn to ws://%s...: %v", c.remoteAddr, err)
		messages <- network.NewMessage("connectionClosed", "failed").WithAddr(c.remoteAddr)
		return
	}
	c.conn = conn
	c.connectionClosed = false

	defer c.closeSubscriberConnection()

	messages <- network.NewMessage("subscribe", nil).WithAddr(c.remoteAddr)

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
				messages <- network.NewMessage("connectionClosed", "normal").WithAddr(c.remoteAddr)
			} else {
				log.Printf("client failed to read to ws://%s...: %v", c.remoteAddr, err)
				messages <- network.NewMessage("connectionClosed", "error").WithAddr(c.remoteAddr)
			}
			return
		} else {
			messages <- message
		}
	}
}

func (c *PClient) ping(messages chan<- network.Message) {
	ping := func() {
		c.Send(network.NewMessage("ping", nil).WithAddr(c.remoteAddr))
		messages <- network.NewMessage("pingServer", nil).WithAddr(c.remoteAddr)
	}

	// send the first ping before waiting the ticker (5 s)
	ping()

	for {
		select {
		case <-c.ticker.Done:
			c.ticker.Ticker.Stop()
			return
		case <-c.ticker.Ticker.C:
			ping()
		}
	}
}

func (c *PClient) closeSubscriberConnection() {
	if c.conn != nil && !c.connectionClosed {
		log.Printf("ws://%s connection closed", c.remoteAddr)

		c.connectionClosed = true
		c.ticker.Done <- true

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

		err := wsjson.Write(ctx, c.conn, msg.WithAddr(c.remoteAddr))
		if err != nil {
			log.Printf("ws://%s failed to send message: %v", c.remoteAddr, err)
		}
	}
}
