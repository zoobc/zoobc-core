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
func (*TXEmpty) GetAmount() int64 {
	return 0
}
func (tx *TXEmpty) GetSize() uint32 {
	return 0
}
