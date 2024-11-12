package udp

import (
	"log"
	"net"
	"time"

	"github.com/joakim-ribier/go-utils/pkg/jsonsutil"
	"github.com/joakim-ribier/pong/internal/network"
	"github.com/joakim-ribier/pong/pkg"
)

// UDPClient represents a client connection
type UDPClient struct {
	serverAddr *net.UDPAddr

	conn             *net.UDPConn
	connectionClosed bool
	ticker           network.Ticker
}

// NewClient builds a new {UDPClient} type
func NewClient(serverAddr string) *UDPClient {
	return &UDPClient{
		connectionClosed: false,
		serverAddr:       pkg.ToUDPAddrUnsafe(serverAddr),
		ticker: network.Ticker{
			Ticker: time.NewTicker(5 * time.Second),
			Done:   make(chan bool),
		},
	}
}

// ListenAndServe listens messages from UDP network on a specific port
// and serves them to the application
func (c *UDPClient) ListenAndServe(messages chan<- network.Message) {
	go c.ticker.Run(c.Send, messages)

	addr := &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 0,
		Zone: "",
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Printf("fail to listen on udp://%s: %v", addr.String(), err)
	}
	c.conn = conn
	c.connectionClosed = false

	defer c.close()
	log.Printf("listening on udp://%s network...", conn.LocalAddr())

	// subscribe the client to the server
	c.Send(network.NewSimpleMessage(network.Subscribe.String()))
	time.Sleep(500 * time.Millisecond)

	c.ticker.Ping(c.Send, messages)
	c.read(messages)
}

// read reads messages from network connection
// and it notifies the application with the {messages} chan
func (c *UDPClient) read(messages chan<- network.Message) {
	for {
		buf := make([]byte, 1024)
		size, networkAddr, err := c.conn.ReadFrom(buf)
		if err != nil {
			continue
		}

		message, err := jsonsutil.Unmarshal[network.Message](buf[:size])
		if err != nil {
			log.Printf("fail to read message from udp://%s: %v", networkAddr.String(), err)
			continue
		}

		log.Printf("read message from udp://%s: %v", networkAddr.String(), message)
		messages <- message.WithAddr(networkAddr.String())

		if message.AsCMD() == network.Shutdown {
			return
		}
	}
}

// close closes the UDP connection
func (c *UDPClient) close() {
	if c.conn != nil && !c.connectionClosed {
		log.Printf("close the udp://%s connection", c.conn.LocalAddr().String())

		c.connectionClosed = true
		c.ticker.Done <- true
		c.Send(network.NewSimpleMessage(network.Shutdown.String())) // notify the server before shutdown

		time.Sleep(1 * time.Second)
		err := c.conn.Close()
		if err != nil {
			log.Printf("fail to close udp://%s connection: %v", c.serverAddr.String(), err)
		}
	}
}

// Shutdown closes the UDP connection
func (c *UDPClient) Shutdown() {
	c.close()
}

// Send sends the {network.Message} to the server
func (c *UDPClient) Send(msg network.Message) {
	if c.conn != nil {
		log.Printf("send a message to udp://%s: %v", c.serverAddr.String(), msg)

		bytes, _ := jsonsutil.Marshal(msg)
		_, err := c.conn.WriteTo(bytes, c.serverAddr)
		if err != nil {
			log.Printf("fail to send message to udp://%s: %v", c.serverAddr.String(), err)
		}
	}
}
