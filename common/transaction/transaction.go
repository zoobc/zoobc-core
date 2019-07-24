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
		Validate() error
		GetAmount() int64
		GetSize() uint32
	}
	TypeActionSwitcher interface {
		GetTransactionType(tx *model.Transaction) TypeAction
	}
	TypeSwitcher struct {
		Executor *query.Executor
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
			return &SendMoney{
				Body:                 tx.GetSendMoneyTransactionBody(),
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
	default:
		return nil
	}
}
