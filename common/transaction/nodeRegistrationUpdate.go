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

	"github.com/zoobc/zoobc-core/common/accounttype"

	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/util"
)

// UpdateNodeRegistration Implement service layer for (new) node registration's transaction
type UpdateNodeRegistration struct {
	TransactionObject            *model.Transaction
	Body                         *model.UpdateNodeRegistrationTransactionBody
	NodeRegistrationQuery        query.NodeRegistrationQueryInterface
	BlockQuery                   query.BlockQueryInterface
	QueryExecutor                query.ExecutorInterface
	AuthPoown                    auth.NodeAuthValidationInterface
	EscrowQuery                  query.EscrowTransactionQueryInterface
	AccountBalanceHelper         AccountBalanceHelperInterface
	EscrowFee                    fee.FeeModelInterface
	NormalFee                    fee.FeeModelInterface
	PendingNodeRegistrationCache storage.TransactionalCache
	ActiveNodeRegistrationCache  storage.TransactionalCache
}

// SkipMempoolTransaction filter out of the mempool a node registration tx if there are other node registration tx in mempool
// to make sure only one node registration tx at the time (the one with highest fee paid) makes it to the same block
func (tx *UpdateNodeRegistration) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	newBlockTimestamp int64,
	newBlockHeight uint32,
) (bool, error) {
	authorizedType := map[model.TransactionType]bool{
		model.TransactionType_ClaimNodeRegistrationTransaction:  true,
		model.TransactionType_UpdateNodeRegistrationTransaction: true,
		model.TransactionType_RemoveNodeRegistrationTransaction: true,
	}
	for _, sel := range selectedTransactions {
		// if we find another node registration tx in currently selected transactions, filter current one out of selection
		if _, ok := authorizedType[model.TransactionType(sel.GetTransactionType())]; ok &&
			bytes.Equal(tx.TransactionObject.SenderAccountAddress, sel.SenderAccountAddress) {
			return true, nil
		}
	}
	return false, nil
}

// ApplyConfirmed method for confirmed the transaction and store into database
func (tx *UpdateNodeRegistration) ApplyConfirmed(blockTimestamp int64) error {
	var (
		effectiveBalanceToLock, lockedBalance int64
		nodePublicKey                         []byte
		nodeReg                               model.NodeRegistration
		row                                   *sql.Row
		err                                   error
	)

	// get the latest node registration by owner (sender account)
	qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.TransactionObject.SenderAccountAddress)
	row, _ = tx.QueryExecutor.ExecuteSelectRow(qry, false, args...)
	err = tx.NodeRegistrationQuery.Scan(&nodeReg, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.AppErr, "NodeNotFoundWithAccountAddress")
	}

	if tx.Body.GetLockedBalance() > 0 {
		lockedBalance = tx.Body.GetLockedBalance()
	} else {
		lockedBalance = nodeReg.GetLockedBalance()
	}

	if len(tx.Body.GetNodePublicKey()) != 0 {
		nodePublicKey = tx.Body.NodePublicKey
	} else {
		nodePublicKey = nodeReg.GetNodePublicKey()
	}

	if tx.Body.LockedBalance > 0 {
		// delta amount to be locked
		effectiveBalanceToLock = tx.Body.GetLockedBalance() - nodeReg.GetLockedBalance()
	}

	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.TransactionObject.SenderAccountAddress,
		-(effectiveBalanceToLock + tx.TransactionObject.Fee),
		model.EventType_EventUpdateNodeRegistrationTransaction,
		tx.TransactionObject.Height,
		tx.TransactionObject.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}
	nodeQueries := tx.NodeRegistrationQuery.UpdateNodeRegistration(&model.NodeRegistration{
		NodeID:             nodeReg.GetNodeID(),
		LockedBalance:      lockedBalance,
		Height:             tx.TransactionObject.Height,
		RegistrationHeight: nodeReg.GetRegistrationHeight(),
		NodePublicKey:      nodePublicKey,
		Latest:             true,
		RegistrationStatus: nodeReg.GetRegistrationStatus(),
		// account address is the only field that can't be updated via update node registration
		AccountAddress: nodeReg.GetAccountAddress(),
	})

	err = tx.QueryExecutor.ExecuteTransactions(nodeQueries)
	if err != nil {
		return err
	}
	// update cache by replace
	switch model.NodeRegistrationState(nodeReg.GetRegistrationStatus()) {
	case model.NodeRegistrationState_NodeQueued:
		err = tx.PendingNodeRegistrationCache.TxSetItem(nodeReg.NodeID, storage.NodeRegistry{
			Node:               nodeReg,
			ParticipationScore: 0, // pending node registry doesn't have participation score yet
		})
	case model.NodeRegistrationState_NodeRegistered:
		var (
			nodeRegFromCache storage.NodeRegistry
		)
		// fetch previous state data to get the participation score
		if activeNonTxCache, ok := tx.ActiveNodeRegistrationCache.(storage.CacheStorageInterface); ok {
			err := activeNonTxCache.GetItem(nodeReg.GetNodeID(), &nodeRegFromCache)
			if err != nil {
				return err
			}
		}
		err = tx.ActiveNodeRegistrationCache.TxSetItem(nodeReg.NodeID, storage.NodeRegistry{
			Node:               nodeReg,
			ParticipationScore: nodeRegFromCache.ParticipationScore,
		})
	}
	return err
}

