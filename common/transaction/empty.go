package transaction

import (
	"github.com/zoobc/zoobc-core/common/model"
)

type TXEmpty struct {
	Body *model.EmptyTransactionBody
}

// FilterMempoolTransaction this tx type has no mempool filter
func (tx *TXEmpty) FilterMempoolTransaction(selectedTransactions []*model.Transaction) (bool, error) {
	return true, nil
}

func (tx *TXEmpty) ApplyConfirmed() error {
	return nil
}
func (tx *TXEmpty) ApplyUnconfirmed() error {
	return nil
}

func (tx *TXEmpty) UndoApplyUnconfirmed() error {
	return nil
}
func (tx *TXEmpty) Validate(bool) error {
	return nil
}
func (*TXEmpty) GetAmount() int64 {
	return 0
}
func (tx *TXEmpty) GetSize() uint32 {
	return 0
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (*TXEmpty) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	return &model.EmptyTransactionBody{}, nil
}

// GetBodyBytes translate tx body to bytes representation
func (*TXEmpty) GetBodyBytes() []byte {
	return []byte{}
}

func (*TXEmpty) GetTransactionBody(transaction *model.Transaction) {}
