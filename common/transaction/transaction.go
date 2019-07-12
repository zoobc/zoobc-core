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

func GetTransactionType(tx *model.Transaction) interface{} {

	var (
		t []byte
	)

	binary.LittleEndian.PutUint32(t, tx.GetTransactionType())

	switch t[0] {
	case 0:
		switch t[1] {
		case 0:
			return &TXEmpty{}
		default:
			return nil
		}
	case 1:
		switch t[1] {
		case 0:
			return &SendMoney{
				Body:               tx.GetSendMoneyTransactionBody(),
				SenderAccountID:    tx.GetSenderAccountID(),
				RecipientAccountID: tx.GetRecipientAccountID(),
				Height:             tx.GetHeight(),
			}
		default:
			return nil
		}
	}
	return nil
}
