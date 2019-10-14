package transactions

import "github.com/zoobc/zoobc-core/common/crypto"

var (
	txTypeMap = map[string][]byte{
		"sendMoney":              {1, 0, 0, 0},
		"registerNode":           {2, 0, 0, 0},
		"updateNodeRegistration": {2, 1, 0, 0},
		"removeNodeRegistration": {2, 2, 0, 0},
		"claimNodeRegistration":  {2, 3, 0, 0},
		"setupAccountDataset":    {3, 0, 0, 0},
		"removeAccountDataset":   {3, 1, 0, 0},
	}
	signature = &crypto.Signature{}

	// Basic transactions data
	outputType              string
	version                 uint32
	timestamp               int64
	senderSeed              string
	recipientAccountAddress string
	fee                     int64

	// Send money transaction
	sendAmount int64

	// node registration transactions
	nodeSeed                string
	nodeOwnerAccountAddress string
	nodeAddress             string
	lockedBalance           int64

	// dataset transactions
	property   string
	value      string
	activeTime uint64
)
