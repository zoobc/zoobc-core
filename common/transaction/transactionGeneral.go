// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package transaction

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/zoobc/zoobc-core/common/crypto"

	"github.com/zoobc/zoobc-core/common/accounttype"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

type (
	UtilInterface interface {
		GetTransactionBytes(transaction *model.Transaction, sign bool) ([]byte, error)
		ParseTransactionBytes(transactionBytes []byte, sign bool) (*model.Transaction, error)
		GetTransactionID(transactionHash []byte) (int64, error)
		ValidateTransaction(tx *model.Transaction, typeAction TypeAction, verifySignature bool) error
		GenerateMultiSigAddress(info *model.MultiSignatureInfo) (hash []byte, address []byte, err error)
	}

	Util struct {
		FeeScaleService     fee.FeeScaleServiceInterface
		MempoolCacheStorage storage.CacheStorageInterface
		QueryExecutor       query.ExecutorInterface
		AccountDatasetQuery query.AccountDatasetQueryInterface
	}

	MultisigTransactionUtilInterface interface {
		CheckMultisigComplete(
			transactionUtil UtilInterface,
			multisignatureInfoHelper MultisignatureInfoHelperInterface,
			signatureInfoHelper SignatureInfoHelperInterface,
			pendingTransactionHelper PendingTransactionHelperInterface,
			body *model.MultiSignatureTransactionBody,
			txHeight uint32,
			multiSignaturesInfo *[]*model.MultiSignatureInfo,
			pendingSignatures *[]*model.PendingSignature,
			pendingTransaction *model.PendingTransaction,
		) error
		ValidatePendingTransactionBytes(
			transactionUtil UtilInterface,
			typeSwitcher TypeActionSwitcher,
			multisigInfoHelper MultisignatureInfoHelperInterface,
			pendingTransactionHelper PendingTransactionHelperInterface,
			multisigInfo *model.MultiSignatureInfo,
			senderAddress, unsignedTxBytes []byte,
			blockHeight uint32,
			dbTx bool,
		) error
		ValidateMultisignatureInfo(info *model.MultiSignatureInfo) error
		ValidateSignatureInfo(
			signature crypto.SignatureInterface, signatureInfo *model.SignatureInfo, multisignatureAddresses map[string]bool,
		) error
		ParseSignatureInfoBytesAsCandidates(
			txHash, multiSignatureAddress, key, value []byte,
			txHeight uint32,
			multiSignaturesInfo *[]*model.MultiSignatureInfo,
			pendingSignatures *[]*model.PendingSignature,
		) (err error)
	}
	MultisigTransactionUtil struct {
	}
)

