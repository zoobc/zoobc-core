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

	"github.com/zoobc/zoobc-core/common/accounttype"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// MultiSignatureTransaction represent wrapper transaction type that require multiple signer to approve the transaction
	// wrapped
	MultiSignatureTransaction struct {
		TransactionObject *model.Transaction
		Body              *model.MultiSignatureTransactionBody
		FeeScaleService   fee.FeeScaleServiceInterface
		TransactionUtil   UtilInterface
		TypeSwitcher      TypeActionSwitcher
		Signature         crypto.SignatureInterface
		EscrowQuery       query.EscrowTransactionQueryInterface
		// multisig helpers
		MultisigUtil             MultisigTransactionUtilInterface
		SignatureInfoHelper      SignatureInfoHelperInterface
		MultisignatureInfoHelper MultisignatureInfoHelperInterface
		PendingTransactionHelper PendingTransactionHelperInterface
		// general helpers
		AccountBalanceHelper AccountBalanceHelperInterface
		TransactionHelper    TransactionHelperInterface
		QueryExecutor        query.ExecutorInterface
	}
	// SignatureInfoHelperInterface multisignature helpers
	SignatureInfoHelperInterface interface {
		InsertPendingSignature(
			pendingSignature *model.PendingSignature,
		) error
		GetPendingSignatureByTransactionHash(
			transactionHash []byte, txHeight uint32,
		) ([]*model.PendingSignature, error)
	}

	MultisignatureInfoHelperInterface interface {
		GetMultisigInfoByAddress(
			multisigInfo *model.MultiSignatureInfo,
			multisigAddress []byte,
			blockHeight uint32,
		) error
		InsertMultisignatureInfo(
			multisigInfo *model.MultiSignatureInfo,
		) error
	}

	PendingTransactionHelperInterface interface {
		InsertPendingTransaction(
			pendingTransaction *model.PendingTransaction,
		) error
		GetPendingTransactionByHash(
			pendingTransaction *model.PendingTransaction,
			pendingTransactionHash []byte,
			pendingTransactionStatuses []model.PendingTransactionStatus,
			blockHeight uint32,
			dbTx bool,
		) error
		GetPendingTransactionBySenderAddress(
			senderAddress []byte, txHeight uint32,
		) ([]*model.PendingTransaction, error)
		ApplyUnconfirmedPendingTransaction(pendingTransactionBytes []byte) error
		UndoApplyUnconfirmedPendingTransaction(pendingTransactionBytes []byte) error
		ApplyConfirmedPendingTransaction(
			pendingTransaction []byte, txHeight uint32, blockTimestamp int64,
		) (*model.Transaction, error)
	}

	SignatureInfoHelper struct {
		PendingSignatureQuery   query.PendingSignatureQueryInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
		QueryExecutor           query.ExecutorInterface
		Signature               crypto.SignatureInterface
	}

	MultisignatureInfoHelper struct {
		MultisignatureInfoQuery        query.MultisignatureInfoQueryInterface
		MultiSignatureParticipantQuery query.MultiSignatureParticipantQueryInterface
		QueryExecutor                  query.ExecutorInterface
	}

	PendingTransactionHelper struct {
		MultisignatureInfoQuery query.MultisignatureInfoQueryInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
		TransactionUtil         UtilInterface
		TypeSwitcher            TypeActionSwitcher
		QueryExecutor           query.ExecutorInterface
	}
)

func (pth *PendingTransactionHelper) GetPendingTransactionByHash(
	pendingTx *model.PendingTransaction,
	pendingTransactionHash []byte,
	pendingTransactionStatuses []model.PendingTransactionStatus,
	blockHeight uint32,
	dbTx bool,
) error {
	var (
		err error
	)
	q, args := pth.PendingTransactionQuery.GetPendingTransactionByHash(
		pendingTransactionHash, pendingTransactionStatuses,
		blockHeight, constant.MinRollbackBlocks,
	)
	row, _ := pth.QueryExecutor.ExecuteSelectRow(q, dbTx, args...)
	err = pth.PendingTransactionQuery.Scan(pendingTx, row)
	if err != nil {
		return err
	}
	return nil
}

