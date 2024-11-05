package network

import "log"

type Subscriber struct {
	RemoteAddr string
	Publish    chan Message
	Shutdown   chan int
}

type Hub struct {
	Subscribers map[*Subscriber]bool
	Broadcast   chan Message
	Register    chan *Subscriber
	Unregister  chan *Subscriber
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:   make(chan Message),
		Register:    make(chan *Subscriber),
		Unregister:  make(chan *Subscriber),
		Subscribers: make(map[*Subscriber]bool),
	}
}

func (h *Hub) Shutdown() {
	for subscriber := range h.Subscribers {
		h.Unregister <- subscriber
	}
}

func (h *Hub) Run() {
	for {
		select {
		case subscriber := <-h.Register:
			log.Printf("hub: register new subscriber %s", subscriber.RemoteAddr)
			h.Subscribers[subscriber] = true
		case subscriber := <-h.Unregister:
			if _, ok := h.Subscribers[subscriber]; ok {
				log.Printf("hub: unregister subscriber %s", subscriber.RemoteAddr)
				subscriber.Shutdown <- 1
				delete(h.Subscribers, subscriber)
			}
		case message := <-h.Broadcast:
			for subscriber := range h.Subscribers {
				//log.Printf("hub: publish message to subscriber ws://%s", subscriber.remoteAddr)
				select {
				case subscriber.Publish <- message:
				default:
					delete(h.Subscribers, subscriber)
				}
			}
		}
	}
}
