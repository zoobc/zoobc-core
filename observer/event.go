package observer

const (
	// block listener event
	BlockPushed   Event = "BlockEven.BlockPushed"
	BlockReceived Event = "BlockEven.BlockReceived"

	// transaction listener event
	TransactionAdded    Event = "TransactionEven.TransactionAdded"
	TransactionReceived Event = "TransactionEven.TransactionReceived"
)
