package constant

var (
	AccountType       uint32 = 4
	NodeAddressLength uint32 = 4
	AccountID         uint32 = 8
	AccountAddress    uint32 = 44
	// NodePublicKey TODO: this is valid for pub keys generated using Ed25519. in future we might have more implementations
	NodePublicKey uint32 = 32
	Balance       uint32 = 8
	BlockHash     uint32 = 64
	Height        uint32 = 4
	// NodeSignature node always signs using Ed25519 algorithm, that produces 64 bytes signatures
	NodeSignature         uint32 = 64
	TransactionType       uint32 = 4
	TransactionVersion    uint32 = 1
	Timestamp             uint32 = 8
	Fee                   uint32 = 8
	TransactionBodyLength uint32 = 4
	// AccountSignature TODO: this is valid for signatures using Ed25519. in future we might have more implementations
	AccountSignature uint32 = 64
)
