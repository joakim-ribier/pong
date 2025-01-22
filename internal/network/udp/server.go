package udp

import (
	"log"
	"net"
	"time"

	"github.com/joakim-ribier/go-utils/pkg/jsonsutil"
	"github.com/joakim-ribier/pong/internal/network"
	"github.com/joakim-ribier/pong/pkg"
)

// UDPServer represents a server connection
type UDPServer struct {
	networkAddr *net.UDPAddr
	closed      bool

	hub    *network.Hub
	ticker network.Ticker

	conn *net.UDPConn
}

// NewServer builds a new {UDPServer} type
func NewServer(serverAddr string) *UDPServer {
	return &UDPServer{
		networkAddr: pkg.ToUDPAddrUnsafe(serverAddr),
		hub:         network.NewHub(),
		ticker: network.Ticker{
			Ticker: time.NewTicker(5 * time.Second),
			Done:   make(chan bool),
		},
	}
}

// ListenAndServe listens messages from UDP network on a specific port
// and serves them to the application
func (s *UDPServer) ListenAndServe(messages chan<- network.Message) {
	go s.hub.Run()
	go s.ticker.Run(s.Send, messages)

	conn, err := net.ListenUDP("udp", s.networkAddr)
	if err != nil {
		log.Printf("fail to listen on udp://%s: %v", s.networkAddr.String(), err)
	}
	s.conn = conn
	s.closed = false

	log.Printf("listening on udp://%s network...", conn.LocalAddr())
	s.read(messages)
}

// Shutdown closes the UDP connection
// and it sends a message to all subscribers
func (s *UDPServer) Shutdown() {
	if !s.closed {
		log.Printf("close the udp://%s connection", s.networkAddr)
		s.hub.Shutdown()
		s.closed = true

		time.Sleep(1 * time.Second)
		if err := s.conn.Close(); err != nil {
			log.Printf("fail to close udp://%s connection: %v", s.networkAddr, err)
			s.closed = false
		}
	}
}

// Send sends the {network.Message} to the specific subscriber
// or it broadcasts the message to all subscribers
func (s *UDPServer) Send(msg network.Message) {
	if subscriber, ok := s.hub.Subscribers[msg.NetworkAddr]; ok {
		subscriber.Publish <- msg
	} else {
		s.hub.Broadcast <- msg
	}
}

// read reads messages from network connection
// and it notifies the application with the {messages} chan
func (s *UDPServer) read(messages chan<- network.Message) {
	for {
		buf := make([]byte, 1024)
		size, remoteAddr, err := s.conn.ReadFrom(buf)
		if err != nil {
			continue
		}

		message, err := jsonsutil.Unmarshal[network.Message](buf[:size])
		if err != nil {
			log.Printf("fail to read message from udp://%s: %v", remoteAddr.String(), err)
			continue
		}
		log.Printf("read message from udp://%s: %v", remoteAddr.String(), message)

		switch message.AsCMD() {
		case network.Subscribe:
			subscriber := &network.Subscriber{
				NetworkAddr: remoteAddr.String(),
				Publish:     make(chan network.Message, 16),
				Shutdown:    make(chan int, 1),
			}
			s.hub.Register <- subscriber
			go s.listen(subscriber)
		case network.Shutdown:
			s.hub.Unregister <- remoteAddr.String()
		}

		messages <- message.WithAddr(remoteAddr.String())

		// send immediatly the first [ping] after [subscribe] message
		if message.AsCMD() == network.Subscribe {
			s.ticker.Ping(s.Send, messages)
		}
	}
}

// listen listens messages on a specific subscriber
func (s *UDPServer) listen(subscriber *network.Subscriber) {
	for {
		select {
		case <-subscriber.Shutdown:
			subscriber.Publish <- network.NewSimpleMessage(network.Shutdown.String())
		case msg := <-subscriber.Publish:
			log.Printf("send a message to udp://%s: %v", subscriber.NetworkAddr, msg)

			bytes, _ := jsonsutil.Marshal(msg)
			_, err := s.conn.WriteTo(bytes, pkg.ToUDPAddrUnsafe(subscriber.NetworkAddr))
			if err != nil {
				log.Printf("fail to send message to udp://%s: %v", s.networkAddr, err)
			}
		}
	}
}
