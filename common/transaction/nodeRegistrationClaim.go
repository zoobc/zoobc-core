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

// ClaimNodeRegistration Implement service layer for claim node registration's transaction
type ClaimNodeRegistration struct {
	TransactionObject       *model.Transaction
	Body                    *model.ClaimNodeRegistrationTransactionBody
	NodeRegistrationQuery   query.NodeRegistrationQueryInterface
	BlockQuery              query.BlockQueryInterface
	QueryExecutor           query.ExecutorInterface
	AuthPoown               auth.NodeAuthValidationInterface
	EscrowQuery             query.EscrowTransactionQueryInterface
	AccountBalanceHelper    AccountBalanceHelperInterface
	EscrowFee               fee.FeeModelInterface
	NormalFee               fee.FeeModelInterface
	NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
	NodeAddressInfoStorage  storage.TransactionalCache
	ActiveNodeRegistryCache storage.TransactionalCache
}

// SkipMempoolTransaction filter out of the mempool a node registration tx if there are other node registration tx in mempool
// to make sure only one node registration tx at the time (the one with highest fee paid) makes it to the same block
func (tx *ClaimNodeRegistration) SkipMempoolTransaction(
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

func (tx *ClaimNodeRegistration) ApplyConfirmed(blockTimestamp int64) error {
	var (
		nodeReg model.NodeRegistration
		row     *sql.Row
		err     error
		queries [][]interface{}
	)

	row, _ = tx.QueryExecutor.ExecuteSelectRow(tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(), false, tx.Body.GetNodePublicKey())
	err = tx.NodeRegistrationQuery.Scan(&nodeReg, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.AppErr, "NodePublicKeyNotRegistered")
	}

	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.TransactionObject.SenderAccountAddress,
		nodeReg.GetLockedBalance()-tx.TransactionObject.Fee,
		model.EventType_EventClaimNodeRegistrationTransaction,
		tx.TransactionObject.Height,
		tx.TransactionObject.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}

	// tag the node as deleted
	nodeQueries := tx.NodeRegistrationQuery.UpdateNodeRegistration(&model.NodeRegistration{
		NodeID:             nodeReg.GetNodeID(),
		LockedBalance:      0,
		Height:             tx.TransactionObject.Height,
		RegistrationHeight: nodeReg.GetRegistrationHeight(),
		NodePublicKey:      tx.Body.NodePublicKey,
		Latest:             true,
		RegistrationStatus: uint32(model.NodeRegistrationState_NodeDeleted),
		// We can't just set accountAddress to an empty string,
		// otherwise it could trigger an error when parsing the transaction from its bytes
		AccountAddress: nodeReg.GetAccountAddress(),
	})
	queries = append(queries, nodeQueries...)
	// remove the node_address_info
	removeNodeAddressInfoQ, removeNodeAddressInfoArgs := tx.NodeAddressInfoQuery.DeleteNodeAddressInfoByNodeID(
		nodeReg.NodeID,
		[]model.NodeAddressStatus{
			model.NodeAddressStatus_NodeAddressPending,
			model.NodeAddressStatus_NodeAddressConfirmed,
			model.NodeAddressStatus_Unset,
		},
	)
	removeNodeAddressInfoQueries := append([]interface{}{removeNodeAddressInfoQ}, removeNodeAddressInfoArgs...)
	queries = append(queries, removeNodeAddressInfoQueries)
	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}

	// Remove Node Address Info on cache storage
	err = tx.NodeAddressInfoStorage.TxRemoveItem(
		storage.NodeAddressInfoStorageKey{
			NodeID: nodeReg.NodeID,
			Statuses: []model.NodeAddressStatus{
				model.NodeAddressStatus_NodeAddressPending,
				model.NodeAddressStatus_NodeAddressConfirmed,
				model.NodeAddressStatus_Unset,
			},
		},
	)
	if err != nil {
		return err
	}
	err = tx.ActiveNodeRegistryCache.TxRemoveItem(nodeReg.NodeID)
	return err
}

