package constant

var (
	// ProofOfOwnershipExpiration number of blocks after a proof of ownership message 'expires' (is considered invalid)
	ProofOfOwnershipExpiration uint32 = 100
	MaxNodeAdmittancePerCycle  uint32 = 1
	// NodeAdmittanceCycle FIXME: this is for testing only. real value should be 1440, if the mainchain is set to
	//       smith every minute on average and if we want to accept new nodes once a day
	NodeAdmittanceCycle       uint32 = 3
	NodeAdmissionGenesisDelay int64  = 12 * 2592000 // 3*2592000 seconds (3 month) in production
	NodeAdmissionBaseDelay    int64  = 3600         // 12*2592000 (1 year) in production
	NodeAdmissionMinDelay     int64  = 60           // 3600 in production
	NodeAdmissionMaxDelay     int64  = 72 * 3600    // 72*3600 (72 hours) in production
)
