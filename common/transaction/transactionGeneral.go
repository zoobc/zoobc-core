package transaction

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/zoobc/zoobc-core/common/accounttype"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
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
		GenerateMultiSigAddress(info *model.MultiSignatureInfo) ([]byte, error)
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
			tx *model.MultiSignatureTransactionBody, txHeight uint32,
		) ([]*model.MultiSignatureTransactionBody, error)
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

	if signed {
		if transaction.Signature == nil {
			return nil, errors.New("TransactionSignatureNotExist")
		}
		buffer.Write(util.ConvertUint32ToBytes(uint32(len(transaction.Signature))))
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
	recipientAddress, err := recipientAccType.GetAccountAddress()
	if err != nil {
		return nil, err
	}
	transaction.RecipientAccountAddress = recipientAddress

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
	approverAddress, err := approverAccType.GetAccountAddress()
	if err != nil {
		return nil, err
	}

	// if approver account is empty (== empty account type), then skip the escrow part
	if approverAccType.GetTypeInt() != int32(model.AccountType_EmptyAccountType) {
		escrow.ApproverAddress = approverAddress

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

	if sign {
		var signatureLengthBytes, err = util.ReadTransactionBytes(buffer, int(constant.TransactionSignatureLength))
		if err != nil {
			return nil, err
		}
		signatureLength := util.ConvertBytesToUint32(signatureLengthBytes)
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
func (u *Util) GenerateMultiSigAddress(info *model.MultiSignatureInfo) ([]byte, error) {
	if info == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	util.SortByteArrays(info.Addresses)
	var (
		buff = bytes.NewBuffer([]byte{})
	)
	buff.Write(util.ConvertUint32ToBytes(info.GetMinimumSignatures()))
	buff.Write(util.ConvertIntToBytes(int(info.GetNonce())))
	buff.Write(util.ConvertUint32ToBytes(uint32(len(info.GetAddresses()))))
	for _, address := range info.GetAddresses() {
		//STEF we don't need to add the address length because we can derive it from addressType (first 4 bytes of accountAddress)
		buff.Write(address)
	}
	hashed := sha3.Sum256(buff.Bytes())
	return hashed[:], nil

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
	body *model.MultiSignatureTransactionBody, txHeight uint32,
) ([]*model.MultiSignatureTransactionBody, error) {
	if body.MultiSignatureInfo != nil {
		var (
			pendingTxs   []*model.PendingTransaction
			dbPendingTxs []*model.PendingTransaction
		)
		multisigAddress := body.MultiSignatureInfo.MultisigAddress
		if len(body.UnsignedTransactionBytes) > 0 {
			txHash := sha3.Sum256(body.UnsignedTransactionBytes)
			pendingTxs = append(pendingTxs, &model.PendingTransaction{
				TransactionHash:  txHash[:],
				TransactionBytes: body.UnsignedTransactionBytes,
				Status:           model.PendingTransactionStatus_PendingTransactionPending,
				BlockHeight:      txHeight,
			})
		}
		dbPendingTxs, err := pendingTransactionHelper.GetPendingTransactionBySenderAddress(multisigAddress, txHeight)
		if err != nil {
			return nil, err
		}

		pendingTxs = append(pendingTxs, dbPendingTxs...)
		if len(pendingTxs) < 1 {
			return nil, nil
		}
		var readyTxs []*model.MultiSignatureTransactionBody
		for _, v := range pendingTxs {
			var (
				sigInfo               *model.SignatureInfo
				pendingSigs           []*model.PendingSignature
				signatures            = make(map[string][]byte)
				validSignatureCounter uint32
			)
			pendingSigs, err := signatureInfoHelper.GetPendingSignatureByTransactionHash(v.TransactionHash, txHeight)
			if err != nil {
				return nil, err
			}

			for _, sig := range pendingSigs {
				signatures[hex.EncodeToString(sig.AccountAddress)] = sig.Signature
			}
			if body.SignatureInfo != nil {
				if bytes.Equal(v.TransactionHash, body.SignatureInfo.TransactionHash) {
					for addr, sig := range body.SignatureInfo.Signatures {
						signatures[addr] = sig
					}
				}
			}
			if len(signatures) < 1 {
				continue
			}
			sigInfo = &model.SignatureInfo{
				TransactionHash: v.TransactionHash,
				Signatures:      signatures,
			}
			for _, addr := range body.MultiSignatureInfo.Addresses {
				if sigInfo.Signatures[hex.EncodeToString(addr)] != nil {
					validSignatureCounter++
				}
			}
			if validSignatureCounter >= body.MultiSignatureInfo.MinimumSignatures {
				// todo: return ready to applyConfirm tx
				cpTx := &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       body.MultiSignatureInfo,
					UnsignedTransactionBytes: v.TransactionBytes,
					SignatureInfo:            sigInfo,
				}
				readyTxs = append(readyTxs, cpTx)
			}
		}
		return readyTxs, nil
	} else if len(body.UnsignedTransactionBytes) > 0 {
		var (
			multisigInfo          model.MultiSignatureInfo
			pendingSigs           []*model.PendingSignature
			validSignatureCounter uint32
			err                   error
		)
		txHash := sha3.Sum256(body.UnsignedTransactionBytes)
		innerTx, err := transactionUtil.ParseTransactionBytes(body.UnsignedTransactionBytes, false)
		if err != nil {
			return nil, blocker.NewBlocker(
				blocker.ValidationErr,
				"FailToParseTransactionBytes",
			)
		}
		err = multisignatureInfoHelper.GetMultisigInfoByAddress(
			&multisigInfo,
			innerTx.SenderAccountAddress,
			txHeight,
		)
		if err != nil {
			if err == sql.ErrNoRows { // multisig info not present
				return nil, nil
			}
			// other database errors
			return nil, err
		}
		body.MultiSignatureInfo = &multisigInfo
		if body.SignatureInfo != nil {
			for addrHex, sig := range body.SignatureInfo.Signatures {
				decodedAddr, err := hex.DecodeString(addrHex)
				if err != nil {
					return nil, blocker.NewBlocker(
						blocker.AppErr,
						"InvalidAccountAddress",
					)
				}
				pendingSigs = append(pendingSigs, &model.PendingSignature{
					TransactionHash: body.SignatureInfo.TransactionHash,
					AccountAddress:  decodedAddr,
					Signature:       sig,
					BlockHeight:     txHeight,
				})
			}
		}
		var dbPendingSigs []*model.PendingSignature
		dbPendingSigs, err = signatureInfoHelper.GetPendingSignatureByTransactionHash(txHash[:], txHeight)
		if err != nil {
			return nil, err
		}

		pendingSigs = append(pendingSigs, dbPendingSigs...)
		body.SignatureInfo = &model.SignatureInfo{
			TransactionHash: txHash[:],
			Signatures:      make(map[string][]byte),
		}
		for _, sig := range pendingSigs {
			body.SignatureInfo.Signatures[hex.EncodeToString(sig.AccountAddress)] = sig.Signature
		}
		if len(body.SignatureInfo.Signatures) < 1 {
			return nil, nil
		}

		for _, addr := range multisigInfo.Addresses {
			if body.SignatureInfo.Signatures[hex.EncodeToString(addr)] != nil {
				validSignatureCounter++
			}
		}
		if validSignatureCounter >= multisigInfo.MinimumSignatures {
			return []*model.MultiSignatureTransactionBody{
				body,
			}, nil
		}
	} else if body.SignatureInfo != nil {
		var (
			pendingTx             model.PendingTransaction
			pendingSigs           []*model.PendingSignature
			multisigInfo          model.MultiSignatureInfo
			validSignatureCounter uint32
			err                   error
		)
		txHash := body.SignatureInfo.TransactionHash

		err = pendingTransactionHelper.GetPendingTransactionByHash(
			&pendingTx,
			txHash,
			[]model.PendingTransactionStatus{model.PendingTransactionStatus_PendingTransactionPending},
			txHeight,
			true,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		body.UnsignedTransactionBytes = pendingTx.TransactionBytes
		innerTx, err := transactionUtil.ParseTransactionBytes(body.UnsignedTransactionBytes, false)
		if err != nil {
			return nil, blocker.NewBlocker(
				blocker.ValidationErr,
				"FailToParseTransactionBytes",
			)
		}
		pendingSigs, err = signatureInfoHelper.GetPendingSignatureByTransactionHash(txHash, txHeight)
		if err != nil {
			return nil, err
		}
		for _, sig := range pendingSigs {
			body.SignatureInfo.Signatures[hex.EncodeToString(sig.AccountAddress)] = sig.Signature
		}
		err = multisignatureInfoHelper.GetMultisigInfoByAddress(
			&multisigInfo,
			innerTx.SenderAccountAddress,
			txHeight,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		// validate signature
		for _, addr := range multisigInfo.Addresses {
			if body.SignatureInfo.Signatures[hex.EncodeToString(addr)] != nil {
				validSignatureCounter++
			}
		}
		if validSignatureCounter >= multisigInfo.MinimumSignatures {
			cpTx := body
			cpTx.UnsignedTransactionBytes = pendingTx.TransactionBytes
			cpTx.MultiSignatureInfo = &multisigInfo
			return []*model.MultiSignatureTransactionBody{
				cpTx,
			}, nil
		}

	}
	return nil, nil
}
