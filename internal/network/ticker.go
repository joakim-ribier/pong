package network

import "time"

type Ticker struct {
	Ticker *time.Ticker
	Done   chan bool
}

func (t *Ticker) Run(send func(Message), messages chan<- Message) {
	for {
		select {
		case <-t.Done:
			t.Ticker.Stop()
			return
		case <-t.Ticker.C:
			t.Ping(send, messages)
		}
	}
}

func (t *Ticker) Ping(send func(Message), messages chan<- Message) {
	send(NewSimpleMessage(Ping.String()))
	messages <- NewSimpleMessage(PingAll.String())
}