/*
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `UpdateNodeRegistration` type:
	- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *UpdateNodeRegistration) ApplyUnconfirmed() error {

	var (
		effectiveBalanceToLock int64
		nodeReg                model.NodeRegistration
		err                    error
		row                    *sql.Row
	)

	// update sender balance by reducing his spendable balance of the tx fee + new balance to be lock
	// (delta between old locked balance and update locked balance)
	if tx.Body.LockedBalance > 0 {
		// get the latest node registration by owner (sender account)
		qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.TransactionObject.SenderAccountAddress)
		row, _ = tx.QueryExecutor.ExecuteSelectRow(qry, false, args...)
		err = tx.NodeRegistrationQuery.Scan(&nodeReg, row)
		if err != nil {
			if err != sql.ErrNoRows {
				return err
			}
			return blocker.NewBlocker(blocker.AppErr, "NodeNotFoundWithAccountAddress")
		}

		// delta amount to be locked
		effectiveBalanceToLock = tx.Body.GetLockedBalance() - nodeReg.GetLockedBalance()
	}

	err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress,
		-(effectiveBalanceToLock + tx.TransactionObject.Fee))
	if err != nil {
		return err
	}
	return nil
}

func (tx *UpdateNodeRegistration) UndoApplyUnconfirmed() error {
	var (
		err                    error
		effectiveBalanceToLock int64
		prevNodeRegistration   model.NodeRegistration
		row                    *sql.Row
	)
	// get the latest nodeRegistration by owner (sender account)
	qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.TransactionObject.SenderAccountAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, false, args...)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	err = tx.NodeRegistrationQuery.Scan(&prevNodeRegistration, row)
	if err != nil {
		if err == sql.ErrNoRows {
			return blocker.NewBlocker(blocker.AppErr, "NodeNotFoundWithAccountAddress")
		}
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	// delta amount to be locked
	effectiveBalanceToLock = tx.Body.LockedBalance - prevNodeRegistration.LockedBalance
	// update sender balance by reducing his spendable balance of the tx fee
	err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, effectiveBalanceToLock+tx.TransactionObject.Fee)
	if err != nil {
		return err
	}
	return nil
}

// Validate validate node registration transaction and tx body
func (tx *UpdateNodeRegistration) Validate(dbTx bool) error {
	var (
		err                    error
		enough                 bool
		effectiveBalanceToLock int64
		prevNodeReg            model.NodeRegistration
		row                    *sql.Row
	)
	// formally validate tx body fields
	if tx.Body.Poown == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "PoownRequired")
	}

	// validate proof of ownership
	err = tx.AuthPoown.ValidateProofOfOwnership(tx.Body.Poown, tx.Body.NodePublicKey, tx.QueryExecutor, tx.BlockQuery)
	if err != nil {
		return err
	}

	nodeRegQ, nodeRegArgs := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.TransactionObject.SenderAccountAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(nodeRegQ, dbTx, nodeRegArgs...)
	if err != nil {
		return err
	}
	err = tx.NodeRegistrationQuery.Scan(&prevNodeReg, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "SenderAccountNotNodeOwner")
	}
	if prevNodeReg.GetRegistrationStatus() == uint32(model.NodeRegistrationState_NodeDeleted) {
		return blocker.NewBlocker(blocker.AuthErr, "NodeDeleted")
	}

	// validate node public key, if we are updating that field
	// note: node pub key must be not already registered for another node
	if len(tx.Body.NodePublicKey) > 0 && !bytes.Equal(prevNodeReg.NodePublicKey, tx.Body.NodePublicKey) {
		err = func() (e error) {
			row, e = tx.QueryExecutor.ExecuteSelectRow(tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(), dbTx, tx.Body.GetNodePublicKey())
			if e != nil {
				return e
			}
			e = tx.NodeRegistrationQuery.Scan(&model.NodeRegistration{}, row)
			if e != nil {
				if e != sql.ErrNoRows {
					return e
				}
				return nil
			}
			return blocker.NewBlocker(blocker.ValidationErr, "NodePublicKeyAlreadyRegistered")
		}()
		if err != nil {
			return err
		}
	}
	// delta amount to be locked
	effectiveBalanceToLock = tx.Body.GetLockedBalance() - prevNodeReg.GetLockedBalance()
	if effectiveBalanceToLock < 0 {
		// cannot lock less than what previously locked
		return blocker.NewBlocker(blocker.ValidationErr, "LockedBalanceLessThenPreviouslyLocked")
	}

	// check aalance
	enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTx, tx.TransactionObject.SenderAccountAddress, tx.TransactionObject.Fee+effectiveBalanceToLock)
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

func (tx *UpdateNodeRegistration) GetAmount() int64 {
	return 0
}

func (tx *UpdateNodeRegistration) GetMinimumFee() (int64, error) {
	if tx.TransactionObject.Escrow != nil && tx.TransactionObject.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.TransactionObject.Escrow.GetApproverAddress(), []byte{}) {
		return tx.EscrowFee.CalculateTxMinimumFee(tx.Body, tx.TransactionObject)
	}
	return tx.NormalFee.CalculateTxMinimumFee(tx.Body, tx.TransactionObject)
}

func (tx *UpdateNodeRegistration) GetSize() (uint32, error) {
	// ProofOfOwnership (message + signature)
	if tx.TransactionObject.SenderAccountAddress == nil {
		return 0, blocker.NewBlocker(blocker.ValidationErr, "SenderAddressRequired")
	}
	accType, err := accounttype.NewAccountTypeFromAccount(tx.TransactionObject.SenderAccountAddress)
	if err != nil {
		return 0, err
	}
	poown := util.GetProofOfOwnershipSize(accType, true)
	return constant.NodePublicKey + constant.Balance + poown, nil
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (tx *UpdateNodeRegistration) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {

	var (
		nodePublicKey []byte
		lockedBalance uint64
		poown         *model.ProofOfOwnership
		err           error
	)
	// read body bytes
	buffer := bytes.NewBuffer(txBodyBytes)

	nodePublicKey, err = util.ReadTransactionBytes(buffer, int(constant.NodePublicKey))
	if err != nil {
		return nil, err
	}

	lockedBalanceBytes, err := util.ReadTransactionBytes(buffer, int(constant.Balance))
	if err != nil {
		return nil, err
	}
	lockedBalance = util.ConvertBytesToUint64(lockedBalanceBytes)

	// get the poown account type by parsing proof of ownership bytes
	var tmpPoownBytes = make([]byte, buffer.Len())
	copy(tmpPoownBytes, buffer.Bytes())
	tmpBuffer := bytes.NewBuffer(tmpPoownBytes)
	poownAccType, err := accounttype.ParseBytesToAccountType(tmpBuffer)
	if err != nil {
		return nil, err
	}
	poownBytes, err := util.ReadTransactionBytes(buffer, int(util.GetProofOfOwnershipSize(poownAccType, true)))
	if err != nil {
		return nil, err
	}
	poown, err = util.ParseProofOfOwnershipBytes(poownBytes)
	if err != nil {
		return nil, err
	}
	return &model.UpdateNodeRegistrationTransactionBody{
		NodePublicKey: nodePublicKey,
		LockedBalance: int64(lockedBalance),
		Poown:         poown,
	}, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *UpdateNodeRegistration) GetBodyBytes() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(tx.Body.NodePublicKey)
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.LockedBalance)))
	// convert ProofOfOwnership (message + signature) to bytes
	buffer.Write(util.GetProofOfOwnershipBytes(tx.Body.Poown))
	return buffer.Bytes(), nil
}

func (tx *UpdateNodeRegistration) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_UpdateNodeRegistrationTransactionBody{
		UpdateNodeRegistrationTransactionBody: tx.Body,
	}
}

/*
Escrowable will check the transaction is escrow or not.
Rebuild escrow if not nil, and can use for whole sibling methods (escrow)
*/
func (tx *UpdateNodeRegistration) Escrowable() (EscrowTypeAction, bool) {
	if tx.TransactionObject.Escrow != nil && tx.TransactionObject.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.TransactionObject.Escrow.GetApproverAddress(), []byte{}) {
		tx.TransactionObject.Escrow = util.PrepareEscrowObjectForAction(tx.TransactionObject)
		return EscrowTypeAction(tx), true
	}
	return nil, false
}

