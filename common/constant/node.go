package constant

var (
	// ProofOfOwnershipExpiration number of blocks after a proof of ownership message 'expires' (is considered invalid)
	ProofOfOwnershipExpiration uint32 = 100
	MaxNodeAdmittancePerCycle  uint32 = 1

	NodeAdmissionGenesisDelay int64 = 21 * OneDay  // 3 weeks for beta network
	NodeAdmissionBaseDelay    int64 = OneHour      // 1 year in production
	NodeAdmissionMinDelay     int64 = 60           // 1 hour in production
	NodeAdmissionMaxDelay     int64 = 72 * OneHour // 72 hours in production
)
