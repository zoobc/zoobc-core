package observer

const (
	// block listener event
	BlockPushed Event = "BlockEvent.BlockPushed"
	// BlockReceived  Event = "BlockEvent.BlockReceived"
	BroadcastBlock Event = "BlockEvent.BroadcastBlock"

	// transaction listener event
	TransactionAdded    Event = "TransactionEvent.TransactionAdded"
	TransactionReceived Event = "TransactionEvent.TransactionReceived"

	// Peer to Peer listener event
	P2PNotifyPeerExplorer Event = "P2PEvent.NotifyPeerExplorer"
)
