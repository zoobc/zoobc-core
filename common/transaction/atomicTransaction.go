package transaction

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// AtomicTransaction field for AtomicTransactionInterface
	AtomicTransaction struct {
		ID                   int64
		Fee                  int64
		SenderAddress        []byte
		Height               uint32
		Body                 *model.AtomicTransactionBody
		Escrow               *model.Escrow
		QueryExecutor        query.ExecutorInterface
		TransactionQuery     query.TransactionQueryInterface
		TypeActionSwitcher   TypeActionSwitcher
		AccountBalanceHelper AccountBalanceHelperInterface
		EscrowFee            fee.FeeModelInterface
		NormalFee            fee.FeeModelInterface
		TransactionUtil      UtilInterface
		Signature            crypto.SignatureInterface
	}
)

func (tx *AtomicTransaction) ApplyConfirmed(blockTimestamp int64) (err error) {

	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.SenderAddress,
		-tx.Fee,
		model.EventType_EventAtomicTransaction,
		tx.Height,
		tx.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}

	for _, atomicInnerTX := range tx.Body.GetAtomicInnerTransactions() {
		var (
			innerTX    *model.Transaction
			typeAction TypeAction
		)

		innerTX, err = tx.TransactionUtil.ParseTransactionBytes(atomicInnerTX.GetAtomicInnerItem()[0], false)
		if err != nil {
			return err
		}
		typeAction, err = tx.TypeActionSwitcher.GetTransactionType(innerTX)
		if err != nil {
			return err
		}
		err = typeAction.ApplyConfirmed(blockTimestamp)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tx *AtomicTransaction) ApplyUnconfirmed() error {

	for _, atomicInnerTX := range tx.Body.GetAtomicInnerTransactions() {
		var (
			innerTX    *model.Transaction
			typeAction TypeAction
		)
		innerTX, err := tx.TransactionUtil.ParseTransactionBytes(atomicInnerTX.GetAtomicInnerItem()[0], false)
		if err != nil {
			return err
		}
		typeAction, err = tx.TypeActionSwitcher.GetTransactionType(innerTX)
		if err != nil {
			return err
		}
		err = typeAction.ApplyUnconfirmed()
		if err != nil {
			return err
		}
	}
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, -tx.Fee)
}

func (tx *AtomicTransaction) UndoApplyUnconfirmed() (err error) {

	for _, atomicInnerTX := range tx.Body.GetAtomicInnerTransactions() {
		var (
			innerTX    *model.Transaction
			typeAction TypeAction
		)

		innerTX, err = tx.TransactionUtil.ParseTransactionBytes(atomicInnerTX.GetAtomicInnerItem()[0], false)
		if err != nil {
			return err
		}
		typeAction, err = tx.TypeActionSwitcher.GetTransactionType(innerTX)
		if err != nil {
			return err
		}
		err = typeAction.UndoApplyUnconfirmed()
		if err != nil {
			return err
		}
	}
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, tx.Fee)
}

func (tx *AtomicTransaction) Validate(dbTx bool) (err error) {

	if len(tx.Body.GetAtomicInnerTransactions()) == 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "EmptyAtomicInnerTransaction")
	}

	for accountAddressHex, atomicInnerTX := range tx.Body.GetAtomicInnerTransactions() {
		var (
			innerTX        *model.Transaction
			typeAction     TypeAction
			accountAddress []byte
		)

		accountAddress, err = hex.DecodeString(accountAddressHex)
		if err != nil {
			return err
		}
		err = tx.Signature.VerifySignature(atomicInnerTX.GetAtomicInnerItem()[0], atomicInnerTX.GetAtomicInnerItem()[1], accountAddress)
		if err != nil {
			return err
		}
		innerTX, err = tx.TransactionUtil.ParseTransactionBytes(atomicInnerTX.GetAtomicInnerItem()[0], false)
		if err != nil {
			return err
		}
		typeAction, err = tx.TypeActionSwitcher.GetTransactionType(innerTX)
		if err != nil {
			return err
		}

		err = typeAction.Validate(dbTx)
		if err != nil {
			return err
		}

	}
	return nil
}

func (tx *AtomicTransaction) GetMinimumFee() (int64, error) {
	if tx.Escrow != nil && tx.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.Escrow.GetApproverAddress(), []byte{}) {
		return tx.EscrowFee.CalculateTxMinimumFee(tx.Body, tx.Escrow)
	}
	return tx.NormalFee.CalculateTxMinimumFee(tx.Body, tx.Escrow)
}

