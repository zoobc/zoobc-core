package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/accounttype"
	"golang.org/x/crypto/sha3"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
)

var transactionUtil = &transaction.Util{}

// GetGenesisTransactions return list of genesis transaction to be executed in the
// very beginning of running the blockchain
func GetGenesisTransactions(
	chainType chaintype.ChainType,
	genesisEntries []constant.GenesisConfigEntry,
) ([]*model.Transaction, error) {
	var genesisTxs []*model.Transaction
	switch chainType.(type) {
	case *chaintype.MainChain:
		for _, genesisEntry := range genesisEntries {
			// pass to genesis the fullAddress (accountType + accountPublicKey) in bytes
			fullAccountAddress, err := accounttype.ParseEncodedAccountToAccountAddress(
				int32(model.AccountType_ZbcAccountType),
				genesisEntry.AccountAddress,
			)
			if err != nil {
				return nil, err
			}
			accType, err := accounttype.NewAccountTypeFromAccount(fullAccountAddress)
			if err != nil {
				return nil, err
			}
			accountFullAddress, err := accType.GetAccountAddress()
			if err != nil {
				return nil, err
			}
			// send funds from genesis account to the fund receiver if the `accountBalance` is non-zero
			if uint64(genesisEntry.AccountBalance) != 0 {
				genesisTx := &model.Transaction{
					Version:                 1,
					TransactionType:         util.ConvertBytesToUint32([]byte{1, 0, 0, 0}),
					Height:                  0,
					Timestamp:               constant.MainchainGenesisBlockTimestamp,
					SenderAccountAddress:    constant.MainchainGenesisAccountAddress,
					RecipientAccountAddress: accountFullAddress,
					Fee:                     0,
					TransactionBodyLength:   8,
					TransactionBody: &model.Transaction_SendMoneyTransactionBody{
						SendMoneyTransactionBody: &model.SendMoneyTransactionBody{
							Amount: genesisEntry.AccountBalance,
						},
					},
					TransactionBodyBytes: util.ConvertUint64ToBytes(uint64(genesisEntry.AccountBalance)),
					Signature:            constant.MainchainGenesisTransactionSignature,
				}

				transactionBytes, err := transactionUtil.GetTransactionBytes(genesisTx, true)
				if err != nil {
					return nil, err
				}
				transactionHash := sha3.Sum256(transactionBytes)
				genesisTx.TransactionHash = transactionHash[:]
				genesisTx.ID, err = transactionUtil.GetTransactionID(transactionHash[:])
				if err != nil {
					return nil, err
				}
				genesisTxs = append(genesisTxs, genesisTx)
			}

			// register the node for the fund receiver, if relative element in GenesisConfig contains a NodePublicKey
			if len(genesisEntry.NodePublicKey) > 0 {
				genesisNodeRegistrationTx, err := GetGenesisNodeRegistrationTx(accountFullAddress, genesisEntry.NodeAddress,
					genesisEntry.LockedBalance, genesisEntry.NodePublicKey)
				if err != nil {
					return nil, err
				}
				genesisTxs = append(genesisTxs, genesisNodeRegistrationTx)
			}
		}

		return genesisTxs, nil
	case *chaintype.SpineChain:
		return make([]*model.Transaction, 0), nil
	default:
		return nil, blocker.NewBlocker(
			blocker.AppErr,
			"GetGenesisTransactions:ChainTypeNotFound",
		)
	}
}