func (pth *PendingTransactionHelper) GetPendingTransactionBySenderAddress(
	senderAddress []byte, txHeight uint32,
) ([]*model.PendingTransaction, error) {
	var pendingTxs []*model.PendingTransaction
	q, args := pth.PendingTransactionQuery.GetPendingTransactionsBySenderAddress(
		senderAddress, model.PendingTransactionStatus_PendingTransactionPending,
		txHeight, constant.MinRollbackBlocks,
	)
	pendingTxRows, err := pth.QueryExecutor.ExecuteSelect(q, false, args...)
	if err != nil {
		return nil, err
	}
	defer pendingTxRows.Close()
	_, err = pth.PendingTransactionQuery.BuildModel(pendingTxs, pendingTxRows)
	if err != nil {
		return nil, err
	}
	return pendingTxs, nil
}

func (pth *PendingTransactionHelper) InsertPendingTransaction(
	pendingTransaction *model.PendingTransaction,
) error {
	insertPendingTxQry := pth.PendingTransactionQuery.InsertPendingTransaction(pendingTransaction)
	err := pth.QueryExecutor.ExecuteTransactions(insertPendingTxQry)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return nil
}

func (pth *PendingTransactionHelper) ApplyUnconfirmedPendingTransaction(
	pendingTransactionBytes []byte,
) error {
	// parse and apply unconfirmed
	innerTx, err := pth.TransactionUtil.ParseTransactionBytes(pendingTransactionBytes, false)
	if err != nil {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"FailToParseTransactionBytes",
		)
	}
	innerTa, err := pth.TypeSwitcher.GetTransactionType(innerTx)
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
	return nil
}

func (pth *PendingTransactionHelper) UndoApplyUnconfirmedPendingTransaction(pendingTransactionBytes []byte) error {
	// parse and apply unconfirmed
	innerTx, err := pth.TransactionUtil.ParseTransactionBytes(pendingTransactionBytes, false)
	if err != nil {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"FailToParseTransactionBytes",
		)
	}
	innerTa, err := pth.TypeSwitcher.GetTransactionType(innerTx)
	if err != nil {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"FailToCastInnerTransaction",
		)
	}
	err = innerTa.UndoApplyUnconfirmed()
	if err != nil {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"FailToUndoApplyUnconfirmedInnerTx",
		)
	}
	return nil
}

func (pth *PendingTransactionHelper) ApplyConfirmedPendingTransaction(
	pendingTransactionBytes []byte, txHeight uint32, blockTimestamp int64,
) (*model.Transaction, error) {
	utx, err := pth.TransactionUtil.ParseTransactionBytes(pendingTransactionBytes, false)
	if err != nil {
		return utx, err
	}
	utx.Height = txHeight
	utxAct, err := pth.TypeSwitcher.GetTransactionType(utx)
	if err != nil {
		return utx, err
	}
	err = utxAct.UndoApplyUnconfirmed()
	if err != nil {
		return utx, blocker.NewBlocker(
			blocker.ValidationErr,
			"FailToApplyUndoUnconfirmedInnerTx",
		)
	}
	// call ApplyConfirmed() to inner transaction
	err = utxAct.ApplyConfirmed(blockTimestamp)
	if err != nil {
		return utx, err
	}
	return utx, nil
}

func (sih *SignatureInfoHelper) InsertPendingSignature(
	pendingSignature *model.PendingSignature,
) error {
	insertPendingSigQ := sih.PendingSignatureQuery.InsertPendingSignature(pendingSignature)
	err := sih.QueryExecutor.ExecuteTransactions(insertPendingSigQ)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return nil
}

func (sih *SignatureInfoHelper) GetPendingSignatureByTransactionHash(transactionHash []byte, txHeight uint32) ([]*model.PendingSignature, error) {
	var pendingSigs []*model.PendingSignature
	q, args := sih.PendingSignatureQuery.GetPendingSignatureByHash(
		transactionHash,
		txHeight, constant.MinRollbackBlocks,
	)
	pendingSigRows, err := sih.QueryExecutor.ExecuteSelect(q, false, args...)
	if err != nil {
		return nil, err
	}
	defer pendingSigRows.Close()
	pendingSigs, err = sih.PendingSignatureQuery.BuildModel(pendingSigs, pendingSigRows)
	if err != nil {
		return nil, err
	}
	return pendingSigs, nil
}

