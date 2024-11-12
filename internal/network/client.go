package network

type PClient interface {
	ListenAndServe(messages chan<- Message)
	Send(msg Message)
	Shutdown()
}
