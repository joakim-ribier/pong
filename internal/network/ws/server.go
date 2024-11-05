package ws

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/joakim-ribier/pong/internal/network"
)

type PServer struct {
	remoteAddr string
	closed     bool

	hub    *network.Hub
	ticker ticker

	serveMux   http.ServeMux
	httpServer *http.Server
}

func NewPServer(addr string) *PServer {
	return &PServer{
		remoteAddr: addr,
		hub:        network.NewHub(),
		ticker: ticker{
			ticker: time.NewTicker(15 * time.Second),
			done:   make(chan bool),
		},
	}
}

func (s *PServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.serveMux.ServeHTTP(w, r)
}

func (s *PServer) ListenAndServe(messages chan<- network.Message) {
	go s.hub.Run()
	go s.ping(messages, false)

	s.closed = false

	l, err := net.Listen("tcp", s.remoteAddr)
	if err != nil {
		log.Printf("failed to listen: %v", err)
	}

	log.Printf("listening on ws://%s", s.remoteAddr)

	s.serveMux.Handle("/", http.FileServer(http.Dir(".")))
	s.serveMux.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		s.subscribe(w, r, messages)
	})

	s.httpServer = &http.Server{
		Handler:      s,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	errc := make(chan error, 1)
	go func() {
		errc <- s.httpServer.Serve(l)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	select {
	case err := <-errc:
		if !s.closed {
			log.Printf("ws://%s failed to serve: %v", s.remoteAddr, err)
			s.Shutdown()
		}
	case sig := <-sigs:
		log.Printf("ws://%s terminating (%v)", s.remoteAddr, sig)
		s.Shutdown()
	}
}

func (s *PServer) ping(messages chan<- network.Message, oneshot bool) {

	ping := func() {
		s.Send(network.NewMessage("ping", nil))
		messages <- network.NewMessage("pingClients", nil)
	}

	if oneshot {
		ping()
		return
	}

	for {
		select {
		case <-s.ticker.done:
			s.ticker.ticker.Stop()
			return
		case <-s.ticker.ticker.C:
			ping()
		}
	}
}

func (s *PServer) Shutdown() {
	if !s.closed {
		s.hub.Shutdown()
		s.closed = true
		log.Printf("try to shutdown ws://%s...", s.remoteAddr)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Printf("ws://%s shutdown failed: %v", s.remoteAddr, err)
			s.closed = false
		} else {
			log.Printf("ws://%s closed!", s.remoteAddr)
			s.remoteAddr = ""
		}
	}
}

func (s *PServer) Send(msg network.Message) {
	s.hub.Broadcast <- msg.CopyId(s.remoteAddr)
}

func (s *PServer) SendTo(subscriberId string, msg network.Message) {
	for subscriber := range s.hub.Subscribers {
		if subscriber.RemoteAddr == subscriberId {
			subscriber.Publish <- msg
			break
		}
	}
}

// subscribeHandler accepts the WebSocket connection and then subscribes
// it to all future messages.
func (s *PServer) subscribe(w http.ResponseWriter, r *http.Request, messages chan<- network.Message) {
	log.Printf("ws://%s subscribe", r.RemoteAddr)

	subscriber := &network.Subscriber{
		RemoteAddr: r.RemoteAddr,
		Publish:    make(chan network.Message, 16),
		Shutdown:   make(chan int, 1),
	}
	s.hub.Register <- subscriber

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("ws://%s failed to serve: %v", subscriber.RemoteAddr, err)
		return
	}

	// send the message to refresh the UI
	messages <- network.NewMessage("subscribe", nil).CopyId(subscriber.RemoteAddr)
	s.ping(messages, true)

	go s.publish(conn, subscriber, messages)
	s.read(conn, subscriber, messages)
}

func (s *PServer) closeSubscriberConnection(conn *websocket.Conn, subscriber *network.Subscriber, messages chan<- network.Message) {
	log.Printf("subscriber ws://%s connection closed", subscriber.RemoteAddr)
	messages <- network.NewMessage("unsubscribe", nil).CopyId(subscriber.RemoteAddr)

	err := conn.Close(websocket.StatusNormalClosure, "server exits the game")
	if err != nil {
		log.Printf("failed to close connection of ws://%s...%v", subscriber.RemoteAddr, err)
	}
}

func (s *PServer) publish(conn *websocket.Conn, subscriber *network.Subscriber, messages chan<- network.Message) {
	defer s.closeSubscriberConnection(conn, subscriber, messages)

	for {
		select {
		case <-subscriber.Shutdown:
			return
		case msg := <-subscriber.Publish:
			//log.Printf("ws://%s sends new message to ws://%s: %v", s.RemoteAddr, subscriber.RemoteAddr, msg)

			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
			defer cancel()

			err := wsjson.Write(ctx, conn, msg.CopyId(s.remoteAddr))
			if err != nil {
				log.Printf("ws://%s failed to send: %v", subscriber.RemoteAddr, err)
			}
		}
	}
}

func (s *PServer) read(conn *websocket.Conn, subscriber *network.Subscriber, messages chan<- network.Message) {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		var message network.Message
		err := wsjson.Read(ctx, conn, &message)
		if err != nil {
			s.hub.Unregister <- subscriber

			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				messages <- network.NewMessage("connectionClosed", "normal").CopyId(subscriber.RemoteAddr)
			} else {
				log.Printf("ws://%s failed to read: %v", subscriber.RemoteAddr, err)
				messages <- network.NewMessage("connectionClosed", "error").CopyId(subscriber.RemoteAddr)
			}
			return
		} else {
			messages <- message.CopyId(subscriber.RemoteAddr)
		}
	}
}