func (msi *MultisignatureInfoHelper) GetMultisigInfoByAddress(
	multisigInfo *model.MultiSignatureInfo,
	multisigAddress []byte,
	blockHeight uint32,
) error {
	var (
		err              error
		multisigInfos    []*model.MultiSignatureInfo
		multisigAccounts [][]byte
	)
	q, args := msi.MultisignatureInfoQuery.GetMultisignatureInfoByAddressWithParticipants(
		multisigAddress, blockHeight, constant.MinRollbackBlocks,
	)
	rows, err := msi.QueryExecutor.ExecuteSelect(q, false, args...)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()
	multisigInfos, err = msi.MultisignatureInfoQuery.BuildModelWithParticipant(multisigInfos, rows)
	if err != nil {
		return err
	}

	if len(multisigInfos) == 0 {
		return blocker.NewBlocker(blocker.AppErr, "EmptyResultSet")
	}
	// make sure we have all data from db when returning
	multisigInfo.MultisigAddress = multisigInfos[0].GetMultisigAddress()
	multisigInfo.Latest = multisigInfos[0].Latest
	multisigInfo.BlockHeight = multisigInfos[0].BlockHeight
	multisigInfo.MinimumSignatures = multisigInfos[0].MinimumSignatures
	multisigInfo.Nonce = multisigInfos[0].Nonce
	for _, msInfo := range multisigInfos {
		if len(msInfo.Addresses[0]) > 0 {
			multisigAccounts = append(multisigAccounts, msInfo.Addresses[0])
		}
	}
	multisigInfo.Addresses = multisigAccounts
	return nil
}

func (msi *MultisignatureInfoHelper) InsertMultisignatureInfo(multisigInfo *model.MultiSignatureInfo) error {
	var (
		queries              = msi.MultisignatureInfoQuery.InsertMultisignatureInfo(multisigInfo)
		participantAddresses []*model.MultiSignatureParticipant
	)
	for k, participant := range multisigInfo.GetAddresses() {
		participantAddresses = append(participantAddresses, &model.MultiSignatureParticipant{
			MultiSignatureAddress: multisigInfo.GetMultisigAddress(),
			AccountAddressIndex:   uint32(k),
			AccountAddress:        participant,
			BlockHeight:           multisigInfo.GetBlockHeight(),
			Latest:                true,
		})
	}
	participantQ := msi.MultiSignatureParticipantQuery.InsertMultisignatureParticipants(participantAddresses)
	queries = append(queries, participantQ...)
	return msi.QueryExecutor.ExecuteTransactions(queries)
}

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
		tx.Body.MultiSignatureInfo.BlockHeight = tx.TransactionObject.Height
		tx.Body.MultiSignatureInfo.Latest = true
		err = tx.MultisignatureInfoHelper.InsertMultisignatureInfo(tx.Body.MultiSignatureInfo)
		if err != nil {
			return err
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
		var pendingTx model.PendingTransaction
		err = tx.PendingTransactionHelper.GetPendingTransactionByHash(&pendingTx, txHash[:], []model.PendingTransactionStatus{
			model.PendingTransactionStatus_PendingTransactionPending,
		}, tx.TransactionObject.Height, true)
		if err == sql.ErrNoRows {
			// apply-unconfirmed on pending transaction
			err = tx.PendingTransactionHelper.ApplyUnconfirmedPendingTransaction(tx.Body.UnsignedTransactionBytes)
			if err != nil {
				return err
			}
			// save the pending transaction
			pendingTx := &model.PendingTransaction{
				SenderAddress:    innerTx.SenderAccountAddress,
				TransactionHash:  txHash[:],
				TransactionBytes: tx.Body.UnsignedTransactionBytes,
				Status:           model.PendingTransactionStatus_PendingTransactionPending,
				BlockHeight:      tx.TransactionObject.Height,
				Latest:           true,
			}
			err = tx.PendingTransactionHelper.InsertPendingTransaction(pendingTx)
			if err != nil {
				return err
			}
		} else {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"PendingTransactionDuplicateOrAlreadyExecuted",
			)
		}
	}
	// if have signature, PendingSignature.AddPendingSignature -> noop duplicate
	if tx.Body.SignatureInfo != nil {
		for addrHex, sig := range tx.Body.SignatureInfo.Signatures {
			addr, err := hex.DecodeString(addrHex)
			if err != nil {
				return nil
			}
			pendingSig := &model.PendingSignature{
				TransactionHash: tx.Body.SignatureInfo.TransactionHash,
				AccountAddress:  addr,
				Signature:       sig,
				BlockHeight:     tx.TransactionObject.Height,
				Latest:          true,
			}
			err = tx.SignatureInfoHelper.InsertPendingSignature(pendingSig)
			if err != nil {
				return blocker.NewBlocker(blocker.DBErr, err.Error())
			}

		}
	}
	// checks for completion, if multisigInfo && txBytes && signatureInfo exist, check if signature info complete
	txs, err := tx.MultisigUtil.CheckMultisigComplete(
		tx.TransactionUtil,
		tx.MultisignatureInfoHelper,
		tx.SignatureInfoHelper,
		tx.PendingTransactionHelper,
		tx.Body,
		tx.TransactionObject.Height,
	)
	if err != nil {
		return err
	}
	// every element in txs will have all three optional field filled, to avoid infinite recursive calls.
	for _, v := range txs {
		cpTx := tx
		cpTx.Body = v
		// parse the UnsignedTransactionBytes
		utx, err := tx.PendingTransactionHelper.ApplyConfirmedPendingTransaction(
			cpTx.Body.UnsignedTransactionBytes,
			tx.TransactionObject.Height,
			blockTimestamp,
		)
		if err != nil {
			return err
		}

		// update pending transaction status
		pendingTx := &model.PendingTransaction{
			SenderAddress:    v.MultiSignatureInfo.MultisigAddress,
			TransactionHash:  v.SignatureInfo.TransactionHash,
			TransactionBytes: v.UnsignedTransactionBytes,
			Status:           model.PendingTransactionStatus_PendingTransactionExecuted,
			BlockHeight:      tx.TransactionObject.Height,
			Latest:           true,
		}
		// update pendingTx
		err = tx.PendingTransactionHelper.InsertPendingTransaction(pendingTx)
		if err != nil {
			return err
		}

		// save multisig_child transaction
		utx.MultisigChild = true
		utx.BlockID = tx.TransactionObject.BlockID
		err = tx.TransactionHelper.InsertTransaction(utx)
		if err != nil {
			return err
		}
	}
	// deduct fee from sender
	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.TransactionObject.SenderAccountAddress,
		-tx.TransactionObject.Fee,
		model.EventType_EventMultiSignatureTransaction,
		tx.TransactionObject.Height,
		tx.TransactionObject.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}
	return nil
}

