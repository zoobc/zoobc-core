package observer

const (
	// block listener event
	BlockPushed    Event = "BlockEvent.BlockPushed"
	BroadcastBlock Event = "BlockEvent.BroadcastBlock"

	// transaction listener event
	TransactionAdded              Event = "TransactionEvent.TransactionAdded"
	TransactionReceived           Event = "TransactionEvent.TransactionReceived"
	ReceivedTransactionValidated  Event = "TransactionEvent.ReceivedTransactionValidated"
	NeedDeleteTransactionCadidate Event = "TransactionEvent.NeedDeleteTransactionCadidate"
)
