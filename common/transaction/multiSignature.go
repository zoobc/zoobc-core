package transaction

import (
	"bytes"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/sha3"

	"github.com/zoobc/zoobc-core/common/query"

	"github.com/zoobc/zoobc-core/common/crypto"

	"github.com/zoobc/zoobc-core/common/blocker"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// MultiSignatureTransaction represent wrapper transaction type that require multiple signer to approve the transcaction
	// wrapped
	MultiSignatureTransaction struct {
		SenderAddress       string
		Fee                 int64
		QueryExecutor       query.ExecutorInterface
		AccountBalanceQuery query.AccountBalanceQueryInterface
		Body                *model.MultiSignatureTransactionBody
		NormalFee           fee.FeeModelInterface
		TransactionUtil     UtilInterface
		TypeSwitcher        TypeActionSwitcher
		Signature           crypto.SignatureInterface
		Height              uint32
		MultisigUtil        MultisigTransactionUtilInterface
		// pending services
		MultisignatureInfoQuery query.MultisignatureInfoQueryInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
		PendingSignatureQuery   query.PendingSignatureQueryInterface
		TransactionQuery        query.TransactionQueryInterface
	}
)

func (tx *MultiSignatureTransaction) ApplyConfirmed(blockTimestamp int64) error {
	var (
		err error
	)
	// if have multisig info, MultisigInfoService.AddMultisigInfo() -> noop duplicate
	if tx.Body.MultiSignatureInfo != nil {
		address, err := tx.TransactionUtil.GenerateMultiSigAddress(tx.Body.MultiSignatureInfo)
		if err != nil {
			return err
		}
		tx.Body.MultiSignatureInfo.MultisigAddress = address
		tx.Body.MultiSignatureInfo.BlockHeight = tx.Height
		q, args := tx.MultisignatureInfoQuery.InsertMultisignatureInfo(tx.Body.MultiSignatureInfo)
		err = tx.QueryExecutor.ExecuteTransaction(q, args...)
		if err != nil {
			sqliteErr := err.(sqlite3.Error)
			if sqliteErr.ExtendedCode.Error() != sqlite3.ErrConstraintUnique.Error() {
				return blocker.NewBlocker(blocker.DBErr, err.Error())
			}
		}
	}
	// if have transaction bytes, PendingTransactionService.AddPendingTransaction() -> noop duplicate
	if len(tx.Body.UnsignedTransactionBytes) > 0 {
		innerTx, err := tx.TransactionUtil.ParseTransactionBytes(tx.Body.UnsignedTransactionBytes, false)
		if err != nil {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"FailToParseTransactionBytes",
			)
		}
		txHash := sha3.Sum256(tx.Body.UnsignedTransactionBytes)
		q, args := tx.PendingTransactionQuery.InsertPendingTransaction(&model.PendingTransaction{
			SenderAddress:    innerTx.SenderAccountAddress,
			TransactionHash:  txHash[:],
			TransactionBytes: tx.Body.UnsignedTransactionBytes,
			Status:           model.PendingTransactionStatus_PendingTransactionPending,
			BlockHeight:      tx.Height,
			Latest:           true,
		})
		err = tx.QueryExecutor.ExecuteTransaction(q, args...)
		if err != nil {
			sqliteErr := err.(sqlite3.Error)
			if sqliteErr.ExtendedCode.Error() != sqlite3.ErrConstraintUnique.Error() {
				return blocker.NewBlocker(blocker.DBErr, err.Error())
			}
		}
	}
	// if have signature, PendingSignature.AddPendingSignature -> noop duplicate
	if tx.Body.SignatureInfo != nil {
		for addr, sig := range tx.Body.SignatureInfo.Signatures {
			q, args := tx.PendingSignatureQuery.InsertPendingSignature(&model.PendingSignature{
				TransactionHash: tx.Body.SignatureInfo.TransactionHash,
				AccountAddress:  addr,
				Signature:       sig,
				BlockHeight:     tx.Height,
			})
			err = tx.QueryExecutor.ExecuteTransaction(q, args...)
			if err != nil {
				sqliteErr := err.(sqlite3.Error)
				if sqliteErr.Code == sqlite3.ErrConstraint &&
					sqliteErr.ExtendedCode.Error() == sqlite3.ErrConstraintUnique.Error() {
					continue
				}
				return blocker.NewBlocker(blocker.DBErr, err.Error())
			}

		}
	}
	// checks for completion, if musigInfo && txBytes && signatureInfo exist, check if signature info complete
	txs, err := tx.MultisigUtil.CheckMultisigComplete(tx.Body, tx.Height)
	if err != nil {
		return err
	}
	// every element in txs will have all three optional field filled, to avoid infinite recursive calls.
	for _, v := range txs {
		cpTx := tx
		cpTx.Body = v
		// parse the UnsignedTransactionBytes
		utx, err := tx.TransactionUtil.ParseTransactionBytes(cpTx.Body.UnsignedTransactionBytes, false)
		if err != nil {
			return err
		}
		utx.Height = tx.Height
		utxAct, err := tx.TypeSwitcher.GetTransactionType(utx)
		if err != nil {
			return err
		}
		err = utxAct.UndoApplyUnconfirmed()
		if err != nil {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"FailToApplyUndoUnconfirmedInnerTx",
			)
		}
		// call ApplyConfirmed() to inner transaction
		err = utxAct.ApplyConfirmed(blockTimestamp)
		if err != nil {
			return err
		}
		// update pending transaction status
		pendingTx := &model.PendingTransaction{
			SenderAddress:    v.MultiSignatureInfo.MultisigAddress,
			TransactionHash:  v.SignatureInfo.TransactionHash,
			TransactionBytes: v.UnsignedTransactionBytes,
			Status:           model.PendingTransactionStatus_PendingTransactionExecuted,
			BlockHeight:      tx.Height,
			Latest:           true,
		}
		updateQueries := tx.PendingTransactionQuery.UpdatePendingTransaction(pendingTx)
		err = tx.QueryExecutor.ExecuteTransactions(updateQueries)
		if err != nil {
			return err
		}

	}
	return nil
}

