package network

type PServer interface {
	ListenAndServe(messages chan<- Message)
	Shutdown()
	Send(msg Message)
}
