// Constants to generate/validate spine genesis block
package constant

const (
	SpinechainGenesisBlockID int64 = 4021392891451042958
)

var (
	SpinechainGenesisBlocksmithID   = []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	SpinechainGenesisBlockSignature = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	SpinechainGenesisTransactionSignature = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	SpinechainGenesisBlockTimestamp = int64(1562117271)
	SpinechainGenesisAccountAddress = "ZBC_AQTEH7K4_L45WJPLL_HCEC65ZH_7XC5N3XD_YNKPHK45_POH7PQME_AFAFBDWM"
	SpinechainGenesisBlockSeed      = make([]byte, 64)
	SpinechainGenesisNodePublicKey  = make([]byte, 32)
)