func (tx *MultiSignatureTransaction) ApplyUnconfirmed() error {
	var (
		err error
	)
	// reduce fee from sender
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-(tx.Fee),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
		},
	)
	err = tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	if err != nil {
		return err
	}
	// Run ApplyUnconfirmed of inner transaction
	if len(tx.Body.UnsignedTransactionBytes) > 0 {
		// parse and apply unconfirmed
		innerTx, err := tx.TransactionUtil.ParseTransactionBytes(tx.Body.UnsignedTransactionBytes, false)
		if err != nil {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"FailToParseTransactionBytes",
			)
		}
		innerTa, err := tx.TypeSwitcher.GetTransactionType(innerTx)
		if err != nil {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"FailToCastInnerTransaction",
			)
		}
		err = innerTa.ApplyUnconfirmed()
		if err != nil {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"FailToApplyUnconfirmedInnerTx",
			)
		}
	}
	return nil
}

func (tx *MultiSignatureTransaction) UndoApplyUnconfirmed() error {
	// recover fee
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		+(tx.Fee),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
		},
	)
	err := tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	if err != nil {
		return err
	}
	return nil
}

// Validate dbTx specify whether validation should read from transaction state or db state
func (tx *MultiSignatureTransaction) Validate(dbTx bool) error {
	body := tx.Body
	if body.MultiSignatureInfo == nil && body.SignatureInfo == nil && body.UnsignedTransactionBytes == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "AtLeastTxBytesSignatureInfoOrMultisignatureInfoMustBe"+
			"Provided")
	}
	if body.MultiSignatureInfo != nil {
		if len(body.MultiSignatureInfo.Addresses) < 2 {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"AtLeastTwoParticipantRequiredForMultisig",
			)
		}
		if body.MultiSignatureInfo.MinimumSignatures < 1 {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"AtLeastOneSignatureRequiredNeedToBeSet",
			)
		}
	}
	if len(body.UnsignedTransactionBytes) > 0 {
		innerTx, err := tx.TransactionUtil.ParseTransactionBytes(tx.Body.UnsignedTransactionBytes, false)
		if err != nil {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"FailToParseTransactionBytes",
			)
		}
		innerTa, err := tx.TypeSwitcher.GetTransactionType(innerTx)
		if err != nil {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"FailToCastInnerTransaction",
			)
		}
		err = innerTa.Validate(dbTx)
		if err != nil {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"FailToValidateInnerTa",
			)
		}

	}
	if body.SignatureInfo != nil {
		if body.SignatureInfo.TransactionHash == nil { // transaction hash has to come with at least one signature
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"TransactionHashRequiredInSignatureInfo",
			)
		}
		if len(body.SignatureInfo.Signatures) < 1 {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"MinimumOneSignatureRequiredInSignatureInfo",
			)
		}
		for addr, sig := range body.SignatureInfo.Signatures {
			if sig == nil {
				return blocker.NewBlocker(
					blocker.ValidationErr,
					"SignatureMissing",
				)
			}
			res := tx.Signature.VerifySignature(body.SignatureInfo.TransactionHash, sig, addr)
			if !res {
				return blocker.NewBlocker(
					blocker.ValidationErr,
					"InvalidSignature",
				)
			}
		}
	}
	return nil
}

