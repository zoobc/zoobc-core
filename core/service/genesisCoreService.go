package service

import (
	"errors"
	"log"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

type (
	fundReceiver struct {
		accountAddress string
		amount         int64
		nodePublicKey  []byte
		nodeAddress    string
		lockedBalance  int64
	}
)

var (
	// 1 ZOO = 100000000 ZOOBIT, node only know the zoobit representation, zoo representation is handled by frontend
	genesisFundReceiver = []*fundReceiver{
		{
			// 04264418e6f758dc777c33957fd652e048ef388bff51e5b84d505027fead1ca9
			accountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			amount:         1000000000000,
			nodePublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99,
				125, 75, 49, 45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
			nodeAddress:   "0.0.0.0",
			lockedBalance: 1000000,
		},
		{
			// 04266749faa93f9b6a15094c4d89037815455a76f254aeef2ebe4e445a538e0b
			accountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			amount:         1000000000000,
			nodePublicKey: []byte{0, 14, 6, 218, 170, 54, 60, 50, 2, 66, 130, 119, 226, 235, 126,
				203, 5, 12, 152, 194, 170, 146, 43, 63, 224, 101, 127, 241, 62, 152, 187, 255},
			nodeAddress:   "0.0.0.0",
			lockedBalance: 1000000,
		},
		{
			// 04264a2ef814619d4a2b1fa3b45f4aa09b248d53ef07d8e92237f3cc8eb30d6d
			accountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			amount:         1000000000000,
			nodePublicKey: []byte{140, 115, 35, 51, 159, 22, 234, 192, 38, 104, 96, 24, 80, 70, 86, 211, 123,
				72, 52, 221, 97, 121, 59, 151, 158, 90, 167, 17, 110, 253, 122, 158},
			nodeAddress:   "0.0.0.0",
			lockedBalance: 1000000,
		},
		{
			// Wallet Develop
			accountAddress: "nK_ouxdDDwuJiogiDAi_zs1LqeN7f5ZsXbFtXGqGc0Pd",
			amount:         10000000000,
			nodePublicKey: []byte{41, 235, 184, 214, 70, 23, 153, 89, 104, 41, 250, 248, 51, 7, 69, 89,
				234, 181, 100, 163, 45, 69, 152, 70, 52, 201, 147, 70, 6, 242, 52, 220},
			nodeAddress:   "0.0.0.0",
			lockedBalance: 1000000,
		},
	}
	genesisSignature = []byte{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
)

// GetGenesisTransactions return list of genesis transaction to be executed in the
// very beginning of running the blockchain
func GetGenesisTransactions(chainType chaintype.ChainType) []*model.Transaction {
	var genesisTxs []*model.Transaction
	switch chainType.(type) {
	case *chaintype.MainChain:
		for _, fundReceiver := range genesisFundReceiver {
			genesisTx := &model.Transaction{
				Version:                 1,
				TransactionType:         util.ConvertBytesToUint32([]byte{1, 0, 0, 0}),
				Height:                  0,
				Timestamp:               1562806389,
				SenderAccountAddress:    constant.GenesisAccountAddress,
				RecipientAccountAddress: fundReceiver.accountAddress,
				Fee:                     0,
				TransactionBodyLength:   8,
				TransactionBody: &model.Transaction_SendMoneyTransactionBody{
					SendMoneyTransactionBody: &model.SendMoneyTransactionBody{
						Amount: fundReceiver.amount,
					},
				},
				TransactionBodyBytes: util.ConvertUint64ToBytes(uint64(fundReceiver.amount)),
				Signature:            genesisSignature,
			}

			transactionBytes, err := util.GetTransactionBytes(genesisTx, true)
			if err != nil {
				log.Fatal(err)
			}
			transactionHash := sha3.Sum256(transactionBytes)
			genesisTx.TransactionHash = transactionHash[:]
			genesisTx.ID, _ = util.GetTransactionID(transactionHash[:])
			genesisTxs = append(genesisTxs, genesisTx)

			genesisNodeRegistrationTx := GetGenesisNodeRegistrationTx(fundReceiver)
			genesisTxs = append(genesisTxs, genesisNodeRegistrationTx)
		}

		return genesisTxs
	default:
		return nil
	}
}

// GetGenesisNodeRegistrationTx given a fundReceiver, returns a nodeRegistrationTransaction for genesis block
func GetGenesisNodeRegistrationTx(fundReceiver *fundReceiver) *model.Transaction {
	// generate a dummy proof of ownership (avoiding to add conditions to tx parsebytes, for genesis block only)
	poownMessage := &model.ProofOfOwnershipMessage{
		AccountAddress: fundReceiver.accountAddress,
		BlockHash:      make([]byte, 32),
		BlockHeight:    0,
	}
	nodeRegistration := transaction.NodeRegistration{
		Body: &model.NodeRegistrationTransactionBody{
			AccountAddress: fundReceiver.accountAddress,
			LockedBalance:  fundReceiver.lockedBalance,
			NodeAddress:    fundReceiver.nodeAddress,
			NodePublicKey:  fundReceiver.nodePublicKey,
			Poown: &model.ProofOfOwnership{
				MessageBytes: util.GetProofOfOwnershipMessageBytes(poownMessage),
				Signature:    make([]byte, int(constant.NodeSignature+constant.SignatureType)),
			},
		},
	}
	genesisTx := &model.Transaction{
		Version:                 1,
		TransactionType:         util.ConvertBytesToUint32([]byte{2, 0, 0, 0}),
		Height:                  0,
		Timestamp:               1562806389,
		SenderAccountAddress:    constant.GenesisAccountAddress,
		RecipientAccountAddress: fundReceiver.accountAddress,
		Fee:                     0,
		TransactionBodyLength:   nodeRegistration.GetSize(),
		TransactionBody: &model.Transaction_NodeRegistrationTransactionBody{
			NodeRegistrationTransactionBody: nodeRegistration.Body,
		},
		TransactionBodyBytes: nodeRegistration.GetBodyBytes(),
		Signature:            genesisSignature,
	}

	transactionBytes, err := util.GetTransactionBytes(genesisTx, true)
	if err != nil {
		log.Fatal(err)
	}
	transactionHash := sha3.Sum256(transactionBytes)
	genesisTx.TransactionHash = transactionHash[:]
	return genesisTx
}

// AddGenesisAccount create genesis account into `account` and `account_balance` table
func AddGenesisAccount(executor query.ExecutorInterface) error {
	// add genesis account
	genesisAccountBalance := model.AccountBalance{
		AccountAddress:   constant.GenesisAccountAddress,
		BlockHeight:      0,
		SpendableBalance: 0,
		Balance:          0,
		PopRevenue:       0,
		Latest:           true,
	}
	genesisAccountBalanceInsertQ, genesisAccountBalanceInsertArgs := query.NewAccountBalanceQuery().InsertAccountBalance(
		&genesisAccountBalance)
	_ = executor.BeginTx()
	var genesisQueries [][]interface{}
	genesisQueries = append(genesisQueries,
		append(
			[]interface{}{genesisAccountBalanceInsertQ}, genesisAccountBalanceInsertArgs...),
	)
	err := executor.ExecuteTransactions(genesisQueries)
	if err != nil {
		_ = executor.RollbackTx()
		return errors.New("fail to add genesis account balance")
	}
	err = executor.CommitTx()
	if err != nil {
		return err
	}
	return nil
}
