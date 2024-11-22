package network

import (
	"log"
	"sync"
)

type Subscriber struct {
	NetworkAddr string
	Publish     chan Message
	Shutdown    chan int
}

type Hub struct {
	Subscribers map[string]*Subscriber
	Broadcast   chan Message
	Register    chan *Subscriber
	Unregister  chan string
	mu          sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:   make(chan Message),
		Register:    make(chan *Subscriber),
		Unregister:  make(chan string),
		Subscribers: make(map[string]*Subscriber),
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
			h.mu.Lock()
			log.Printf("hub: register new subscriber [%s]", subscriber.NetworkAddr)
			h.Subscribers[subscriber.NetworkAddr] = subscriber
			h.mu.Unlock()
		case networkAddr := <-h.Unregister:
			h.mu.Lock()
			if subscriber, ok := h.Subscribers[networkAddr]; ok {
				log.Printf("hub: unregister subscriber [%s]", networkAddr)
				subscriber.Shutdown <- 1
				delete(h.Subscribers, networkAddr)
			}
			h.mu.Unlock()
		case message := <-h.Broadcast:
			for networkAddr, subscriber := range h.Subscribers {
				log.Printf("hub: publish message to subscriber [%s]", networkAddr)
				select {
				case subscriber.Publish <- message:
				default:
					delete(h.Subscribers, networkAddr)
				}
			}
		}
	}
}