func (tx *MultiSignatureTransaction) GetMinimumFee() (int64, error) {
	minFee, err := tx.NormalFee.CalculateTxMinimumFee(tx.Body, nil)
	if err != nil {
		return 0, err
	}
	return minFee, err
}

func (*MultiSignatureTransaction) GetAmount() int64 {
	return 0
}

func (tx *MultiSignatureTransaction) GetSize() uint32 {
	var (
		txByteSize, signaturesSize, multisigInfoSize uint32
	)
	// MultisigInfo
	multisigInfo := tx.Body.GetMultiSignatureInfo()
	multisigInfoSize += constant.MultisigFieldLength
	if multisigInfo != nil {
		multisigInfoSize += constant.MultiSigInfoMinSignature
		multisigInfoSize += constant.MultiSigInfoNonce
		multisigInfoSize += constant.MultiSigNumberOfAddress
		for _, v := range multisigInfo.GetAddresses() {
			multisigInfoSize += constant.MultiSigAddressLength
			multisigInfoSize += uint32(len([]byte(v)))
		}
	}
	// TransactionBytes
	txByteSize = constant.MultiSigUnsignedTxBytesLength + uint32(len(tx.Body.GetUnsignedTransactionBytes()))
	// SignatureInfo
	signaturesSize += constant.MultisigFieldLength
	if tx.Body.GetSignatureInfo() != nil {
		signaturesSize += constant.MultiSigTransactionHash
		signaturesSize += constant.MultiSigNumberOfSignatures
		for address, sig := range tx.Body.SignatureInfo.Signatures {
			signaturesSize += constant.MultiSigSignatureAddressLength
			signaturesSize += uint32(len([]byte(address)))
			signaturesSize += constant.MultiSigSignatureLength
			signaturesSize += uint32(len(sig))
		}
	}

	return txByteSize + signaturesSize + multisigInfoSize
}