// GetTransactionBytes translate transaction model to its byte representation
// provide sign = true to translate transaction with its signature, sign = false
// for without signature (used for verify signature)
func (*Util) GetTransactionBytes(transaction *model.Transaction, signed bool) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(transaction.TransactionType))
	buffer.Write(util.ConvertUint32ToBytes(transaction.Version)[:constant.TransactionVersion])
	buffer.Write(util.ConvertUint64ToBytes(uint64(transaction.Timestamp)))

	// Address format (byte array): [account type][address public key]
	buffer.Write(transaction.SenderAccountAddress)

	// Address format (byte array): [account type][address public key]
	if transaction.GetRecipientAccountAddress() == nil || bytes.Equal(transaction.GetRecipientAccountAddress(), []byte{}) {
		emptyAccType, err := accounttype.NewAccountType(int32(model.AccountType_EmptyAccountType), make([]byte, 0))
		if err != nil {
			return nil, err
		}
		emptyAccAddress, err := emptyAccType.GetAccountAddress()
		if err != nil {
			return nil, err
		}
		buffer.Write(emptyAccAddress)
	} else {
		buffer.Write(transaction.RecipientAccountAddress)
	}
	buffer.Write(util.ConvertUint64ToBytes(uint64(transaction.Fee)))
	// transaction body length
	buffer.Write(util.ConvertUint32ToBytes(transaction.TransactionBodyLength))
	buffer.Write(transaction.TransactionBodyBytes)
	/***
	Escrow part
	1. ApproverAddress
	2. Commission
	3. Timeout
	4. Instruction
	*/
	if transaction.GetEscrow() != nil && transaction.GetEscrow().GetApproverAddress() != nil {
		// Address format (byte array): [account type][address public key]
		buffer.Write(transaction.GetEscrow().GetApproverAddress())

		buffer.Write(util.ConvertUint64ToBytes(uint64(transaction.GetEscrow().GetCommission())))
		buffer.Write(util.ConvertUint64ToBytes(transaction.GetEscrow().GetTimeout()))

		buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(transaction.GetEscrow().GetInstruction())))))
		buffer.Write([]byte(transaction.GetEscrow().GetInstruction()))
	} else {
		// if no escrow, write an empty account for approver
		emptyAccType, err := accounttype.NewAccountType(int32(model.AccountType_EmptyAccountType), []byte{})
		if err != nil {
			return nil, err
		}
		emptyAccAddr, err := emptyAccType.GetAccountAddress()
		if err != nil {
			return nil, err
		}
		buffer.Write(emptyAccAddr)
	}

	// transaction message
	msgLength := len(transaction.GetMessage())
	buffer.Write(util.ConvertUint32ToBytes(uint32(msgLength)))
	if msgLength > 0 {
		buffer.Write(transaction.GetMessage())
	}

	if signed {
		if transaction.Signature == nil {
			return nil, errors.New("TransactionSignatureNotExist")
		}
		buffer.Write(transaction.Signature)
	}
	return buffer.Bytes(), nil
}

