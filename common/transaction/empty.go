package transaction

import (
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
)

type TXEmpty struct {
	Body *model.EmptyTransactionBody
}

func (tx *TXEmpty) Apply(chainType contract.ChainType) error {
	return nil
}
func (tx *TXEmpty) Unconfirmed(chainType contract.ChainType) error {
	return nil
}
func (tx *TXEmpty) Validate(chainType contract.ChainType) error {
	return nil
}
