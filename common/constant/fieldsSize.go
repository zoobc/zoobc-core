package constant

var (
	AccountAddressTypeLength   uint32 = 4
	TransactionSignatureLength uint32 = 4
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
	// SignatureType variables
	SignatureType   uint32 = 4
	AuthRequestType        = 4
	AuthTimestamp          = 8
	// DatasetPropertyLength is max length of string property name in dataset
	DatasetPropertyLength uint32 = 4
	// DatasetValueLength is max length of string property value in dataset
	DatasetValueLength uint32 = 4

	EscrowApproverAddressLength uint32 = 4
	EscrowCommissionLength      uint32 = 8
	EscrowTimeoutLength         uint32 = 8
	EscrowApproval              uint32 = 4
	EscrowID                    uint32 = 8
	EscrowApprovalBytesLength          = EscrowApproval + EscrowID
	EscrowInstructionLength     uint32 = 4
	MultisigFieldLength         uint32 = 4
	// MultiSigFieldMissing indicate fields is missing, no need to read the bytes
	MultiSigFieldMissing uint32
	// MultiSigFieldPresent indicate fields is present, parse the byte accordingly
	MultiSigFieldPresent           uint32 = 1
	MultiSigAddressLength          uint32 = 4
	MultiSigSignatureLength        uint32 = 4
	MultiSigSignatureAddressLength uint32 = 4
	MultiSigNumberOfAddress        uint32 = 4
	MultiSigNumberOfSignatures     uint32 = 4
	MultiSigUnsignedTxBytesLength  uint32 = 4
	MultiSigInfoSize               uint32 = 4
	MultiSigInfoSignatureInfoSize  uint32 = 4
	MultiSigInfoNonce              uint32 = 8
	MultiSigInfoMinSignature       uint32 = 4
	MultiSigTransactionHash        uint32 = 32

	// FeeVote part
	FeeVote              uint32 = 8
	RecentBlockHeight    uint32 = 4
	VoterSignatureLength uint32 = 4

	// Liquid Transaction

	LiquidPaymentCompleteMinutesLength uint32 = 8
	TransactionID                      uint32 = 8
)
