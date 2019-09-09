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

// GetGenesisTransactions return list of genesis transaction to be executed in the
// very beginning of running the blockchain
func GetGenesisTransactions(chainType chaintype.ChainType) []*model.Transaction {
	var genesisTxs []*model.Transaction
	switch chainType.(type) {
	case *chaintype.MainChain:
		for _, fundReceiver := range constant.MainchainGenesisFundReceivers {
			// send funds from genesis account to the fund receiver
			genesisTx := &model.Transaction{
				Version:                 1,
				TransactionType:         util.ConvertBytesToUint32([]byte{1, 0, 0, 0}),
				Height:                  0,
				Timestamp:               1562806389,
				SenderAccountAddress:    constant.MainchainGenesisAccountAddress,
				RecipientAccountAddress: fundReceiver.AccountAddress,
				Fee:                     0,
				TransactionBodyLength:   8,
				TransactionBody: &model.Transaction_SendMoneyTransactionBody{
					SendMoneyTransactionBody: &model.SendMoneyTransactionBody{
						Amount: fundReceiver.Amount,
					},
				},
				TransactionBodyBytes: util.ConvertUint64ToBytes(uint64(fundReceiver.Amount)),
				Signature:            constant.MainchainGenesisBlockSignature,
			}

			transactionBytes, err := util.GetTransactionBytes(genesisTx, true)
			if err != nil {
				log.Fatal(err)
			}
			transactionHash := sha3.Sum256(transactionBytes)
			genesisTx.TransactionHash = transactionHash[:]
			genesisTx.ID, _ = util.GetTransactionID(transactionHash[:])
			genesisTxs = append(genesisTxs, genesisTx)

			// register the node for the fund receiver
			genesisNodeRegistrationTx := GetGenesisNodeRegistrationTx(fundReceiver.AccountAddress, fundReceiver.NodeAddress,
				fundReceiver.LockedBalance, fundReceiver.NodePublicKey)
			genesisTxs = append(genesisTxs, genesisNodeRegistrationTx)
		}

		return genesisTxs
	default:
		return nil
	}
}

// GetGenesisNodeRegistrationTx given a fundReceiver, returns a nodeRegistrationTransaction for genesis block
func GetGenesisNodeRegistrationTx(accountAddress, nodeAddress string, lockedBalance int64, nodePublicKey []byte) *model.Transaction {
	// generate a dummy proof of ownership (avoiding to add conditions to tx parsebytes, for genesis block only)
	poownMessage := &model.ProofOfOwnershipMessage{
		AccountAddress: accountAddress,
		BlockHash:      make([]byte, 32),
		BlockHeight:    0,
	}
	nodeRegistration := transaction.NodeRegistration{
		Body: &model.NodeRegistrationTransactionBody{
			AccountAddress: accountAddress,
			LockedBalance:  lockedBalance,
			NodeAddress:    nodeAddress,
			NodePublicKey:  nodePublicKey,
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
		SenderAccountAddress:    constant.MainchainGenesisAccountAddress,
		RecipientAccountAddress: accountAddress,
		Fee:                     0,
		TransactionBodyLength:   nodeRegistration.GetSize(),
		TransactionBody: &model.Transaction_NodeRegistrationTransactionBody{
			NodeRegistrationTransactionBody: nodeRegistration.Body,
		},
		TransactionBodyBytes: nodeRegistration.GetBodyBytes(),
		Signature:            constant.MainchainGenesisBlockSignature,
	}

	transactionBytes, err := util.GetTransactionBytes(genesisTx, true)
	if err != nil {
		log.Fatal(err)
	}
	transactionHash := sha3.Sum256(transactionBytes)
	genesisTx.TransactionHash = transactionHash[:]
	genesisTx.ID, _ = util.GetTransactionID(transactionHash[:])
	return genesisTx
}

// AddGenesisAccount create genesis account into `account` and `account_balance` table
func AddGenesisAccount(executor query.ExecutorInterface) error {
	// add genesis account
	genesisAccountBalance := model.AccountBalance{
		AccountAddress:   constant.MainchainGenesisAccountAddress,
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