// ParseTransactionBytes build transaction from transaction bytes
func (u *Util) ParseTransactionBytes(transactionBytes []byte, sign bool) (*model.Transaction, error) {
	var (
		chunkedBytes  []byte
		mempoolObject storage.MempoolCacheObject
		transaction   model.Transaction
		buffer        *bytes.Buffer
		escrow        model.Escrow
		err           error
	)
	txHash := sha3.Sum256(transactionBytes)
	txID, err := u.GetTransactionID(txHash[:])
	if err != nil {
		return &transaction, err
	}
	err = u.MempoolCacheStorage.GetItem(txID, &mempoolObject)
	if err != nil {
		return nil, err
	}
	if mempoolObject.Tx.TransactionHash != nil {
		return &mempoolObject.Tx, nil
	}

	buffer = bytes.NewBuffer(transactionBytes)

	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.TransactionType))
	if err != nil {
		return nil, err
	}
	transaction.TransactionType = util.ConvertBytesToUint32(chunkedBytes)

	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.TransactionVersion))
	if err != nil {
		return nil, err
	}
	transaction.Version = uint32(chunkedBytes[0])

	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.Timestamp))
	if err != nil {
		return nil, err
	}
	transaction.Timestamp = int64(util.ConvertBytesToUint64(chunkedBytes))

	senderAccType, err := accounttype.ParseBytesToAccountType(buffer)
	if err != nil {
		return nil, err
	}
	senderAddress, err := senderAccType.GetAccountAddress()
	if err != nil {
		return nil, err
	}
	transaction.SenderAccountAddress = senderAddress

	recipientAccType, err := accounttype.ParseBytesToAccountType(buffer)
	if err != nil {
		return nil, err
	}
	if recipientAccType.GetTypeInt() != int32(model.AccountType_EmptyAccountType) {
		transaction.RecipientAccountAddress, err = recipientAccType.GetAccountAddress()
		if err != nil {
			return nil, err
		}
	}
	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.Fee))
	if err != nil {
		return nil, err
	}
	transaction.Fee = int64(util.ConvertBytesToUint64(chunkedBytes))

	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.TransactionBodyLength))
	if err != nil {
		return nil, err
	}
	transaction.TransactionBodyLength = util.ConvertBytesToUint32(chunkedBytes)
	transaction.TransactionBodyBytes, err = util.ReadTransactionBytes(buffer, int(transaction.TransactionBodyLength))
	if err != nil {
		return nil, err
	}
	/***
	Escrow part
	1. ApproverAddress
	2. Commission
	3. Timeout
	4. Instruction
	*/
	approverAccType, err := accounttype.ParseBytesToAccountType(buffer)
	if err != nil {
		return nil, err
	}
	// if approver account is empty (== empty account type), then skip the escrow part
	if approverAccType.GetTypeInt() != int32(model.AccountType_EmptyAccountType) {
		escrow.ApproverAddress, err = approverAccType.GetAccountAddress()
		if err != nil {
			return nil, err
		}

		chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.EscrowCommissionLength))
		if err != nil {
			return nil, err
		}
		escrow.Commission = int64(util.ConvertBytesToUint64(chunkedBytes))

		chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.EscrowTimeoutLength))
		if err != nil {
			return nil, err
		}
		escrow.Timeout = util.ConvertBytesToUint64(chunkedBytes)

		chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.EscrowInstructionLength))
		if err != nil {
			return nil, err
		}
		instructionLength := int(util.ConvertBytesToUint32(chunkedBytes))
		instruction, err := util.ReadTransactionBytes(buffer, instructionLength)
		if err != nil {
			return nil, err
		}
		escrow.Instruction = string(instruction)

		transaction.Escrow = &escrow
	}

	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.TxMessageBytesLength))
	if err != nil {
		return nil, err
	}
	messageLength := int(util.ConvertBytesToUint32(chunkedBytes))
	if messageLength > 0 {
		messageBytes, err := util.ReadTransactionBytes(buffer, messageLength)
		if err != nil {
			return nil, err
		}
		transaction.Message = messageBytes
	}

	if sign {
		signatureLength := senderAccType.GetSignatureLength()
		transaction.Signature, err = util.ReadTransactionBytes(buffer, int(signatureLength))
		if err != nil {
			return nil, blocker.NewBlocker(
				blocker.ParserErr,
				"no transaction signature",
			)
		}
	}
	// compute and return tx hash and ID too
	transactionHash := sha3.Sum256(transactionBytes)
	txID, _ = u.GetTransactionID(transactionHash[:])
	transaction.ID = txID
	transaction.TransactionHash = transactionHash[:]
	return &transaction, nil
}

// GetTransactionID calculate and returns a transaction ID given a transaction model
func (*Util) GetTransactionID(transactionHash []byte) (int64, error) {
	if len(transactionHash) == 0 {
		return -1, errors.New("InvalidTransactionHash")
	}
	ID := int64(util.ConvertBytesToUint64(transactionHash))
	return ID, nil
}

