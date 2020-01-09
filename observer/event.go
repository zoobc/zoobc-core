package observer

const (
	// block listener event
	BlockPushed              Event = "BlockEvent.BlockPushed"
	BroadcastBlock           Event = "BlockEvent.BroadcastBlock"
	BlockRequestTransactions Event = "BlockEvent.BlockRequestTransaction"

	// transaction listener event
	TransactionAdded    Event = "TransactionEvent.TransactionAdded"
	TransactionReceived Event = "TransactionEvent.TransactionReceived"
)
