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
			sendMoneyTxAmount := util.ConvertBytesToUint64(tx.GetTransactionBodyBytes())
			return &SendMoney{
				Body: &model.SendMoneyTransactionBody{
					Amount: int64(sendMoneyTxAmount),
				},
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
			return &NodeRegistration{
				Body:                tx.GetNodeRegistrationTransactionBody(),
				SenderAddress:       tx.GetSenderAccountAddress(),
				SenderAccountType:   tx.GetSenderAccountType(),
				Height:              tx.GetHeight(),
				AccountQuery:        query.NewAccountQuery(),
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				QueryExecutor:       ts.Executor,
			}
		default:
			return nil
		}
	default:
		return nil
	}
}