// ValidateTransaction take in transaction object and execute basic validation
func (u *Util) ValidateTransaction(tx *model.Transaction, typeAction TypeAction, verifySignature bool) error {
	var (
		err      error
		feeScale model.FeeScale
	)

	if tx.Fee <= 0 {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"TxFeeZero",
		)
	}
	if len(tx.Message) > constant.MaxMessageLength {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"TxMessageMaxLengthExceeded",
		)
	}
	err = u.FeeScaleService.GetLatestFeeScale(&feeScale)
	if err != nil {
		return err
	}
	minimumFee, err := typeAction.GetMinimumFee()
	if err != nil {
		return err
	}
	if tx.Fee < int64(math.Floor(float64(minimumFee)*(float64(feeScale.FeeScale)/float64(constant.OneZBC)))) {
		return blocker.NewBlocker(blocker.ValidationErr, fmt.Sprintf("MinimumFeeIs:%v", minimumFee*feeScale.FeeScale))
	}

	if tx.SenderAccountAddress == nil {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"TxSenderEmpty",
		)
	}
	// check if transaction is coming from future / comparison in second
	// There is additional time offset for the transaction timestamp before comparing with time now
	if time.Duration(tx.Timestamp)*time.Second-constant.TransactionTimeOffset > time.Duration(time.Now().UnixNano()) {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"TxComeFromFuture",
		)
	}

	// formally validate transaction body
	if len(tx.TransactionBodyBytes) != int(tx.TransactionBodyLength) {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"TxInvalidBodyFormat",
		)
	}

	// Checking the recipient has an model.AccountDatasetProperty_AccountDatasetEscrowApproval
	// when tx is not escrowed
	if tx.GetRecipientAccountAddress() != nil && (tx.Escrow != nil &&
		(tx.Escrow.GetApproverAddress() == nil || bytes.Equal(tx.Escrow.GetApproverAddress(), []byte{}))) {
		var (
			accountDataset model.AccountDataset
			row            *sql.Row
		)
		accDatasetQ, accDatasetArgs := u.AccountDatasetQuery.GetAccountDatasetEscrowApproval(tx.RecipientAccountAddress)
		row, _ = u.QueryExecutor.ExecuteSelectRow(accDatasetQ, false, accDatasetArgs...)
		err = u.AccountDatasetQuery.Scan(&accountDataset, row)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if accountDataset.GetIsActive() {
			return fmt.Errorf("RecipientRequireEscrow")
		}
	}

	unsignedTransactionBytes, err := u.GetTransactionBytes(tx, false)
	if err != nil {
		return err
	}
	// verify the signature of the transaction
	if verifySignature {
		txBytesHash := sha3.Sum256(unsignedTransactionBytes)
		err = crypto.NewSignature().VerifySignature(txBytesHash[:], tx.Signature, tx.SenderAccountAddress)
		if err != nil {
			return err
		}
	}

	return nil
}

// GenerateMultiSigAddress assembling MultiSignatureInfo to be an account address
// that is multi signature account address
func (u *Util) GenerateMultiSigAddress(info *model.MultiSignatureInfo) (hash, addr []byte, err error) {
	if info == nil {
		return hash, nil, fmt.Errorf("params cannot be nil")
	}
	util.SortByteArrays(info.Addresses)
	var (
		buff    = bytes.NewBuffer([]byte{})
		accType accounttype.AccountTypeInterface
	)
	buff.Write(util.ConvertUint32ToBytes(info.GetMinimumSignatures()))
	buff.Write(util.ConvertUint64ToBytes(uint64(info.GetNonce())))
	buff.Write(util.ConvertUint32ToBytes(uint32(len(info.GetAddresses()))))
	for _, add := range info.GetAddresses() {
		buff.Write(add)
	}
	hashed := sha3.Sum256(buff.Bytes())
	accType, err = accounttype.NewAccountType(int32(model.AccountType_MultiSignatureAccountType), hashed[:])
	if err != nil {
		return hash, nil, err
	}

	addr, err = accType.GetAccountAddress()
	return buff.Bytes(), addr, err

}

func NewMultisigTransactionUtil() *MultisigTransactionUtil {
	return &MultisigTransactionUtil{}
}

func (mtu *MultisigTransactionUtil) ValidateSignatureInfo(
	signature crypto.SignatureInterface,
	signatureInfo *model.SignatureInfo,
	multiSignatureInfoAddresses map[string]bool,
) error {
	// check for pending transaction first
	if signatureInfo.TransactionHash == nil { // transaction hash has to come with at least one signature
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"TransactionHashRequiredInSignatureInfo",
		)
	}
	if len(signatureInfo.Signatures) < 1 {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"MinimumOneSignatureRequiredInSignatureInfo",
		)
	}
	for addrHex, sig := range signatureInfo.Signatures {
		if sig == nil {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"SignatureMissing",
			)
		}
		if _, ok := multiSignatureInfoAddresses[addrHex]; !ok {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"SignerNotInParticipantList",
			)
		}
		decodedAcc, err := hex.DecodeString(addrHex)
		if err != nil {
			return err
		}
		err = signature.VerifySignature(signatureInfo.TransactionHash, sig, decodedAcc)
		if err != nil {
			signatureType := util.ConvertBytesToUint32(sig)
			if model.SignatureType(signatureType) != model.SignatureType_MultisigSignature {
				return err
			}
		}
	}
	return nil
}