// EscrowValidate validate node registration transaction and tx body
func (tx *UpdateNodeRegistration) EscrowValidate(dbTx bool) error {
	var (
		effectiveBalanceToLock int64
		err                    error
		enough                 bool
		prevNodeReg            model.NodeRegistration
	)

	err = util.ValidateBasicEscrow(tx.TransactionObject)
	if err != nil {
		return err
	}

	err = tx.Validate(dbTx)
	if err != nil {
		return err
	}

	nodeRegQ, nodeRegArgs := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.TransactionObject.SenderAccountAddress)
	row, _ := tx.QueryExecutor.ExecuteSelectRow(nodeRegQ, dbTx, nodeRegArgs...)
	err = tx.NodeRegistrationQuery.Scan(&prevNodeReg, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "SenderAccountNotNodeOwner")
	}
	effectiveBalanceToLock = tx.Body.GetLockedBalance() - prevNodeReg.GetLockedBalance()

	enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTx,
		tx.TransactionObject.SenderAccountAddress,
		tx.TransactionObject.Fee+tx.TransactionObject.Escrow.GetCommission()+effectiveBalanceToLock)
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

/*
EscrowApplyUnconfirmed is func that for applying to unconfirmed Transaction `UpdateNodeRegistration` type,
perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *UpdateNodeRegistration) EscrowApplyUnconfirmed() error {

	var (
		effectiveBalanceToLock int64
		nodeRegistration       model.NodeRegistration
		err                    error
		row                    *sql.Row
	)

	// update sender balance by reducing his spendable balance of the tx fee + new balance to be lock
	// (delta between old locked balance and update locked balance)
	if tx.Body.LockedBalance > 0 {
		// get the latest node registration by owner (sender account)
		qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.TransactionObject.SenderAccountAddress)
		row, err = tx.QueryExecutor.ExecuteSelectRow(qry, false, args...)
		if err != nil {
			return err
		}
		err = tx.NodeRegistrationQuery.Scan(&nodeRegistration, row)
		if err != nil {
			if err != sql.ErrNoRows {
				return err
			}
			// assume no row
			return blocker.NewBlocker(blocker.AppErr, "NodeNotFoundWithAccountAddress")

		}
		// delta amount to be locked
		effectiveBalanceToLock = tx.Body.GetLockedBalance() - nodeRegistration.GetLockedBalance()
	}

	err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, -(effectiveBalanceToLock + tx.TransactionObject.Fee + tx.TransactionObject.Escrow.GetCommission()))
	if err != nil {
		return err
	}
	return nil
}

/*
EscrowUndoApplyUnconfirmed func that perform on apply confirm preparation
*/
func (tx *UpdateNodeRegistration) EscrowUndoApplyUnconfirmed() error {
	var (
		effectiveBalanceToLock int64
		nodeRegistration       model.NodeRegistration
		row                    *sql.Row
		err                    error
	)

	// get the latest node registration by owner (sender account)
	qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.TransactionObject.SenderAccountAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, false, args...)
	if err != nil {
		return err

	}
	err = tx.NodeRegistrationQuery.Scan(&nodeRegistration, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.AppErr, "NodeNotFoundWithAccountAddress")
	}

	// delta amount to be locked
	effectiveBalanceToLock = tx.Body.GetLockedBalance() - nodeRegistration.GetLockedBalance()

	err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress,
		effectiveBalanceToLock+tx.TransactionObject.Fee+tx.TransactionObject.Escrow.GetCommission())
	if err != nil {
		return err
	}
	return nil
}

