package observer

const (
	// block listener event
	BlockPushed   Event = "BlockEvent.BlockPushed"
	BlockReceived Event = "BlockEvent.BlockReceived"

	// transaction listener event
	TransactionAdded    Event = "TransactionEvent.TransactionAdded"
	TransactionReceived Event = "TransactionEvent.TransactionReceived"
)
