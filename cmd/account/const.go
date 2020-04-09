package account

var (
	seed string
	// ed25519
	ed25519UseSlip10 bool
	// bitcoin
	bitcoinPrivateKeyLength int32
	bitcoinPublicKeyFormat  int32
	// multisig
	multisigAddresses []string
	multisigMinimSigs uint32
	multiSigNonce     int64
)
