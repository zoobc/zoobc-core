package constant

var (
	// ProofOfOwnershipExpiration number of blocks after a proof of ownership message 'expires' (is considered invalid)
	ProofOfOwnershipExpiration uint32 = 100
	MaxNodeAdmittancePerCycle  uint32 = 1
	// NodeAdmittanceCycle FIXME: this is for testing only. real value should be 1440, if the mainchain is set to
	//       smith every minute on average and if we want to accept new nodes once a day
	NodeAdmittanceCycle uint32 = 2
)

const (
	// NodeRegistered 'registred' node status (= 0): a node in node registry with this status is registered
	NodeRegistered = iota
	// NodeQueued 'queued' node status (= 1): a node in node registry with this status is queued, or 'pending registered'
	NodeQueued
	// NodeDeleted 'deleted' node status (= 2): a node in node registry with this status is marked as deleted
	NodeDeleted
)
