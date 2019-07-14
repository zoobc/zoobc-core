package transaction

import (
	"encoding/binary"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	TypeAction interface {
		Apply() error
		Unconfirmed() error
		Validate() error
	}
)

func GetTransactionType(tx *model.Transaction) TypeAction {

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, tx.GetTransactionType())

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
			}
		default:
			return nil
		}
	default:
		return nil
	}
}
