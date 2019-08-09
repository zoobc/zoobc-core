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
				Body:                 sendMoneyBody.(*model.SendMoneyTransactionBody),
				Fee:                  tx.Fee,
				SenderAddress:        tx.GetSenderAccountAddress(),
				SenderAccountType:    tx.GetSenderAccountType(),
				RecipientAddress:     tx.GetRecipientAccountAddress(),
				RecipientAccountType: tx.GetRecipientAccountType(),
				Height:               tx.GetHeight(),
				AccountQuery:         query.NewAccountQuery(),
				AccountBalanceQuery:  query.NewAccountBalanceQuery(),
				QueryExecutor:        ts.Executor,
			}
		default:
			return nil
		}
	case 2:
		switch buf[1] {
		case 0:
			nodeRegistrationBody := new(NodeRegistration).ParseBodyBytes(tx.TransactionBodyBytes)
			return &NodeRegistration{
				Body:                  nodeRegistrationBody.(*model.NodeRegistrationTransactionBody),
				Fee:                   tx.Fee,
				SenderAddress:         tx.GetSenderAccountAddress(),
				SenderAccountType:     tx.GetSenderAccountType(),
				Height:                tx.GetHeight(),
				AccountQuery:          query.NewAccountQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         ts.Executor,
			}
		default:
			return nil
		}
	default:
		return nil
	}
}
