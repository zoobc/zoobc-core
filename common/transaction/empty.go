package transaction

import (
	"github.com/zoobc/zoobc-core/common/model"
)

type TXEmpty struct {
	Body *model.EmptyTransactionBody
}

func (tx *TXEmpty) ApplyConfirmed() error {
	return nil
}
func (tx *TXEmpty) ApplyUnconfirmed() error {
	return nil
}
func (tx *TXEmpty) Validate() error {
	return nil
}
