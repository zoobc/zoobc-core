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
	message                    string
	outputType                 string
	version                    uint32
	timestamp                  int64
	senderSeed                 string
	recipientAccountAddressHex string
	fee                        int64
	post                       bool
	postHost                   string
	senderAddressHex           string
	sign                       bool

	// Send money transaction
	sendAmount int64

	// node registration transaction
	nodeSeed                   string
	nodeOwnerAccountAddressHex string
	lockedBalance              int64
	proofOfOwnershipHex        string
	databasePath               string
	databaseName               string

	// dataset transaction
	property string
	value    string
	// escrowable
	escrow               bool
	esApproverAddressHex string
	esCommission         int64
	esTimeout            uint64
	esInstruction        string

	// escrowApproval
	approval      bool
	transactionID int64

	// multiSignature
	nonce            int64
	minSignature     uint32
	participantSeeds []string
	nested           int

	// fee vote
	recentBlockHeight uint32
	feeVote           int64
	dbPath, dBName    string
	// liquidPayment
	completeMinutes uint64
)
