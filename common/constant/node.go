package constant

var (
	// ProofOfOwnershipExpiration number of blocks after a proof of ownership message 'expires' (is considered invalid)
	ProofOfOwnershipExpiration uint32 = 100
	MaxNodeAdmittancePerCycle  uint32 = 1
	// NodeAdmittanceCycle FIXME: this is for testing only. real value should be 1440, if the mainchain is set to
	//       smith every minute on average and if we want to accept new nodes once a day
	NodeAdmittanceCycle       uint32 = 2
	DeletedNodeAccountAddress string = "00000000000000000000000000000000000000000000"
)

const (
	NodeRegistered = iota
	NodeQueued
	NodeDeleted
)