func (tx *MultiSignatureTransaction) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	var (
		addresses     []string
		signatures    = make(map[string][]byte)
		multisigInfo  *model.MultiSignatureInfo
		signatureInfo *model.SignatureInfo
	)
	bufferBytes := bytes.NewBuffer(txBodyBytes)
	// MultisigInfo
	multisigInfoPresent := util.ConvertBytesToUint32(bufferBytes.Next(int(constant.MultisigFieldLength)))
	if multisigInfoPresent == constant.MultiSigFieldPresent {
		minSignatures := util.ConvertBytesToUint32(bufferBytes.Next(int(constant.MultiSigInfoMinSignature)))
		nonce := util.ConvertBytesToUint64(bufferBytes.Next(int(constant.MultiSigInfoNonce)))
		addressesLength := util.ConvertBytesToUint32(bufferBytes.Next(int(constant.MultiSigNumberOfAddress)))
		for i := 0; i < int(addressesLength); i++ {
			addressLength := util.ConvertBytesToUint32(bufferBytes.Next(int(constant.MultiSigAddressLength)))
			address, err := util.ReadTransactionBytes(bufferBytes, int(addressLength))
			if err != nil {
				return nil, err
			}
			addresses = append(addresses, string(address))
		}
		multisigInfo = &model.MultiSignatureInfo{
			MinimumSignatures: minSignatures,
			Nonce:             int64(nonce),
			Addresses:         addresses,
		}
	}
	// TransactionBytes
	unsignedTxLength := util.ConvertBytesToUint32(bufferBytes.Next(int(constant.MultiSigUnsignedTxBytesLength)))
	unsignedTx, err := util.ReadTransactionBytes(bufferBytes, int(unsignedTxLength))
	if err != nil {
		return nil, err
	}
	// SignatureInfo
	signatureInfoPresent := util.ConvertBytesToUint32(bufferBytes.Next(int(constant.MultisigFieldLength)))
	if signatureInfoPresent == constant.MultiSigFieldPresent {
		transactionHash, err := util.ReadTransactionBytes(bufferBytes, int(constant.MultiSigTransactionHash))
		if err != nil {
			return nil, err
		}
		signaturesLength := util.ConvertBytesToUint32(bufferBytes.Next(int(constant.MultiSigNumberOfSignatures)))
		for i := 0; i < int(signaturesLength); i++ {
			addressLength := util.ConvertBytesToUint32(bufferBytes.Next(int(constant.MultiSigAddressLength)))
			address, err := util.ReadTransactionBytes(bufferBytes, int(addressLength))
			if err != nil {
				return nil, err
			}
			signatureLength := util.ConvertBytesToUint32(bufferBytes.Next(int(constant.MultiSigSignatureLength)))
			signature, err := util.ReadTransactionBytes(bufferBytes, int(signatureLength))
			if err != nil {
				return nil, err
			}
			signatures[string(address)] = signature
		}
		signatureInfo = &model.SignatureInfo{
			TransactionHash: transactionHash,
			Signatures:      signatures,
		}
	}

	return &model.MultiSignatureTransactionBody{
		MultiSignatureInfo:       multisigInfo,
		UnsignedTransactionBytes: unsignedTx,
		SignatureInfo:            signatureInfo,
	}, nil
}

func (tx *MultiSignatureTransaction) GetBodyBytes() []byte {
	var (
		buffer = bytes.NewBuffer([]byte{})
	)
	// Multisig Info
	if tx.Body.GetMultiSignatureInfo() != nil {
		buffer.Write(util.ConvertUint32ToBytes(constant.MultiSigFieldPresent))
		buffer.Write(util.ConvertUint32ToBytes(tx.Body.GetMultiSignatureInfo().GetMinimumSignatures()))
		buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.GetMultiSignatureInfo().GetNonce())))
		buffer.Write(util.ConvertUint32ToBytes(uint32(len(tx.Body.GetMultiSignatureInfo().GetAddresses()))))
		for _, v := range tx.Body.GetMultiSignatureInfo().GetAddresses() {
			buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(v)))))
			buffer.Write([]byte(v))
		}
	} else {
		buffer.Write(util.ConvertUint32ToBytes(constant.MultiSigFieldMissing))
	}
	// Transaction Bytes
	buffer.Write(util.ConvertUint32ToBytes(uint32(len(tx.Body.GetUnsignedTransactionBytes()))))
	buffer.Write(tx.Body.GetUnsignedTransactionBytes())
	// SignatureInfo
	if tx.Body.GetSignatureInfo() != nil {
		buffer.Write(util.ConvertUint32ToBytes(constant.MultiSigFieldPresent))
		buffer.Write(tx.Body.GetSignatureInfo().GetTransactionHash())
		buffer.Write(util.ConvertUint32ToBytes(uint32(len(tx.Body.GetSignatureInfo().GetSignatures()))))
		for address, sig := range tx.Body.GetSignatureInfo().GetSignatures() {
			buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(address)))))
			buffer.Write([]byte(address))
			buffer.Write(util.ConvertUint32ToBytes(uint32(len(sig))))
			buffer.Write(sig)
		}
	} else {
		buffer.Write(util.ConvertUint32ToBytes(constant.MultiSigFieldMissing))
	}

	return buffer.Bytes()
}

func (tx *MultiSignatureTransaction) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_MultiSignatureTransactionBody{
		MultiSignatureTransactionBody: tx.Body,
	}
}

func (*MultiSignatureTransaction) SkipMempoolTransaction(selectedTransactions []*model.Transaction) (bool, error) {
	return false, nil
}

func (*MultiSignatureTransaction) Escrowable() (EscrowTypeAction, bool) {
	return nil, false
}
