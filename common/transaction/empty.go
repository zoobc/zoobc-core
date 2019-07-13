package transaction

import (
	"github.com/zoobc/zoobc-core/common/model"
)

type TXEmpty struct {
	Body *model.EmptyTransactionBody
}

func (tx *TXEmpty) Apply() error {
	return nil
}
func (tx *TXEmpty) Unconfirmed() error {
	return nil
}
func (tx *TXEmpty) Validate() error {
	return nil
}
