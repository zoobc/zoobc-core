package transaction

import (
	"github.com/zoobc/zoobc-core/common/crypto"
)

var (
	txTypeMap = map[string][]byte{
		"sendMoney":              {1, 0, 0, 0},
		"registerNode":           {2, 0, 0, 0},
		"updateNodeRegistration": {2, 1, 0, 0},
		"removeNodeRegistration": {2, 2, 0, 0},
		"claimNodeRegistration":  {2, 3, 0, 0},
		"setupAccountDataset":    {3, 0, 0, 0},
		"removeAccountDataset":   {3, 1, 0, 0},
		"approvalEscrow":         {4, 0, 0, 0},
		"multiSignature":         {5, 0, 0, 0},
		"liquidPayment":          {6, 0, 0, 0},
		"liquidPaymentStop":      {6, 1, 0, 0},
		"feeVoteCommit":          {7, 0, 0, 0},
		"feeVoteReveal":          {7, 1, 0, 0},
	}
	signature = &crypto.Signature{}

	// Basic transaction data
	outputType              string
	version                 uint32
	timestamp               int64
	senderSeed              string
	recipientAccountAddress string
	fee                     int64
	post                    bool
	postHost                string
	senderAddress           string
	senderSignatureType     int32
	sign                    bool

	// Send money transaction
	sendAmount int64

	// node registration transaction
	nodeSeed                string
	nodeOwnerAccountAddress string
	nodeAddress             string
	lockedBalance           int64
	proofOfOwnershipHex     string
	databasePath            string
	databaseName            string

	// dataset transaction
	property string
	value    string
	// escrowable
	escrow            bool
	esApproverAddress string
	esCommission      int64
	esTimeout         uint64
	esInstruction     string

	// escrowApproval
	approval      bool
	transactionID int64

	// multiSignature
	unsignedTxHex     string
	addressSignatures map[string]string
	txHash            string
	addresses         []string
	nonce             int64
	minSignature      uint32

	// fee vote
	recentBlockHeight uint32
	feeVote           int64
	dbPath, dBName    string
	// liquidPayment
	completeMinutes uint64
)
