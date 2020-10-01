package transaction

import (
	"bytes"
	"database/sql"
	"github.com/zoobc/zoobc-core/common/storage"

	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

// UpdateNodeRegistration Implement service layer for (new) node registration's transaction
type UpdateNodeRegistration struct {
	ID                           int64
	Fee                          int64
	SenderAddress                []byte
	Height                       uint32
	Timestamp                    int64
	Body                         *model.UpdateNodeRegistrationTransactionBody
	Escrow                       *model.Escrow
	AccountBalanceQuery          query.AccountBalanceQueryInterface
	NodeRegistrationQuery        query.NodeRegistrationQueryInterface
	BlockQuery                   query.BlockQueryInterface
	QueryExecutor                query.ExecutorInterface
	AuthPoown                    auth.NodeAuthValidationInterface
	AccountLedgerQuery           query.AccountLedgerQueryInterface
	EscrowQuery                  query.EscrowTransactionQueryInterface
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
			bytes.Equal(tx.SenderAddress, sel.SenderAccountAddress) {
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
		queries                               [][]interface{}
		row                                   *sql.Row
		err                                   error
	)

	// get the latest node registration by owner (sender account)
	qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.SenderAddress)
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

	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		-(effectiveBalanceToLock + tx.Fee),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)
	queries = append(queries, accountBalanceSenderQ...)

	nodeQueries := tx.NodeRegistrationQuery.UpdateNodeRegistration(&model.NodeRegistration{
		NodeID:             nodeReg.GetNodeID(),
		LockedBalance:      lockedBalance,
		Height:             tx.Height,
		RegistrationHeight: nodeReg.GetRegistrationHeight(),
		NodePublicKey:      nodePublicKey,
		Latest:             true,
		RegistrationStatus: nodeReg.GetRegistrationStatus(),
		// account address is the only field that can't be updated via update node registration
		AccountAddress: nodeReg.GetAccountAddress(),
	})
	queries = append(queries, nodeQueries...)

	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  -(effectiveBalanceToLock + tx.Fee),
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventUpdateNodeRegistrationTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	senderAccountLedgerArgs = append([]interface{}{senderAccountLedgerQ}, senderAccountLedgerArgs...)
	queries = append(queries, senderAccountLedgerArgs)

	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}
	// update cache by replace
	switch model.NodeRegistrationState(nodeReg.GetRegistrationStatus()) {
	case model.NodeRegistrationState_NodeQueued:
		err = tx.PendingNodeRegistrationCache.TxSetItem(nodeReg.NodeID, nodeReg)
	case model.NodeRegistrationState_NodeRegistered:
		err = tx.ActiveNodeRegistrationCache.TxSetItem(nodeReg.NodeID, nodeReg)
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
		qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.SenderAddress)
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

	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-(effectiveBalanceToLock + tx.Fee),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
		},
	)
	err = tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
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
	)
	// get the latest noderegistration by owner (sender account)
	qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.SenderAddress)
	row, err := tx.QueryExecutor.ExecuteSelectRow(qry, false, args...)
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
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		effectiveBalanceToLock+tx.Fee,
		map[string]interface{}{
			"account_address": tx.SenderAddress,
		},
	)
	err = tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	if err != nil {
		return err
	}
	return nil
}