/*
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `ClaimNodeRegistration` type:
	- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *ClaimNodeRegistration) ApplyUnconfirmed() error {

	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, -(tx.TransactionObject.Fee))
	if err != nil {
		return err
	}

	return nil
}

func (tx *ClaimNodeRegistration) UndoApplyUnconfirmed() error {

	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, tx.TransactionObject.Fee)
	if err != nil {
		return err
	}
	return nil
}

// Validate validate node registration transaction and tx body
func (tx *ClaimNodeRegistration) Validate(dbTx bool) error {
	var (
		nodeRegistration model.NodeRegistration
		row              *sql.Row
		err              error
		enough           bool
	)

	// validate proof of ownership
	if tx.Body.Poown == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "PoownRequired")
	}
	err = tx.AuthPoown.ValidateProofOfOwnership(
		tx.Body.Poown, tx.Body.NodePublicKey,
		tx.QueryExecutor,
		tx.BlockQuery)
	if err != nil {
		return err
	}

	row, err = tx.QueryExecutor.ExecuteSelectRow(tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(), dbTx, tx.Body.NodePublicKey)
	if err != nil {
		return err
	}
	err = tx.NodeRegistrationQuery.Scan(&nodeRegistration, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "NodePublicKeyNotRegistered")
	}

	if nodeRegistration.GetRegistrationStatus() == uint32(model.NodeRegistrationState_NodeDeleted) {
		return blocker.NewBlocker(blocker.ValidationErr, "NodeAlreadyClaimedOrDeleted")
	}

	enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTx, tx.TransactionObject.SenderAccountAddress, tx.TransactionObject.Fee)
	if err != nil {
		return err
	}
	if !enough {
		return blocker.NewBlocker(blocker.ValidationErr, "BalanceNotEnough")
	}

	return nil
}

func (tx *ClaimNodeRegistration) GetAmount() int64 {
	return 0
}

func (tx *ClaimNodeRegistration) GetMinimumFee() (int64, error) {
	if tx.TransactionObject.Escrow != nil && tx.TransactionObject.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.TransactionObject.Escrow.GetApproverAddress(), []byte{}) {
		return tx.EscrowFee.CalculateTxMinimumFee(tx.Body, tx.TransactionObject)
	}
	return tx.NormalFee.CalculateTxMinimumFee(tx.Body, tx.TransactionObject)
}

func (tx *ClaimNodeRegistration) GetSize() (uint32, error) {
	// ProofOfOwnership (message + signature)
	if tx.TransactionObject.SenderAccountAddress == nil {
		return 0, blocker.NewBlocker(blocker.ValidationErr, "SenderAddressRequired")
	}
	senderAccType, err := accounttype.NewAccountTypeFromAccount(tx.TransactionObject.SenderAccountAddress)
	if err != nil {
		return 0, err
	}
	poownSize := util.GetProofOfOwnershipSize(senderAccType, true)
	accountAddressSize := constant.AccountAddressTypeLength + senderAccType.GetAccountPublicKeyLength()
	return accountAddressSize + constant.NodePublicKey + poownSize, nil
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (tx *ClaimNodeRegistration) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	buffer := bytes.NewBuffer(txBodyBytes)
	nodePublicKey, err := util.ReadTransactionBytes(buffer, int(constant.NodePublicKey))
	if err != nil {
		return nil, err
	}
	// get the poown account type by parsing proof of ownership bytes
	var tmpPoownBytes = make([]byte, buffer.Len())
	copy(tmpPoownBytes, buffer.Bytes())
	tmpBuffer := bytes.NewBuffer(tmpPoownBytes)
	poownAccType, err := accounttype.ParseBytesToAccountType(tmpBuffer)
	if err != nil {
		return nil, err
	}
	poown, err := util.ParseProofOfOwnershipBytes(buffer.Next(int(util.GetProofOfOwnershipSize(poownAccType, true))))
	if err != nil {
		return nil, err
	}
	return &model.ClaimNodeRegistrationTransactionBody{
		NodePublicKey: nodePublicKey,
		Poown:         poown,
	}, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *ClaimNodeRegistration) GetBodyBytes() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(tx.Body.NodePublicKey)
	// convert ProofOfOwnership (message + signature) to bytes
	buffer.Write(util.GetProofOfOwnershipBytes(tx.Body.Poown))
	return buffer.Bytes(), nil
}

func (tx *ClaimNodeRegistration) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_ClaimNodeRegistrationTransactionBody{
		ClaimNodeRegistrationTransactionBody: tx.Body,
	}
}

/*
Escrowable will check the transaction is escrow or not.
Rebuild escrow if not nil, and can use for whole sibling methods (escrow)
*/
func (tx *ClaimNodeRegistration) Escrowable() (EscrowTypeAction, bool) {
	if tx.TransactionObject.Escrow != nil && tx.TransactionObject.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.TransactionObject.Escrow.GetApproverAddress(), []byte{}) {
		tx.TransactionObject.Escrow = util.PrepareEscrowObjectForAction(tx.TransactionObject)
		return EscrowTypeAction(tx), true
	}
	return nil, false
}

