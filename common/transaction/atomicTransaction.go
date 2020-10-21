package transaction

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"errors"

	"github.com/zoobc/zoobc-core/common/accounttype"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

type (
	// AtomicTransaction field for AtomicTransactionInterface
	AtomicTransaction struct {
		ID                     int64
		Fee                    int64
		SenderAddress          []byte
		Height                 uint32
		Body                   *model.AtomicTransactionBody
		AtomicTransactionQuery query.AtomicTransactionQueryInterface
		Escrow                 *model.Escrow
		EscrowQuery            query.EscrowTransactionQueryInterface
		QueryExecutor          query.ExecutorInterface
		TransactionQuery       query.TransactionQueryInterface
		TypeActionSwitcher     TypeActionSwitcher
		AccountBalanceHelper   AccountBalanceHelperInterface
		EscrowFee              fee.FeeModelInterface
		NormalFee              fee.FeeModelInterface
		TransactionUtil        UtilInterface
		Signature              crypto.SignatureInterface
	}
)

func (tx *AtomicTransaction) ApplyConfirmed(blockTimestamp int64) (err error) {

	var (
		atomics []*model.Atomic
		txs     []*model.Transaction
		queries [][]interface{}
	)

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

	for k, unsignedTX := range tx.Body.GetUnsignedTransactionBytes() {
		var (
			innerTX    *model.Transaction
			typeAction TypeAction
		)

		innerTX, err = tx.TransactionUtil.ParseTransactionBytes(unsignedTX, false)
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

		atomics = append(atomics, &model.Atomic{
			ID:                  innerTX.GetID(),
			TransactionID:       tx.ID,
			SenderAddress:       innerTX.GetSenderAccountAddress(),
			BlockHeight:         tx.Height,
			UnsignedTransaction: unsignedTX,
			Signature:           tx.Body.GetSignatures()[hex.EncodeToString(innerTX.GetSenderAccountAddress())],
			AtomicIndex:         uint32(k),
		})

		innerTX.ChildType = model.TransactionChildType_AtomicChild
		txs = append(txs, innerTX)
	}
	txQ, txArgs := tx.TransactionQuery.InsertTransactions(txs)
	queries = append(queries, append([]interface{}{txQ}, txArgs...))
	atomicQ, atomicArgs := tx.AtomicTransactionQuery.InsertAtomicTransactions(atomics)
	queries = append(queries, append([]interface{}{atomicQ}, atomicArgs...))
	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}
	return nil
}

func (tx *AtomicTransaction) ApplyUnconfirmed() error {

	for _, unsignedTX := range tx.Body.GetUnsignedTransactionBytes() {
		var (
			innerTX    *model.Transaction
			typeAction TypeAction
		)
		innerTX, err := tx.TransactionUtil.ParseTransactionBytes(unsignedTX, false)
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

	for _, unsignedTX := range tx.Body.GetUnsignedTransactionBytes() {
		var (
			innerTX    *model.Transaction
			typeAction TypeAction
		)

		innerTX, err = tx.TransactionUtil.ParseTransactionBytes(unsignedTX, false)
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

	if len(tx.Body.GetUnsignedTransactionBytes()) == 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "EmptyAtomicInnerTransaction")
	}
	if len(tx.Body.GetSignatures()) == 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "EmptySignatures")
	}

	var unsignedHash []byte
	for _, unsignedTX := range tx.Body.GetUnsignedTransactionBytes() {
		var (
			innerTX    *model.Transaction
			typeAction TypeAction
		)
		innerTX, err = tx.TransactionUtil.ParseTransactionBytes(unsignedTX, false)
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
		hashed := sha3.Sum256(unsignedTX)
		unsignedHash = append(unsignedHash, hashed[:]...)
	}
	for accountAddressHex, signature := range tx.Body.GetSignatures() {
		var accountAddress []byte
		accountAddress, err = hex.DecodeString(accountAddressHex)
		if err != nil {
			return err
		}
		err = tx.Signature.VerifySignature(unsignedHash, signature, accountAddress)
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
	return 0
}

func (tx *AtomicTransaction) GetSize() (uint32, error) {
	var size int

	size += len(tx.Body.GetUnsignedTransactionBytes())
	for _, unsignedTX := range tx.Body.GetUnsignedTransactionBytes() {
		size += 4 // unsignedTX length
		size += len(unsignedTX)
	}

	for addressHex, signature := range tx.Body.GetSignatures() {
		size += len(addressHex)
		size += len(signature)
	}

	return uint32(size), nil
}

