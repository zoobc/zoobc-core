package transaction

import (
	"bytes"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// MultiSignatureTransaction represent wrapper transaction type that require multiple signer to approve the transcaction
	// wrapped
	MultiSignatureTransaction struct {
		Body      *model.MultiSignatureTransactionBody
		NormalFee fee.FeeModelInterface
	}
)

func (*MultiSignatureTransaction) ApplyConfirmed(blockTimestamp int64) error {
	return nil
}

func (*MultiSignatureTransaction) ApplyUnconfirmed() error {
	return nil
}

func (*MultiSignatureTransaction) UndoApplyUnconfirmed() error {
	return nil
}

// Validate dbTx specify whether validation should read from transaction state or db state
func (*MultiSignatureTransaction) Validate(dbTx bool) error {
	// make sure at least one of 3 optional body field is filled (multisig-info, tx-bytes, signature)
	// if signature exist, make sure tx-bytes exist
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

	multisigInfo := tx.Body.GetMultiSignatureInfo()
	multisigInfoSize += constant.MultiSigInfoMinSignature
	multisigInfoSize += constant.MultiSigInfoNonce
	for _, v := range multisigInfo.GetAddresses() {
		multisigInfoSize += uint32(len([]byte(v)))
	}

	txByteSize = uint32(len(tx.Body.GetUnsignedTransactionBytes()))
	for _, v := range tx.Body.GetSignatures() {
		signaturesSize += uint32(len(v))
	}
	return txByteSize + signaturesSize + multisigInfoSize
}

func (tx *MultiSignatureTransaction) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	var (
		addresses  []string
		signatures [][]byte
	)
	bufferBytes := bytes.NewBuffer(txBodyBytes)
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
	multisigInfo := &model.MultiSignatureInfo{
		MinimumSignatures: minSignatures,
		Nonce:             int64(nonce),
		Addresses:         addresses,
	}
	unsignedTxLength := util.ConvertUint32ToBytes(tx.Body.GetMultiSignatureInfo().GetMinimumSignatures())
	unsignedTx, err := util.ReadTransactionBytes(bufferBytes, int(unsignedTxLength))
	if err != nil {
		return nil, err
	}
	signaturesLength := util.ConvertBytesToUint32(bufferBytes.Next(int(constant.MultiSigNumberOfSignatures)))
	for i := 0; i < int(signaturesLength); i++ {
		signatureLength := util.ConvertBytesToUint32(bufferBytes.Next(int(constant.MultiSigSignatureLength)))
		signature, err := util.ReadTransactionBytes(bufferBytes, int(signatureLength))
		if err != nil {
			return nil, err
		}
		signatures = append(signatures, signature)
	}
	return &model.MultiSignatureTransactionBody{
		MultiSignatureInfo:       multisigInfo,
		UnsignedTransactionBytes: unsignedTx,
		Signatures:               signatures,
	}, nil
}

func (tx *MultiSignatureTransaction) GetBodyBytes() []byte {
	var (
		buffer = bytes.NewBuffer([]byte{})
	)
	buffer.Write(util.ConvertUint32ToBytes(tx.Body.GetMultiSignatureInfo().GetMinimumSignatures()))
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.GetMultiSignatureInfo().GetNonce())))
	buffer.Write(util.ConvertUint32ToBytes(uint32(len(tx.Body.GetMultiSignatureInfo().GetAddresses()))))
	for _, v := range tx.Body.GetMultiSignatureInfo().GetAddresses() {
		buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(v)))))
		buffer.Write([]byte(v))
	}
	buffer.Write(util.ConvertUint32ToBytes(uint32(len(tx.Body.GetUnsignedTransactionBytes()))))
	buffer.Write(tx.Body.GetUnsignedTransactionBytes())
	buffer.Write(util.ConvertUint32ToBytes(uint32(len(tx.Body.GetSignatures()))))
	for _, v := range tx.Body.GetSignatures() {
		buffer.Write(util.ConvertUint32ToBytes(uint32(len(v))))
		buffer.Write(v)
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