func (tx *MultiSignatureTransaction) ApplyUnconfirmed() error {
	var (
		err error
	)
	// reduce fee from sender
	err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, -tx.TransactionObject.Fee)
	if err != nil {
		return err
	}
	// Run ApplyUnconfirmed of inner transaction
	if len(tx.Body.UnsignedTransactionBytes) > 0 {
		// parse and apply unconfirmed
		err = tx.PendingTransactionHelper.ApplyUnconfirmedPendingTransaction(tx.Body.UnsignedTransactionBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tx *MultiSignatureTransaction) UndoApplyUnconfirmed() error {
	// recover fee
	err := tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, tx.TransactionObject.Fee)
	if err != nil {
		return err
	}
	if len(tx.Body.UnsignedTransactionBytes) > 0 {
		err = tx.PendingTransactionHelper.UndoApplyUnconfirmedPendingTransaction(tx.Body.UnsignedTransactionBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

// Validate dbTx specify whether validation should read from transaction state or db state
func (tx *MultiSignatureTransaction) Validate(dbTx bool) error {
	var (
		body                  = tx.Body
		multisigInfoAddresses = make(map[string]bool)
		err                   error
		accountBalance        model.AccountBalance
	)
	if body.MultiSignatureInfo == nil && body.SignatureInfo == nil && body.UnsignedTransactionBytes == nil {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"AtLeastTxBytesSignatureInfoOrMultisignatureInfoMustBeProvided",
		)
	}

	// check existing & balance account sender
	err = tx.AccountBalanceHelper.GetBalanceByAccountAddress(&accountBalance, tx.TransactionObject.SenderAccountAddress, dbTx)
	if err != nil {
		return err
	}

	if accountBalance.SpendableBalance < tx.TransactionObject.Fee {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"UserBalanceNotEnough",
		)
	}

	if body.MultiSignatureInfo != nil {
		err := tx.MultisigUtil.ValidateMultisignatureInfo(body.MultiSignatureInfo)
		if err != nil {
			return err
		}
		for _, address := range body.MultiSignatureInfo.Addresses {
			multisigInfoAddresses[hex.EncodeToString(address)] = true
		}
		if len(body.UnsignedTransactionBytes) > 0 {
			err := tx.MultisigUtil.ValidatePendingTransactionBytes(
				tx.TransactionUtil,
				tx.TypeSwitcher,
				tx.MultisignatureInfoHelper,
				tx.PendingTransactionHelper,
				body.MultiSignatureInfo,
				tx.TransactionObject.SenderAccountAddress,
				body.UnsignedTransactionBytes,
				tx.TransactionObject.Height,
				dbTx,
			)
			if err != nil {
				return err
			}
		}
		if body.SignatureInfo != nil {
			err = tx.MultisigUtil.ValidateSignatureInfo(tx.Signature, body.SignatureInfo, multisigInfoAddresses)
			if err != nil {
				return err
			}
		}
	} else {
		var (
			err          error
			pendingTx    model.PendingTransaction
			multisigInfo model.MultiSignatureInfo
		)
		if len(body.UnsignedTransactionBytes) > 0 {
			err = tx.MultisigUtil.ValidatePendingTransactionBytes(
				tx.TransactionUtil,
				tx.TypeSwitcher,
				tx.MultisignatureInfoHelper,
				tx.PendingTransactionHelper,
				&multisigInfo,
				tx.TransactionObject.SenderAccountAddress,
				body.UnsignedTransactionBytes,
				tx.TransactionObject.Height,
				dbTx,
			)
			if err != nil {
				return err
			}
		}
		if body.SignatureInfo != nil {
			if len(multisigInfo.Addresses) == 0 {
				err = tx.PendingTransactionHelper.GetPendingTransactionByHash(
					&pendingTx,
					body.SignatureInfo.TransactionHash,
					[]model.PendingTransactionStatus{
						model.PendingTransactionStatus_PendingTransactionPending,
						model.PendingTransactionStatus_PendingTransactionExecuted,
					},
					tx.TransactionObject.Height,
					dbTx,
				)
				if err != nil {
					return err
				}
				if len(pendingTx.TransactionBytes) == 0 {
					return blocker.NewBlocker(blocker.ValidationErr, "NoPendingTransactionWithProvidedTransactionHash")
				}
				err = tx.MultisignatureInfoHelper.GetMultisigInfoByAddress(&multisigInfo, pendingTx.SenderAddress, tx.TransactionObject.Height)
				if err != nil {
					if err == sql.ErrNoRows {
						return blocker.NewBlocker(
							blocker.ValidationErr,
							"MultisignatureInfoHasNotBeenPosted",
						)
					}
					return err
				}
				for _, address := range multisigInfo.Addresses {
					multisigInfoAddresses[hex.EncodeToString(address)] = true
				}
			}
			err = tx.MultisigUtil.ValidateSignatureInfo(tx.Signature, body.SignatureInfo, multisigInfoAddresses)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (tx *MultiSignatureTransaction) GetMinimumFee() (int64, error) {
	var lastFeeScale model.FeeScale
	err := tx.FeeScaleService.GetLatestFeeScale(&lastFeeScale)
	if err != nil {
		return 0, err
	}
	return fee.CalculateTxMinimumFee(tx.TransactionObject, lastFeeScale.FeeScale)
}

func (*MultiSignatureTransaction) GetAmount() int64 {
	return 0
}

func (tx *MultiSignatureTransaction) GetSize() (uint32, error) {
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
			multisigInfoSize += uint32(len(v))
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

	return txByteSize + signaturesSize + multisigInfoSize, nil
}

func (tx *MultiSignatureTransaction) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	var (
		addresses     [][]byte
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
			accType, err := accounttype.ParseBytesToAccountType(bufferBytes)
			if err != nil {
				return nil, err
			}
			address, err := accType.GetAccountAddress()
			if err != nil {
				return nil, err
			}
			addresses = append(addresses, address)
		}
		// MultisigAddress is not provided because that data is not present in the body bytes.
		// It always matches with the sender of the multisig transaction, so external application can get that data from the sender data
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
			accType, err := accounttype.ParseBytesToAccountType(bufferBytes)
			if err != nil {
				return nil, err
			}
			address, err := accType.GetAccountAddress()
			if err != nil {
				return nil, err
			}
			signatureLength := util.ConvertBytesToUint32(bufferBytes.Next(int(constant.MultiSigSignatureLength)))
			signature, err := util.ReadTransactionBytes(bufferBytes, int(signatureLength))
			if err != nil {
				return nil, err
			}
			// encode address to hex string to be able to build the map (signature order must be preserved, so we can't use slices)
			signatures[hex.EncodeToString(address)] = signature
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

func (tx *MultiSignatureTransaction) GetBodyBytes() ([]byte, error) {
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
			buffer.Write(v)
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
		for addressHex, sig := range tx.Body.GetSignatureInfo().GetSignatures() {
			accountAddress, err := hex.DecodeString(addressHex)
			if err != nil {
				return nil, err
			}
			buffer.Write(accountAddress)
			buffer.Write(util.ConvertUint32ToBytes(uint32(len(sig))))
			buffer.Write(sig)
		}
	} else {
		buffer.Write(util.ConvertUint32ToBytes(constant.MultiSigFieldMissing))
	}

	return buffer.Bytes(), nil
}

func (tx *MultiSignatureTransaction) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_MultiSignatureTransactionBody{
		MultiSignatureTransactionBody: tx.Body,
	}
}

func (*MultiSignatureTransaction) SkipMempoolTransaction([]*model.Transaction, int64, uint32) (bool, error) {
	return false, nil
}

func (tx *MultiSignatureTransaction) Escrowable() (EscrowTypeAction, bool) {
	if tx.TransactionObject.Escrow != nil &&
		tx.TransactionObject.Escrow.GetApproverAddress() != nil &&
		!bytes.Equal(tx.TransactionObject.Escrow.GetApproverAddress(), []byte{}) {
		tx.TransactionObject.Escrow = util.PrepareEscrowObjectForAction(tx.TransactionObject)
		return EscrowTypeAction(tx), true
	}
	return nil, false
}

func (tx *MultiSignatureTransaction) EscrowApplyConfirmed(blockTimestamp int64) error {
	return tx.AccountBalanceHelper.AddAccountBalance(
		tx.TransactionObject.SenderAccountAddress,
		-(tx.TransactionObject.Fee + tx.TransactionObject.Escrow.GetCommission()),
		model.EventType_EventEscrowedTransaction,
		tx.TransactionObject.Height,
		tx.TransactionObject.ID,
		uint64(blockTimestamp),
	)
}

func (tx *MultiSignatureTransaction) EscrowApplyUnconfirmed() error {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, -tx.TransactionObject.Fee)
}

func (tx *MultiSignatureTransaction) EscrowUndoApplyUnconfirmed() error {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, tx.TransactionObject.Fee)
}

