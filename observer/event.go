package observer

const (
	// block listener event
	BlockPushed                Event = "BlockEvent.BlockPushed"
	BroadcastBlock             Event = "BlockEvent.BroadcastBlock"
	BlockRequestTransactions   Event = "BlockEvent.BlockRequestTransaction"
	BlockTransactionsRequested Event = "BlockEvent.BlockTransactionsRequested"

	// transaction listener event
	TransactionAdded                   Event = "TransactionEvent.TransactionAdded"
	TransactionReceived                Event = "TransactionEvent.TransactionReceived"
	ReceivedBlockTransactionsValidated Event = "TransactionEvent.ReceivedBlockTransactionsValidated"
	SendBlockTransactions              Event = "TransactionEvent.SendBlockTransactions"
	ExpiringEscrowTransactions         Event = "TransactionEvent.ExpiringEscrowTransaction"
)
