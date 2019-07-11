package transaction

import (
	"errors"

	"github.com/zoobc/zoobc-core/common/model"
)

type SendMoney struct {
	Body               *model.SendMoneyTransactionBody
	SenderAccountID    []byte
	RecipientAccountID []byte
	Heigh              uint32
}

func (tx *SendMoney) Apply() error {
	return nil
}

func (tx *SendMoney) Unconfirmed() error {
	return nil
}
func (tx *SendMoney) Validate() error {
	if tx.Body.GetAmount() <= 0 {
		return errors.New("transaction must have an amount more than 0")
	}
	if tx.Heigh != 0 {
		if tx.RecipientAccountID == nil {
			return errors.New("transaction must have a valid recipient account id")
		}
		if tx.SenderAccountID == nil {
			return errors.New("transaction must hav a valid sender account id")
		}
	}

	return nil
}
