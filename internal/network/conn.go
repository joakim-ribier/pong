package network

type Conn interface {
	ListenAndServe(messages chan<- Message)
	Send(msg Message)
	Shutdown()
}
