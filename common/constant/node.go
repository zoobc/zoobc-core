package constant

var (
	// ProofOfOwnershipExpiration number of blocks after a proof of ownership message 'expires' (is considered invalid)
	ProofOfOwnershipExpiration uint32 = 100
	MaxNodeAdmittancePerCycle  uint32 = 1
	NodeAdmissionGenesisDelay  int64  = 1 * 2592000 // 3*2592000 seconds (3 month) in production
	NodeAdmissionBaseDelay     int64  = 3600        // 12*2592000 (1 year) in production
	NodeAdmissionMinDelay      int64  = 60          // 3600 in production
	NodeAdmissionMaxDelay      int64  = 72 * 3600   // 72*3600 (72 hours) in production
)
