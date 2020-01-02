package transaction

import (
	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	TypeAction interface {
		ApplyConfirmed() error
		ApplyUnconfirmed() error
		UndoApplyUnconfirmed() error
		// dbTx specify wether validation should read from transaction state or db state
		Validate(dbTx bool) error
		GetAmount() int64
		GetSize() uint32
		ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error)
		GetBodyBytes() []byte
		GetTransactionBody(transaction *model.Transaction)
		SkipMempoolTransaction(selectedTransactions []*model.Transaction) (bool, error)
	}
	TypeActionSwitcher interface {
		GetTransactionType(tx *model.Transaction) (TypeAction, error)
	}
	TypeSwitcher struct {
		Executor query.ExecutorInterface
	}
)

func (ts *TypeSwitcher) GetTransactionType(tx *model.Transaction) (TypeAction, error) {
	buf := util.ConvertUint32ToBytes(tx.GetTransactionType())
	switch buf[0] {
	case 0:
		switch buf[1] {
		case 0:
			return &TXEmpty{}, nil
		default:
			return nil, nil
		}
	case 1:
		switch buf[1] {
		case 0:
			sendMoneyBody, err := new(SendMoney).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &SendMoney{
				Body:                sendMoneyBody.(*model.SendMoneyTransactionBody),
				Fee:                 tx.Fee,
				SenderAddress:       tx.GetSenderAccountAddress(),
				RecipientAddress:    tx.GetRecipientAccountAddress(),
				Height:              tx.GetHeight(),
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				QueryExecutor:       ts.Executor,
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
			}, nil
		default:
			return nil, nil
		}
	case 2:
		switch buf[1] {
		case 0:
			nodeRegistrationBody, err := (&NodeRegistration{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			}).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &NodeRegistration{
				ID:                      tx.ID, // assign tx ID to nodeID
				Body:                    nodeRegistrationBody.(*model.NodeRegistrationTransactionBody),
				Fee:                     tx.Fee,
				SenderAddress:           tx.GetSenderAccountAddress(),
				Height:                  tx.GetHeight(),
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				AuthPoown:               &auth.ProofOfOwnershipValidation{},
				QueryExecutor:           ts.Executor,
				AccountLedgerQuery:      query.NewAccountLedgerQuery(),
			}, nil
		case 1:
			nodeRegistrationBody, err := (&UpdateNodeRegistration{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			}).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &UpdateNodeRegistration{
				Body:                  nodeRegistrationBody.(*model.UpdateNodeRegistrationTransactionBody),
				Fee:                   tx.Fee,
				SenderAddress:         tx.GetSenderAccountAddress(),
				Height:                tx.GetHeight(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &auth.ProofOfOwnershipValidation{},
				QueryExecutor:         ts.Executor,
				AccountLedgerQuery:    query.NewAccountLedgerQuery(),
			}, nil
		case 2:
			removeNodeRegistrationBody, err := new(RemoveNodeRegistration).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &RemoveNodeRegistration{
				Body:                  removeNodeRegistrationBody.(*model.RemoveNodeRegistrationTransactionBody),
				Fee:                   tx.Fee,
				SenderAddress:         tx.GetSenderAccountAddress(),
				Height:                tx.GetHeight(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         ts.Executor,
				AccountLedgerQuery:    query.NewAccountLedgerQuery(),
			}, nil
		case 3:
			claimNodeRegistrationBody, err := new(ClaimNodeRegistration).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &ClaimNodeRegistration{
				Body:                  claimNodeRegistrationBody.(*model.ClaimNodeRegistrationTransactionBody),
				Fee:                   tx.Fee,
				SenderAddress:         tx.GetSenderAccountAddress(),
				Height:                tx.GetHeight(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &auth.ProofOfOwnershipValidation{},
				QueryExecutor:         ts.Executor,
				AccountLedgerQuery:    query.NewAccountLedgerQuery(),
			}, nil
		default:
			return nil, nil
		}
	case 3:
		switch buf[1] {
		case 0:
			setupAccountDatasetTransactionBody, err := new(SetupAccountDataset).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &SetupAccountDataset{
				Body:                setupAccountDatasetTransactionBody.(*model.SetupAccountDatasetTransactionBody),
				Fee:                 tx.Fee,
				SenderAddress:       tx.GetSenderAccountAddress(),
				Height:              tx.GetHeight(),
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       ts.Executor,
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
			}, nil
		case 1:
			removeAccountDatasetTransactionBody, err := new(RemoveAccountDataset).ParseBodyBytes(tx.TransactionBodyBytes)
			if err != nil {
				return nil, err
			}
			return &RemoveAccountDataset{
				Body:                removeAccountDatasetTransactionBody.(*model.RemoveAccountDatasetTransactionBody),
				Fee:                 tx.Fee,
				SenderAddress:       tx.GetSenderAccountAddress(),
				Height:              tx.GetHeight(),
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       ts.Executor,
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
			}, nil
		default:
			return nil, nil
		}
	default:
		return nil, nil
	}
}
