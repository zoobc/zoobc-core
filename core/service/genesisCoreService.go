package service

import (
	"errors"
	"log"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

var genesisFundReceiver = []map[string]int64{ // address : amount | public key hex
	{"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE": 10000000}, // 04264418e6f758dc777c33957fd652e048ef388bff51e5b84d505027fead1ca9
	{"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN": 10000000}, // 04266749faa93f9b6a15094c4d89037815455a76f254aeef2ebe4e445a538e0b
	{"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J": 10000000}, // 04264a2ef814619d4a2b1fa3b45f4aa09b248d53ef07d8e92237f3cc8eb30d6d
}

var genesisSignature = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
}

// GetGenesisTransactions return list of genesis transaction to be executed in the
// very beginning of running the blockchain
func GetGenesisTransactions(chainType contract.ChainType) []*model.Transaction {
	var genesisTxs []*model.Transaction
	switch chainType.(type) {
	case *chaintype.MainChain:
		for _, fundReceiver := range genesisFundReceiver {
			for receiver, amount := range fundReceiver {
				genesisTx := &model.Transaction{
					Version:                       1,
					TransactionType:               util.ConvertBytesToUint32([]byte{1, 0, 0, 0}),
					Height:                        0,
					Timestamp:                     1562806389280,
					SenderAccountAddressLength:    uint32(len([]byte(constant.GenesisAccountAddress))),
					SenderAccountAddress:          constant.GenesisAccountAddress,
					RecipientAccountAddressLength: uint32(len([]byte(receiver))),
					RecipientAccountAddress:       receiver,
					Fee:                           0,
					TransactionBodyLength:         8,
					TransactionBody: &model.Transaction_SendMoneyTransactionBody{
						SendMoneyTransactionBody: &model.SendMoneyTransactionBody{
							Amount: amount,
						},
					},
					TransactionBodyBytes: util.ConvertUint64ToBytes(uint64(amount)),
					Signature:            genesisSignature,
				}

				transactionBytes, err := util.GetTransactionBytes(genesisTx, true)
				if err != nil {
					//TODO: return error instead?
					log.Fatal(err)
				}
				transactionHash := sha3.Sum256(transactionBytes)
				genesisTx.TransactionHash = transactionHash[:]
				genesisTx.ID, _ = util.GetTransactionID(transactionHash[:])
				genesisTxs = append(genesisTxs, genesisTx)
			}
		}

		return genesisTxs
	default:
		return nil
	}
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