// EscrowApplyConfirmed method for confirmed the transaction and store into database
func (tx *UpdateNodeRegistration) EscrowApplyConfirmed(blockTimestamp int64) error {
	var (
		effectiveBalanceToLock int64
		nodeRegistration       model.NodeRegistration
		row                    *sql.Row
		err                    error
	)

	// get the latest node registration by owner (sender account)
	qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.TransactionObject.SenderAccountAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, false, args...)
	if err != nil {
		return err
	}
	err = tx.NodeRegistrationQuery.Scan(&nodeRegistration, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.AppErr, "NodeNotFoundWithAccountAddress")
	}

	if tx.Body.LockedBalance > 0 {
		// delta amount to be locked
		effectiveBalanceToLock = tx.Body.GetLockedBalance() - nodeRegistration.GetLockedBalance()
	}

	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.TransactionObject.SenderAccountAddress,
		-(effectiveBalanceToLock + tx.TransactionObject.Fee + tx.TransactionObject.Escrow.GetCommission()),
		model.EventType_EventEscrowedTransaction,
		tx.TransactionObject.Height,
		tx.TransactionObject.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}

	return nil
}

/*
EscrowApproval handle approval an escrow transaction, execute tasks that was skipped when escrow pending.
like: spreading commission and fee, and also more pending tasks
*/
func (tx *UpdateNodeRegistration) EscrowApproval(
	blockTimestamp int64,
	txBody *model.ApprovalEscrowTransactionBody,
) error {
	var (
		nodeRegistration       model.NodeRegistration
		err                    error
		row                    *sql.Row
		effectiveBalanceToLock int64
	)

	// get the latest node registration by owner (sender account)
	qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.TransactionObject.SenderAccountAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, false, args...)
	if err != nil {
		return err
	}
	err = tx.NodeRegistrationQuery.Scan(&nodeRegistration, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.AppErr, "NodeNotFoundWithAccountAddress")
	}

	if tx.Body.LockedBalance > 0 {
		// delta amount to be locked
		effectiveBalanceToLock = tx.Body.GetLockedBalance() - nodeRegistration.GetLockedBalance()
	}

	switch txBody.GetApproval() {
	case model.EscrowApproval_Approve:
		tx.TransactionObject.Escrow.Status = model.EscrowStatus_Approved
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.TransactionObject.SenderAccountAddress,
			effectiveBalanceToLock+tx.TransactionObject.Fee,
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
			tx.TransactionObject.SenderAccountAddress,
			effectiveBalanceToLock,
			model.EventType_EventApprovalEscrowTransaction,
			tx.TransactionObject.Height,
			tx.TransactionObject.ID,
			uint64(blockTimestamp),
		)
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
	default:
		tx.TransactionObject.Escrow.Status = model.EscrowStatus_Expired
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.TransactionObject.SenderAccountAddress,
			effectiveBalanceToLock+tx.TransactionObject.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.TransactionObject.Height,
			tx.TransactionObject.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	}

	// Insert Escrow
	escrowQ := tx.EscrowQuery.InsertEscrowTransaction(tx.TransactionObject.Escrow)
	err = tx.QueryExecutor.ExecuteTransactions(escrowQ)
	if err != nil {
		return err
	}
	return nil
}