func (tx *AtomicTransaction) ParseBodyBytes(txBodyBytes []byte) (body model.TransactionBodyInterface, err error) {
	var (
		buff        = bytes.NewBuffer(txBodyBytes)
		unsignedTXs = make([][]byte, 0)
		signatures  = make(map[string][]byte)
	)

	innerTXCount := util.ConvertBytesToUint32(buff.Next(4))
	if innerTXCount == 0 {
		return body, errors.New("empty inner tx")
	}

	for i := 0; i < int(innerTXCount); i++ {
		var (
			unsignedTX []byte
		)

		unsignedTXLength := util.ConvertBytesToUint32(buff.Next(4))
		unsignedTX, err = util.ReadTransactionBytes(buff, int(unsignedTXLength))
		if err != nil {
			return body, err
		}
		unsignedTXs = append(unsignedTXs, unsignedTX)
	}

	signatureCount := util.ConvertBytesToUint32(buff.Next(4))
	if signatureCount == 0 {
		return body, errors.New("empty signature")
	}
	for i := 0; i < int(signatureCount); i++ {
		var (
			signature      []byte
			accType        accounttype.AccountTypeInterface
			accountAddress []byte
		)
		accType, err = accounttype.ParseBytesToAccountType(buff)
		if err != nil {
			return body, err
		}
		signature, err = util.ReadTransactionBytes(buff, int(accType.GetSignatureLength()))
		if err != nil {
			return body, err
		}
		accountAddress, err = accType.GetAccountAddress()
		if err != nil {
			return body, err
		}
		signatures[hex.EncodeToString(accountAddress)] = signature
	}

	body = &model.AtomicTransactionBody{
		UnsignedTransactionBytes: unsignedTXs,
		Signatures:               signatures,
	}
	return body, nil
}

func (tx *AtomicTransaction) GetBodyBytes() (bodyBytes []byte, err error) {
	var (
		buff = bytes.NewBuffer([]byte{})
	)

	buff.Write(util.ConvertUint32ToBytes(uint32(len(tx.Body.GetUnsignedTransactionBytes())))) // total unsigned inner transactions
	for _, unsignedTX := range tx.Body.GetUnsignedTransactionBytes() {
		buff.Write(util.ConvertUint32ToBytes(uint32(len(unsignedTX))))
		buff.Write(unsignedTX)
	}
	buff.Write(util.ConvertUint32ToBytes(uint32(len(tx.Body.GetSignatures())))) // total signatures
	for addressHex, signature := range tx.Body.GetSignatures() {
		var addressBytes []byte
		addressBytes, err = hex.DecodeString(addressHex)
		if err != nil {
			return bodyBytes, err
		}
		buff.Write(addressBytes)
		buff.Write(signature)
	}

	return buff.Bytes(), nil
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
	var (
		err error
	)

	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.SenderAddress,
		-(tx.Fee + tx.Escrow.GetCommission()),
		model.EventType_EventEscrowedTransaction,
		tx.Height,
		tx.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}
	addEscrowQ := tx.EscrowQuery.InsertEscrowTransaction(tx.Escrow)
	err = tx.QueryExecutor.ExecuteTransactions(addEscrowQ)
	if err != nil {
		return err
	}
	return nil
}

func (tx *AtomicTransaction) EscrowApplyUnconfirmed() error {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, -(tx.Fee + tx.Escrow.GetCommission()))
}

func (tx *AtomicTransaction) EscrowUndoApplyUnconfirmed() error {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, tx.Fee+tx.Escrow.GetCommission())
}

func (tx *AtomicTransaction) EscrowValidate(dbTx bool) error {
	var (
		err    error
		enough bool
	)
	if tx.Escrow.GetCommission() <= 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "CommissionNotEnough")
	}
	if tx.Escrow.GetApproverAddress() == nil || bytes.Equal(tx.Escrow.GetApproverAddress(), []byte{}) {
		return blocker.NewBlocker(blocker.ValidationErr, "ApproverAddressRequired")
	}
	if tx.Escrow.GetRecipientAddress() == nil || bytes.Equal(tx.Escrow.GetRecipientAddress(), []byte{}) {
		return blocker.NewBlocker(blocker.ValidationErr, "RecipientAddressRequired")
	}
	if tx.Escrow.GetTimeout() > uint64(constant.MinRollbackBlocks) {
		return blocker.NewBlocker(blocker.ValidationErr, "TimeoutLimitExceeded")
	}

	err = tx.Validate(dbTx)
	if err != nil {
		return err
	}
	enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTx, tx.SenderAddress, tx.Fee+tx.Escrow.GetCommission())
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "AccountBalanceNotFound")
	}
	if !enough {
		return blocker.NewBlocker(blocker.ValidationErr, "AccountBalanceNotEnough")
	}
	return nil
}

func (tx *AtomicTransaction) EscrowApproval(blockTimestamp int64, txBody *model.ApprovalEscrowTransactionBody) (err error) {

	switch txBody.GetApproval() {
	case model.EscrowApproval_Approve:
		tx.Escrow.Status = model.EscrowStatus_Approved
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.SenderAddress,
			tx.Fee,
			model.EventType_EventEscrowedTransaction,
			tx.Height,
			tx.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
		err = tx.ApplyConfirmed(blockTimestamp)
		if err != nil {
			return err
		}
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.Escrow.GetApproverAddress(),
			tx.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.Height,
			tx.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	case model.EscrowApproval_Reject:
		tx.Escrow.Status = model.EscrowStatus_Rejected
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.Escrow.GetApproverAddress(),
			tx.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.Height,
			tx.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	default:
		tx.Escrow.Status = model.EscrowStatus_Expired
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.SenderAddress,
			tx.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.Height,
			tx.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	}
	addEscrowQ := tx.EscrowQuery.InsertEscrowTransaction(tx.Escrow)
	err = tx.QueryExecutor.ExecuteTransactions(addEscrowQ)
	if err != nil {
		return err
	}
	return nil
}
