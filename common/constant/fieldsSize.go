package constant

var (
	AccountAddressLength      uint32 = 4
	NodeAddressLength         uint32 = 4
	AccountAddress            uint32 = 44
	AccountAddressEmptyLength uint32
	// NodePublicKey TODO: this is valid for pub keys generated using Ed25519. in future we might have more implementations
	NodePublicKey uint32 = 32
	Balance       uint32 = 8
	BlockHash     uint32 = 32
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
	// SignatureType variables
	SignatureType   uint32 = 4
	AuthRequestType        = 4
	AuthTimestamp          = 8
	// DatasetPropertyLength is max length of string property name in dataset
	DatasetPropertyLength uint32 = 4
	// DatasetValueLength is max length of string property value in dataset
	DatasetValueLength uint32 = 4

	EscrowApproverAddressLength   uint32 = 4
	EscrowCommissionLength        uint32 = 8
	EscrowTimeoutLength           uint32 = 8
	EscrowApproval                uint32 = 4
	EscrowID                      uint32 = 8
	EscrowApprovalBytesLength            = EscrowApproval + EscrowID
	EscrowInstructionLength       uint32 = 4
	MultiSigAddressLength         uint32 = 4
	MultiSigSignatureLength       uint32 = 4
	MultiSigNumberOfAddress       uint32 = 4
	MultiSigNumberOfSignatures    uint32 = 4
	MultiSigUnsignedTxBytesLength uint32 = 4
	MultiSigInfoNonce             uint32 = 8
	MultiSigInfoMinSignature      uint32 = 4
)
