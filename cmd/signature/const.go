package signature

var (
	// sign
	seed                string
	dataHex             string
	dataBytes           string
	hash                bool
	senderSignatureType int32
	ed25519UseSpli10    bool

	// verify
	signatureBytes string
	signatureHex   string
	accountAddress string
)