func (*MultisigTransactionUtil) ValidateMultisignatureInfo(multisigInfo *model.MultiSignatureInfo) error {
	if len(multisigInfo.Addresses) < 2 {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"AtLeastTwoParticipantRequiredForMultisig",
		)
	}
	if multisigInfo.MinimumSignatures < 1 {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"AtLeastOneMinimumSignatures",
		)
	}
	return nil
}

func (mtu *MultisigTransactionUtil) ValidatePendingTransactionBytes(
	transactionUtil UtilInterface,
	typeSwitcher TypeActionSwitcher,
	multisigInfoHelper MultisignatureInfoHelperInterface,
	pendingTransactionHelper PendingTransactionHelperInterface,
	multisigInfo *model.MultiSignatureInfo,
	senderAddress, unsignedTxBytes []byte,
	blockHeight uint32,
	dbTx bool,
) error {
	var (
		pendingTx     model.PendingTransaction
		isParticipant = false
	)
	innerTx, err := transactionUtil.ParseTransactionBytes(unsignedTxBytes, false)
	if err != nil {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"FailToParseTransactionBytes",
		)
	}
	// check if multisignatureInfo has been submitted
	if len(multisigInfo.Addresses) == 0 {
		err = multisigInfoHelper.GetMultisigInfoByAddress(
			multisigInfo, innerTx.SenderAccountAddress, blockHeight,
		)
		if err != nil {
			return err
		}
	}
	// check if tx.Sender is participant in submitted multisignatureInfo
	for _, address := range multisigInfo.Addresses {
		if bytes.Equal(address, senderAddress) {
			isParticipant = true
		}
	}
	if !isParticipant {
		return blocker.NewBlocker(blocker.ValidationErr, "SenderNotParticipantOfMultisigAddress")
	}
	innerTa, err := typeSwitcher.GetTransactionType(innerTx)
	if err != nil {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"FailToCastInnerTransaction",
		)
	}
	err = transactionUtil.ValidateTransaction(innerTx, innerTa, false)
	if err != nil {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"FailToValidateInnerTx-GeneralValidation",
		)
	}

	err = innerTa.Validate(dbTx)
	if err != nil {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"FailToValidateInnerTx-TransactionTypeValidation",
		)
	}
	txHash := sha3.Sum256(unsignedTxBytes)
	err = pendingTransactionHelper.GetPendingTransactionByHash(
		&pendingTx,
		txHash[:],
		[]model.PendingTransactionStatus{
			model.PendingTransactionStatus_PendingTransactionExecuted,
			model.PendingTransactionStatus_PendingTransactionPending,
		},
		blockHeight,
		dbTx,
	)
	if err != sql.ErrNoRows {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"DuplicateOrPendingTransactionAlreadyExecuted",
		)
	}
	return nil
}