func (tx *MultiSignatureTransaction) EscrowValidate(dbTx bool) error {

	var (
		err    error
		enough bool
	)

	err = util.ValidateBasicEscrow(tx.TransactionObject)
	if err != nil {
		return err
	}

	err = tx.Validate(dbTx)
	if err != nil {
		return err
	}

	enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTx, tx.TransactionObject.SenderAccountAddress,
		tx.TransactionObject.Fee+tx.TransactionObject.Escrow.GetCommission())
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

func (tx *MultiSignatureTransaction) EscrowApproval(blockTimestamp int64, txBody *model.ApprovalEscrowTransactionBody) (err error) {
	switch txBody.GetApproval() {
	case model.EscrowApproval_Approve:
		tx.TransactionObject.Escrow.Status = model.EscrowStatus_Approved
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.TransactionObject.SenderAccountAddress,
			tx.TransactionObject.Fee,
			model.EventType_EventEscrowedTransaction,
			tx.TransactionObject.Height,
			tx.TransactionObject.ID,
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
			tx.TransactionObject.Escrow.GetApproverAddress(),
			tx.TransactionObject.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.TransactionObject.Height,
			tx.TransactionObject.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	case model.EscrowApproval_Reject:
		tx.TransactionObject.Escrow.Status = model.EscrowStatus_Rejected
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.TransactionObject.Escrow.GetApproverAddress(),
			tx.TransactionObject.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.TransactionObject.Height,
			tx.TransactionObject.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	default:
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.TransactionObject.SenderAccountAddress,
			tx.TransactionObject.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.TransactionObject.Height,
			tx.TransactionObject.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	}
	escrowQ := tx.EscrowQuery.InsertEscrowTransaction(tx.TransactionObject.Escrow)
	err = tx.QueryExecutor.ExecuteTransactions(escrowQ)
	if err != nil {
		return err
	}
	return nil
}