// EscrowValidate validate node registration transaction and tx body
func (tx *ClaimNodeRegistration) EscrowValidate(dbTX bool) error {
	var (
		err    error
		enough bool
	)

	err = util.ValidateBasicEscrow(tx.TransactionObject)
	if err != nil {
		return err
	}

	err = tx.Validate(dbTX)
	if err != nil {
		return err
	}

	enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTX, tx.TransactionObject.SenderAccountAddress, tx.TransactionObject.Fee+tx.TransactionObject.Escrow.GetCommission())
	if err != nil {
		return err
	}
	if !enough {
		return blocker.NewBlocker(blocker.ValidationErr, "AccountBalanceNotEnough")
	}
	return nil
}

/*
EscrowApplyUnconfirmed is func that for applying to unconfirmed Transaction `ClaimNodeRegistration` type:
	- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *ClaimNodeRegistration) EscrowApplyUnconfirmed() error {
	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress,
		-(tx.TransactionObject.Fee + tx.TransactionObject.Escrow.GetCommission()))
	if err != nil {
		return err
	}
	return nil
}

/*
EscrowUndoApplyUnconfirmed func that perform on apply confirm preparation
*/
func (tx *ClaimNodeRegistration) EscrowUndoApplyUnconfirmed() error {
	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress,
		tx.TransactionObject.Fee+tx.TransactionObject.Escrow.GetCommission())
	if err != nil {
		return err
	}
	return nil
}

/*
EscrowApplyConfirmed func that for applying pending escrow transaction.
*/
func (tx *ClaimNodeRegistration) EscrowApplyConfirmed(blockTimestamp int64) error {
	var (
		prevNodeRegistration *model.NodeRegistration
		err                  error
		row                  *sql.Row
	)

	row, err = tx.QueryExecutor.ExecuteSelectRow(
		tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(),
		false,
		tx.Body.NodePublicKey,
	)
	if err != nil {
		return err
	}
	err = row.Scan(&prevNodeRegistration)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.AppErr, "NodePublicKeyNotRegistered")
	}

	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.TransactionObject.SenderAccountAddress,
		-(tx.TransactionObject.Fee + tx.TransactionObject.Escrow.GetCommission()),
		model.EventType_EventEscrowedTransaction,
		tx.TransactionObject.Height,
		tx.TransactionObject.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}

	escrowQ := tx.EscrowQuery.InsertEscrowTransaction(tx.TransactionObject.Escrow)
	err = tx.QueryExecutor.ExecuteTransactions(escrowQ)
	if err != nil {
		return err
	}
	return nil
}

/*
EscrowApproval handle approval an escrow transaction, execute tasks that was skipped when escrow pending.
like: spreading commission and fee, and also more pending tasks
*/
func (tx *ClaimNodeRegistration) EscrowApproval(
	blockTimestamp int64,
	txBody *model.ApprovalEscrowTransactionBody,
) error {
	var (
		prevNodeRegistration model.NodeRegistration
		row                  *sql.Row
		err                  error
	)

	row, err = tx.QueryExecutor.ExecuteSelectRow(
		tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(),
		false,
		tx.Body.NodePublicKey,
	)
	if err != nil {
		return err
	}
	err = row.Scan(&prevNodeRegistration)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.AppErr, "NodePublicKeyNotRegistered")
	}

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
		tx.TransactionObject.Escrow.Status = model.EscrowStatus_Expired
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
	// Insert Escrow
	escrowQ := tx.EscrowQuery.InsertEscrowTransaction(tx.TransactionObject.Escrow)
	err = tx.QueryExecutor.ExecuteTransactions(escrowQ)
	if err != nil {
		return err
	}

	return nil
}