func (tx *AtomicTransaction) GetAmount() int64 {
	panic("implement me")
}

func (tx *AtomicTransaction) GetSize() (uint32, error) {
	panic("implement me")
}

func (tx *AtomicTransaction) ParseBodyBytes(txBodyBytes []byte) (body model.TransactionBodyInterface, err error) {
	var (
		buff        = bytes.NewBuffer(txBodyBytes)
		atomicInner = make(map[string]*model.AtomicInnerTransaction)
	)

	innerItemCount := util.ConvertBytesToUint32(buff.Next(4))
	if innerItemCount <= 0 {
		return nil, errors.New("empty inner tx")
	}

	for i := 0; i < int(innerItemCount); i++ {
		var (
			unsignedTX, signatureBytes []byte
		)
		unsignedTXLength := util.ConvertBytesToUint32(buff.Next(4))
		unsignedTX, err = util.ReadTransactionBytes(buff, int(unsignedTXLength))
		if err != nil {
			return nil, err
		}
		sigLength := util.ConvertBytesToUint32(buff.Next(4))
		signatureBytes, err = util.ReadTransactionBytes(buff, int(sigLength))
		if err != nil {
			return nil, err
		}

		atomicInner[hex.EncodeToString(unsignedTX)] = &model.AtomicInnerTransaction{
			AtomicInnerItem: [][]byte{
				unsignedTX,
				signatureBytes,
			},
		}
	}
	body = &model.AtomicTransactionBody{
		AtomicInnerTransactions: atomicInner,
	}
	return body, nil
}

func (tx *AtomicTransaction) GetBodyBytes() ([]byte, error) {
	var buff = bytes.NewBuffer([]byte{})

	if tx.Body.GetAtomicInnerTransactions() != nil {
		buff.Write(util.ConvertUint32ToBytes(uint32(len(tx.Body.GetAtomicInnerTransactions()))))
		for _, atomicInner := range tx.Body.GetAtomicInnerTransactions() {
			buff.Write(util.ConvertUint32ToBytes(uint32(len(atomicInner.GetAtomicInnerItem()[0]))))
			buff.Write(atomicInner.GetAtomicInnerItem()[0]) // unsigned bytes
			buff.Write(util.ConvertUint32ToBytes(uint32(len(atomicInner.GetAtomicInnerItem()))))
			buff.Write(atomicInner.GetAtomicInnerItem()[1]) // signature bytes
		}
		return buff.Bytes(), nil
	}
	return nil, errors.New("damn")
}

func (tx *AtomicTransaction) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_AtomicTransactionBody{
		AtomicTransactionBody: tx.Body,
	}
}

func (tx *AtomicTransaction) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	blockTimestamp int64,
	blockHeight uint32,
) (bool, error) {
	return false, nil
}

func (tx *AtomicTransaction) Escrowable() (EscrowTypeAction, bool) {
	if tx.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.Escrow.GetApproverAddress(), []byte{}) {
		tx.Escrow = &model.Escrow{
			ID:              tx.ID,
			SenderAddress:   tx.SenderAddress,
			ApproverAddress: tx.Escrow.GetApproverAddress(),
			Commission:      tx.Escrow.GetCommission(),
			Timeout:         tx.Escrow.GetTimeout(),
			Status:          tx.Escrow.GetStatus(),
			BlockHeight:     tx.Height,
			Latest:          true,
			Instruction:     tx.Escrow.GetInstruction(),
		}
		return EscrowTypeAction(tx), true
	}
	return nil, false
}

func (tx *AtomicTransaction) EscrowApplyConfirmed(blockTimestamp int64) error {
	panic("implement me")
}

func (tx *AtomicTransaction) EscrowApplyUnconfirmed() error {
	panic("implement me")
}

func (tx *AtomicTransaction) EscrowUndoApplyUnconfirmed() error {
	panic("implement me")
}

func (tx *AtomicTransaction) EscrowValidate(dbTx bool) error {
	panic("implement me")
}

func (tx *AtomicTransaction) EscrowApproval(blockTimestamp int64, txBody *model.ApprovalEscrowTransactionBody) error {
	panic("implement me")
}