// GetGenesisNodeRegistrationTx given a genesisEntry, returns a nodeRegistrationTransaction for genesis block
func GetGenesisNodeRegistrationTx(
	accountAddress []byte,
	nodeAddress string,
	lockedBalance int64,
	nodePublicKey []byte,
) (*model.Transaction, error) {
	// generate a dummy proof of ownership (avoiding to add conditions to tx parsebytes, for genesis block only)
	poownMessage := &model.ProofOfOwnershipMessage{
		AccountAddress: accountAddress,
		BlockHash:      make([]byte, 32),
		BlockHeight:    0,
	}

	nodeRegistrationQuery := query.NewNodeRegistrationQuery()
	nodeRegistration := transaction.NodeRegistration{
		SenderAddress: constant.MainchainGenesisAccountAddress,
		Body: &model.NodeRegistrationTransactionBody{
			AccountAddress: accountAddress,
			LockedBalance:  lockedBalance,
			NodePublicKey:  nodePublicKey,
			Poown: &model.ProofOfOwnership{
				MessageBytes: util.GetProofOfOwnershipMessageBytes(poownMessage),
				Signature:    make([]byte, int(constant.NodeSignature+constant.SignatureType)),
			},
		},
		NodeRegistrationQuery: nodeRegistrationQuery,
	}
	nrSize, err := nodeRegistration.GetSize()
	if err != nil {
		return nil, err
	}

	nodeRegistrationTxBodyBytes, err := nodeRegistration.GetBodyBytes()
	if err != nil {
		return nil, err
	}

	genesisTx := &model.Transaction{
		Version:                 1,
		TransactionType:         util.ConvertBytesToUint32([]byte{2, 0, 0, 0}),
		Height:                  0,
		Timestamp:               constant.MainchainGenesisBlockTimestamp,
		SenderAccountAddress:    constant.MainchainGenesisAccountAddress,
		RecipientAccountAddress: accountAddress,
		Fee:                     0,
		TransactionBodyLength:   nrSize,
		TransactionBody: &model.Transaction_NodeRegistrationTransactionBody{
			NodeRegistrationTransactionBody: nodeRegistration.Body,
		},
		TransactionBodyBytes: nodeRegistrationTxBodyBytes,
		Signature:            constant.MainchainGenesisTransactionSignature,
	}

	transactionBytes, err := transactionUtil.GetTransactionBytes(genesisTx, true)
	if err != nil {
		return nil, err
	}
	transactionHash := sha3.Sum256(transactionBytes)
	genesisTx.TransactionHash = transactionHash[:]
	genesisTx.ID, err = transactionUtil.GetTransactionID(transactionHash[:])
	if err != nil {
		return nil, err
	}
	return genesisTx, nil
}

// AddGenesisNextNodeAdmission create genesis next node admission timestamp
func AddGenesisNextNodeAdmission(
	executor query.ExecutorInterface,
	genesisBlockTimestamp int64,
	nextNodeAdmissionTimestampStorage storage.CacheStorageInterface,
) error {
	var (
		err           error
		nodeAdmission = &model.NodeAdmissionTimestamp{
			Timestamp:   genesisBlockTimestamp + constant.NodeAdmissionGenesisDelay,
			BlockHeight: 0,
			Latest:      true,
		}
		insertQueries = query.NewNodeAdmissionTimestampQuery().InsertNextNodeAdmission(nodeAdmission)
	)
	err = executor.BeginTx()
	if err != nil {
		return err
	}
	err = executor.ExecuteTransactions(insertQueries)
	if err != nil {

		rollbackErr := executor.RollbackTx()
		if rollbackErr != nil {
			log.Errorln(rollbackErr.Error())
		}
		return blocker.NewBlocker(blocker.AppErr, "fail to add genesis next node admission timestamp")

	}
	err = executor.CommitTx()
	if err != nil {
		return err
	}
	// update storage cache of next node admission timestamp
	err = nextNodeAdmissionTimestampStorage.SetItem(nil, *nodeAdmission)
	if err != nil {
		return err
	}
	return nil
}

// AddGenesisAccount create genesis account into `account` and `account_balance` table
func AddGenesisAccount(executor query.ExecutorInterface) error {
	var (
		// add genesis account
		genesisAccountBalance = model.AccountBalance{
			AccountAddress:   constant.MainchainGenesisAccountAddress,
			BlockHeight:      0,
			SpendableBalance: 0,
			Balance:          0,
			PopRevenue:       0,
			Latest:           true,
		}
		genesisAccountBalanceInsertQ, genesisAccountBalanceInsertArgs = query.NewAccountBalanceQuery().InsertAccountBalance(
			&genesisAccountBalance)

		genesisQueries [][]interface{}
		err            error
	)

	err = executor.BeginTx()
	if err != nil {
		return err
	}
	genesisQueries = append(genesisQueries,
		append(
			[]interface{}{genesisAccountBalanceInsertQ}, genesisAccountBalanceInsertArgs...),
	)
	err = executor.ExecuteTransactions(genesisQueries)
	if err != nil {
		rollbackErr := executor.RollbackTx()
		if rollbackErr != nil {
			log.Errorln(rollbackErr.Error())
		}
		return blocker.NewBlocker(blocker.AppErr, "fail to add genesis account balance")
	}
	err = executor.CommitTx()
	if err != nil {
		return err
	}
	return nil
}
