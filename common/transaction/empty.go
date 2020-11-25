package transaction

import (
	"github.com/zoobc/zoobc-core/common/model"
)

type TXEmpty struct {
	Body *model.EmptyTransactionBody
}

func (tx *TXEmpty) Escrowable() (EscrowTypeAction, bool) {
	return nil, false
}

// SkipMempoolTransaction this tx type has no mempool filter
func (tx *TXEmpty) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	newBlockTimestamp int64,
	newBlockHeight uint32,
) (bool, error) {
	return false, nil
}

func (tx *TXEmpty) ApplyConfirmed(int64) error {
	return nil
}
func (tx *TXEmpty) ApplyUnconfirmed(applyInCache bool) error {
	return nil
}

func (tx *TXEmpty) UndoApplyUnconfirmed() error {
	return nil
}
func (tx *TXEmpty) Validate(bool) error {
	return nil
}

func (*TXEmpty) GetMinimumFee() (int64, error) {
	return 0, nil
}

func (*TXEmpty) GetAmount() int64 {
	return 0
}
func (tx *TXEmpty) GetSize() (uint32, error) {
	return 0, nil
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (*TXEmpty) ParseBodyBytes([]byte) (model.TransactionBodyInterface, error) {
	return &model.EmptyTransactionBody{}, nil
}

// GetBodyBytes translate tx body to bytes representation
func (*TXEmpty) GetBodyBytes() ([]byte, error) {
	return []byte{}, nil
}

func (*TXEmpty) GetTransactionBody(*model.Transaction) {}