// Validate validate node registration transaction and tx body
func (tx *UpdateNodeRegistration) Validate(dbTx bool) error {
	var (
		accountBalance                                          model.AccountBalance
		prevNodeRegistration                                    *model.NodeRegistration
		tempNodeRegistrationResult, tempNodeRegistrationResult2 []*model.NodeRegistration
	)
	// formally validate tx body fields
	if tx.Body.Poown == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "PoownRequired")
	}

	// validate proof of ownership
	if err := tx.AuthPoown.ValidateProofOfOwnership(
		tx.Body.Poown, tx.Body.NodePublicKey,
		tx.QueryExecutor,
		tx.BlockQuery); err != nil {
		return err
	}
	err := func() error {
		// check that sender is node's owner
		qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.SenderAddress)
		rows, err := tx.QueryExecutor.ExecuteSelect(qry, dbTx, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		tempNodeRegistrationResult, err = tx.NodeRegistrationQuery.BuildModel(tempNodeRegistrationResult, rows)
		if (err != nil) || len(tempNodeRegistrationResult) > 0 {
			prevNodeRegistration = tempNodeRegistrationResult[0]
			if prevNodeRegistration.RegistrationStatus == uint32(model.NodeRegistrationState_NodeDeleted) {
				return blocker.NewBlocker(blocker.AuthErr, "NodeDeleted")
			}
		} else {
			return blocker.NewBlocker(blocker.ValidationErr, "SenderAccountNotNodeOwner")
		}
		return nil
	}()
	if err != nil {
		return err
	}

	// validate node public key, if we are updating that field
	// note: node pub key must be not already registered for another node
	if len(tx.Body.NodePublicKey) > 0 && !bytes.Equal(prevNodeRegistration.NodePublicKey, tx.Body.NodePublicKey) {
		err := func() error {
			rows2, err := tx.QueryExecutor.ExecuteSelect(tx.NodeRegistrationQuery.
				GetNodeRegistrationByNodePublicKey(), false, tx.Body.NodePublicKey)
			if err != nil {
				return err
			}
			defer rows2.Close()

			tempNodeRegistrationResult2, err = tx.NodeRegistrationQuery.BuildModel(tempNodeRegistrationResult2, rows2)
			if (err != nil) || len(tempNodeRegistrationResult2) > 0 {
				return blocker.NewBlocker(blocker.ValidationErr, "NodePublicKeyAlredyRegistered")
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}

	// delta amount to be locked
	var effectiveBalanceToLock = tx.Body.LockedBalance - prevNodeRegistration.LockedBalance
	if effectiveBalanceToLock < 0 {
		// cannot lock less than what previously locked
		return blocker.NewBlocker(blocker.ValidationErr, "LockedBalanceLessThenPreviouslyLocked")
	}

	// check balance
	qry, args := tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
	row3, err := tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	err = tx.AccountBalanceQuery.Scan(&accountBalance, row3)
	if err != nil {
		if err == sql.ErrNoRows {
			return blocker.NewBlocker(blocker.AppErr, "SenderAccountAddressNotFound")
		}
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	if accountBalance.SpendableBalance < tx.Fee+effectiveBalanceToLock {
		return blocker.NewBlocker(blocker.ValidationErr, "UserBalanceNotEnough")
	}

	return nil
}

func (tx *UpdateNodeRegistration) GetAmount() int64 {
	return tx.Body.LockedBalance
}

func (*UpdateNodeRegistration) GetMinimumFee() (int64, error) {
	return 0, nil
}

func (tx *UpdateNodeRegistration) GetSize() uint32 {
	// ProofOfOwnership (message + signature)
	poown := util.GetProofOfOwnershipSize(true)
	return constant.NodePublicKey + constant.Balance + poown
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

	// parse ProofOfOwnership (message + signature) bytes
	poownBytes, err := util.ReadTransactionBytes(buffer, int(util.GetProofOfOwnershipSize(true)))
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
func (tx *UpdateNodeRegistration) GetBodyBytes() []byte {

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(tx.Body.NodePublicKey)
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.LockedBalance)))
	// convert ProofOfOwnership (message + signature) to bytes
	buffer.Write(util.GetProofOfOwnershipBytes(tx.Body.Poown))
	return buffer.Bytes()
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
	return nil, false
}

// EscrowValidate validate node registration transaction and tx body
func (tx *UpdateNodeRegistration) EscrowValidate(dbTx bool) error {
	return nil
}

/*
EscrowApplyUnconfirmed is func that for applying to unconfirmed Transaction `UpdateNodeRegistration` type,
perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *UpdateNodeRegistration) EscrowApplyUnconfirmed() error {
	return nil
}

/*
EscrowUndoApplyUnconfirmed func that perform on apply confirm preparation
*/
func (tx *UpdateNodeRegistration) EscrowUndoApplyUnconfirmed() error {
	return nil
}

// EscrowApplyConfirmed method for confirmed the transaction and store into database
func (tx *UpdateNodeRegistration) EscrowApplyConfirmed(blockTimestamp int64) error {

	return nil
}

/*
EscrowApproval handle approval an escrow transaction, execute tasks that was skipped when escrow pending.
like: spreading commission and fee, and also more pending tasks
*/
func (tx *UpdateNodeRegistration) EscrowApproval(int64, *model.ApprovalEscrowTransactionBody) error {
	return nil
}