func (mtu *MultisigTransactionUtil) CheckMultisigComplete(
	transactionUtil UtilInterface,
	multisignatureInfoHelper MultisignatureInfoHelperInterface,
	signatureInfoHelper SignatureInfoHelperInterface,
	pendingTransactionHelper PendingTransactionHelperInterface,
	body *model.MultiSignatureTransactionBody,
	txHeight uint32,
	multiSignaturesInfo *[]*model.MultiSignatureInfo,
	pendingSignatures *[]*model.PendingSignature,
	pendingTransaction *model.PendingTransaction,
) (err error) {
	var (
		multiSignatureInfo       = body.GetMultiSignatureInfo()
		unsignedTransactionBytes = body.GetUnsignedTransactionBytes()
		signatureInfo            = body.GetSignatureInfo()
	)

	// Get MultiSignatureInfo and PendingTransaction
	if multiSignatureInfo == nil {
		if len(unsignedTransactionBytes) != 0 {
			var innerTX *model.Transaction
			innerTX, err = transactionUtil.ParseTransactionBytes(unsignedTransactionBytes, false)
			if err != nil {
				return err
			}
			err = multisignatureInfoHelper.GetMultisigInfoByAddress(multiSignatureInfo, innerTX.GetSenderAccountAddress(), txHeight)
			if err != nil {
				return err
			}
			if pendingTransaction == nil {
				txHash := sha3.Sum256(unsignedTransactionBytes)
				pendingTransaction = &model.PendingTransaction{
					SenderAddress:    nil,
					TransactionHash:  txHash[:],
					TransactionBytes: unsignedTransactionBytes,
					Status:           model.PendingTransactionStatus_PendingTransactionPending,
					BlockHeight:      txHeight,
					Latest:           true,
				}
			}
		} else if signatureInfo != nil {
			if pendingTransaction == nil {
				err = pendingTransactionHelper.GetPendingTransactionByHash(
					pendingTransaction,
					signatureInfo.GetTransactionHash(),
					[]model.PendingTransactionStatus{model.PendingTransactionStatus_PendingTransactionPending},
					txHeight,
					false,
				)
				if err != nil {
					return err
				}
			}
			if pendingTransaction != nil {
				err = multisignatureInfoHelper.GetMultisigInfoByAddress(
					multiSignatureInfo,
					pendingTransaction.GetSenderAddress(),
					txHeight,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	// Make sure parts is not nil
	if multiSignatureInfo == nil || pendingTransaction == nil {
		return blocker.NewBlocker(blocker.MultiSignatureNotComplete, "NotComplete")
	}

	// Check signatures already complete
	var pendingSigs []*model.PendingSignature
	pendingSigs, err = signatureInfoHelper.GetPendingSignatureByTransactionHash(pendingTransaction.GetTransactionHash(), txHeight)
	if err != nil {
		return err
	}

	m := multiSignatureInfo
	*multiSignaturesInfo = append(*multiSignaturesInfo, m)
	*pendingSignatures = append(*pendingSignatures, pendingSigs...)

	for _, mi := range *multiSignaturesInfo {
		var count uint32
		for _, ps := range *pendingSignatures {
			if bytes.Equal(ps.GetMultiSignatureAddress(), mi.GetMultisigAddress()) {
				count++
			}
			if count == mi.GetMinimumSignatures() {
				break
			}
		}
		if count < mi.GetMinimumSignatures() {
			return blocker.NewBlocker(blocker.MultiSignatureNotComplete, "NotComplete")
		}
	}

	return nil
}

func (mtu *MultisigTransactionUtil) ParseSignatureInfoBytesAsCandidates(
	txHash, multiSignatureAddress, key, value []byte,
	txHeight uint32,
	multiSignaturesInfo *[]*model.MultiSignatureInfo,
	pendingSignatures *[]*model.PendingSignature,
) (err error) {

	var (
		asMultiSigInfo, prev, valuePrev uint32
		accType                         accounttype.AccountTypeInterface
	)

	asMultiSigInfo = util.ConvertBytesToUint32(key[prev:][:constant.AccountAddressTypeLength])
	if model.AccountType(asMultiSigInfo) == model.AccountType_MultiSignatureAccountType {

		if len(*multiSignaturesInfo) == int(constant.MaxAllowedMultiSignatureTransactions) {
			return fmt.Errorf("MaxAllowedNestedMultiSignatureOffChainExceeded")
		}

		prev += constant.AccountAddressTypeLength
		var (
			multiSignatureInfo model.MultiSignatureInfo
			musigAddress       []byte
		)
		multiSigInfoLength := util.ConvertBytesToUint32(key[prev:][:constant.MultiSigInfoSize])
		prev += constant.MultiSigInfoSize

		err = mtu.ParseMultiSignatureInfoBytes(key[prev:][:multiSigInfoLength], txHeight, &multiSignatureInfo)
		if err != nil {
			return err
		}
		_, musigAddress, err = (&Util{}).GenerateMultiSigAddress(&multiSignatureInfo)
		if err != nil {
			return err
		}
		prev += multiSigInfoLength
		m := multiSignatureInfo
		*multiSignaturesInfo = append(*multiSignaturesInfo, &m)

		for _, participant := range multiSignatureInfo.GetAddresses() {
			nextValue := util.ConvertBytesToUint32(value[int(valuePrev):][:int(constant.MultiSignatureOffchainSignatureLength)])
			valuePrev += constant.MultiSignatureOffchainSignatureLength

			err = mtu.ParseSignatureInfoBytesAsCandidates(
				txHash,
				musigAddress,
				participant,
				value[valuePrev:][:nextValue],
				txHeight,
				multiSignaturesInfo,
				pendingSignatures,
			)
			if err != nil {
				return err
			}
			valuePrev += nextValue
		}
	} else {
		accType, err = accounttype.NewAccountType(int32(asMultiSigInfo), []byte{})
		if err != nil {
			return err
		}
		pendingSignature := &model.PendingSignature{
			AccountAddress:        key[prev:][:constant.AccountAddressTypeLength+accType.GetAccountPublicKeyLength()],
			MultiSignatureAddress: multiSignatureAddress,
			Signature:             value[valuePrev:][:accType.GetSignatureLength()],
			TransactionHash:       txHash,
			BlockHeight:           txHeight,
			Latest:                true,
		}
		p := pendingSignature
		*pendingSignatures = append(*pendingSignatures, p)

	}
	return nil
}

func (mtu *MultisigTransactionUtil) ParseMultiSignatureInfoBytes(
	multiSignatureInfoBytes []byte,
	txHeight uint32,
	multiSignatureInfo *model.MultiSignatureInfo,
) (err error) {
	var (
		buff      = bytes.NewBuffer(multiSignatureInfoBytes)
		addresses [][]byte
	)

	minSignatures := util.ConvertBytesToUint32(buff.Next(int(constant.MultiSigInfoMinSignature)))
	nonce := util.ConvertBytesToUint64(buff.Next(int(constant.MultiSigInfoNonce)))
	addressesLength := util.ConvertBytesToUint32(buff.Next(int(constant.MultiSigNumberOfAddress)))
	for i := 0; i < int(addressesLength); i++ {
		var (
			accType     accounttype.AccountTypeInterface
			address     []byte
			accTypeUint uint32
		)
		accTypeUint = util.ConvertBytesToUint32(buff.Next(int(constant.AccountAddressTypeLength)))
		if model.AccountType(accTypeUint) == model.AccountType_MultiSignatureAccountType {
			lenUint := util.ConvertBytesToUint32(buff.Next(int(constant.MultiSigAddressLength)))
			address = buff.Next(int(lenUint))
		} else {
			accType, err = accounttype.NewAccountType(int32(accTypeUint), []byte{})
			if err != nil {
				return err
			}
			address = append(util.ConvertUint32ToBytes(accTypeUint), buff.Next(int(accType.GetAccountPublicKeyLength()))...)
		}
		addresses = append(addresses, address)
	}

	*multiSignatureInfo = model.MultiSignatureInfo{
		MinimumSignatures: minSignatures,
		Nonce:             int64(nonce),
		Addresses:         addresses,
		BlockHeight:       txHeight,
	}
	return nil
}
