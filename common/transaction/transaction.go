package transaction

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	TypeAction interface {
		ApplyConfirmed() error
		ApplyUnconfirmed() error
		UndoApplyUnconfirmed() error
		Validate() error
		GetAmount() int64
		GetSize() uint32
		ParseBodyBytes(txBodyBytes []byte) model.TransactionBodyInterface
		GetBodyBytes() []byte
	}
	TypeActionSwitcher interface {
		GetTransactionType(tx *model.Transaction) TypeAction
	}
	TypeSwitcher struct {
		Executor query.ExecutorInterface
	}
)

func (ts *TypeSwitcher) GetTransactionType(tx *model.Transaction) TypeAction {
	buf := util.ConvertUint32ToBytes(tx.GetTransactionType())
	switch buf[0] {
	case 0:
		switch buf[1] {
		case 0:
			return &TXEmpty{}
		default:
			return nil
		}
	case 1:
		switch buf[1] {
		case 0:
			sendMoneyBody := new(SendMoney).ParseBodyBytes(tx.TransactionBodyBytes)
			return &SendMoney{
				Body:                sendMoneyBody.(*model.SendMoneyTransactionBody),
				Fee:                 tx.Fee,
				SenderAddress:       tx.GetSenderAccountAddress(),
				RecipientAddress:    tx.GetRecipientAccountAddress(),
				Height:              tx.GetHeight(),
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				QueryExecutor:       ts.Executor,
			}
		default:
			return nil
		}
	case 2:
		switch buf[1] {
		case 0:
			nodeRegistrationBody := new(NodeRegistration).ParseBodyBytes(tx.TransactionBodyBytes)
			return &NodeRegistration{
				ID:                    tx.ID, // assign tx ID to nodeID
				Body:                  nodeRegistrationBody.(*model.NodeRegistrationTransactionBody),
				Fee:                   tx.Fee,
				SenderAddress:         tx.GetSenderAccountAddress(),
				Height:                tx.GetHeight(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         ts.Executor,
			}
		case 1:
			nodeRegistrationBody := new(UpdateNodeRegistration).ParseBodyBytes(tx.TransactionBodyBytes)
			return &UpdateNodeRegistration{
				Body:                  nodeRegistrationBody.(*model.UpdateNodeRegistrationTransactionBody),
				Fee:                   tx.Fee,
				SenderAddress:         tx.GetSenderAccountAddress(),
				Height:                tx.GetHeight(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         ts.Executor,
			}
		default:
			return nil
		}
	case 3:
		switch buf[1] {
		case 0:
			setupAccountDatasetTransactionBody := new(SetupAccountDataset).ParseBodyBytes(tx.TransactionBodyBytes)
			return &SetupAccountDataset{
				Body:                setupAccountDatasetTransactionBody.(*model.SetupAccountDatasetTransactionBody),
				Fee:                 tx.Fee,
				SenderAddress:       tx.GetSenderAccountAddress(),
				Height:              tx.GetHeight(),
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				QueryExecutor:       ts.Executor,
			}
		default:
			return nil
		}
	default:
		return nil
	}
}
