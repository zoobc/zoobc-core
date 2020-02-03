package transaction

import (
	"bytes"
	"database/sql"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

// RemoveNodeRegistration Implement service layer for (new) node registration's transaction
type RemoveNodeRegistration struct {
	ID                    int64
	Fee                   int64
	SenderAddress         string
	Height                uint32
	Body                  *model.RemoveNodeRegistrationTransactionBody
	Escrow                *model.Escrow
	AccountBalanceQuery   query.AccountBalanceQueryInterface
	NodeRegistrationQuery query.NodeRegistrationQueryInterface
	NodeAddressInfoQuery  query.NodeAddressInfoQueryInterface
	QueryExecutor         query.ExecutorInterface
	AccountLedgerQuery    query.AccountLedgerQueryInterface
	AccountBalanceHelper  AccountBalanceHelperInterface
	EscrowQuery           query.EscrowTransactionQueryInterface
}

// SkipMempoolTransaction filter out of the mempool a node registration tx if there are other node registration tx in mempool
// to make sure only one node registration tx at the time (the one with highest fee paid) makes it to the same block
func (tx *RemoveNodeRegistration) SkipMempoolTransaction(
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
		if _, ok := authorizedType[model.TransactionType(sel.GetTransactionType())]; ok && tx.SenderAddress == sel.SenderAccountAddress {
			return true, nil
		}
	}
	return false, nil
}

// ApplyConfirmed method for confirmed the transaction and store into database
func (tx *RemoveNodeRegistration) ApplyConfirmed(blockTimestamp int64) error {

	var (
		nodeReg model.NodeRegistration
		queries [][]interface{}
		row     *sql.Row
		err     error
	)

	row, _ = tx.QueryExecutor.ExecuteSelectRow(tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(), false, tx.Body.GetNodePublicKey())
	err = tx.NodeRegistrationQuery.Scan(&nodeReg, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.AppErr, "NodeNotRegistered")
	}

	// update sender balance by refunding the locked balance
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		nodeReg.GetLockedBalance()-tx.Fee,
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)
	queries = append(queries, accountBalanceSenderQ...)

	// tag the node as deleted
	nodeQueries := tx.NodeRegistrationQuery.UpdateNodeRegistration(&model.NodeRegistration{
		NodeID:             nodeReg.GetNodeID(),
		LockedBalance:      0,
		Height:             tx.Height,
		RegistrationHeight: nodeReg.GetRegistrationHeight(),
		NodePublicKey:      tx.Body.GetNodePublicKey(),
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
	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  nodeReg.GetLockedBalance() - tx.Fee,
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventRemoveNodeRegistrationTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	senderAccountLedgerArgs = append([]interface{}{senderAccountLedgerQ}, senderAccountLedgerArgs...)
	queries = append(queries, senderAccountLedgerArgs)

	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}

	return nil
}

/*
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `RemoveNodeRegistration` type:
	- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *RemoveNodeRegistration) ApplyUnconfirmed() error {

	var (
		err error
	)

	// update sender balance by reducing his spendable balance of the tx fee
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

	return nil
}

func (tx *RemoveNodeRegistration) UndoApplyUnconfirmed() error {
	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		tx.Fee,
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

// Validate validate node registration transaction and tx body
func (tx *RemoveNodeRegistration) Validate(dbTx bool) error {
	var (
		nodeReg        model.NodeRegistration
		err            error
		row            *sql.Row
		accountBalance model.AccountBalance
	)

	// check for duplication
	row, err = tx.QueryExecutor.ExecuteSelectRow(tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(), dbTx, tx.Body.GetNodePublicKey())
	if err != nil {
		return err
	}
	err = tx.NodeRegistrationQuery.Scan(&nodeReg, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.AppErr, "NodeNotRegistered")
	}

	// sender must be node owner
	if tx.SenderAddress != nodeReg.GetAccountAddress() {
		return blocker.NewBlocker(blocker.AuthErr, "AccountNotNodeOwner")
	}
	if nodeReg.GetRegistrationStatus() == uint32(model.NodeRegistrationState_NodeDeleted) {
		return blocker.NewBlocker(blocker.AuthErr, "NodeAlreadyDeleted")
	}
	// check existing & balance account sender
	err = tx.AccountBalanceHelper.GetBalanceByAccountID(&accountBalance, tx.SenderAddress, dbTx)
	if err != nil {
		return err
	}
	if accountBalance.GetSpendableBalance() < tx.Fee {
		return blocker.NewBlocker(blocker.ValidationErr, "BalanceNotEnough")
	}
	return nil
}

func (tx *RemoveNodeRegistration) GetAmount() int64 {
	return 0
}

func (*RemoveNodeRegistration) GetMinimumFee() (int64, error) {
	return 0, nil
}

func (tx *RemoveNodeRegistration) GetSize() uint32 {
	return constant.NodePublicKey
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (tx *RemoveNodeRegistration) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	// read body bytes
	buffer := bytes.NewBuffer(txBodyBytes)
	nodePublicKey, err := util.ReadTransactionBytes(buffer, int(constant.NodePublicKey))
	if err != nil {
		return nil, err
	}
	txBody := &model.RemoveNodeRegistrationTransactionBody{
		NodePublicKey: nodePublicKey,
	}
	return txBody, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *RemoveNodeRegistration) GetBodyBytes() []byte {

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(tx.Body.NodePublicKey)
	return buffer.Bytes()
}

func (tx *RemoveNodeRegistration) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_RemoveNodeRegistrationTransactionBody{
		RemoveNodeRegistrationTransactionBody: tx.Body,
	}
}

/*
Escrowable will check the transaction is escrow or not.
Rebuild escrow if not nil, and can use for whole sibling methods (escrow)
*/
func (tx *RemoveNodeRegistration) Escrowable() (EscrowTypeAction, bool) {
	if tx.Escrow.GetApproverAddress() != "" {
		tx.Escrow = &model.Escrow{
			ID:              tx.ID,
			SenderAddress:   tx.SenderAddress,
			ApproverAddress: tx.Escrow.GetApproverAddress(),
			Commission:      tx.Escrow.GetCommission(),
			Timeout:         tx.Escrow.GetTimeout(),
			Status:          tx.Escrow.GetStatus(),
			BlockHeight:     tx.Height,
			Latest:          true,
		}

		return EscrowTypeAction(tx), true
	}
	return nil, false
}

// EscrowValidate validate node registration transaction and tx body
func (tx *RemoveNodeRegistration) EscrowValidate(dbTx bool) error {
	var (
		nodeRegistration model.NodeRegistration
		row              *sql.Row
		err              error
	)
	if tx.Escrow.GetApproverAddress() == "" {
		return blocker.NewBlocker(blocker.ValidationErr, "ApproverAddressRequired")
	}
	if tx.Escrow.GetCommission() <= 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "CommissionNotEnough")
	}

	// check for duplication
	row, err = tx.QueryExecutor.ExecuteSelectRow(
		tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(),
		dbTx,
		tx.Body.NodePublicKey,
	)
	if err != nil {
		return err
	}
	err = tx.NodeRegistrationQuery.Scan(&nodeRegistration, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "NodeNotRegistered")
	}

	// sender must be node owner
	if tx.SenderAddress != nodeRegistration.GetAccountAddress() {
		return blocker.NewBlocker(blocker.ValidationErr, "AccountNotNodeOwner")
	}
	if nodeRegistration.GetRegistrationStatus() == uint32(model.NodeRegistrationState_NodeDeleted) {
		return blocker.NewBlocker(blocker.ValidationErr, "NodeAlreadyDeleted")
	}
	return nil
}

/*
EscrowApplyUnconfirmed is func that for applying to unconfirmed Transaction `RemoveNodeRegistration` type.
Perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *RemoveNodeRegistration) EscrowApplyUnconfirmed() error {

	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-(tx.Fee + tx.Escrow.GetCommission()),
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

/*
EscrowUndoApplyUnconfirmed func that perform on apply confirm preparation
*/
func (tx *RemoveNodeRegistration) EscrowUndoApplyUnconfirmed() error {
	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		tx.Fee+tx.Escrow.GetCommission(),
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

// EscrowApplyConfirmed method for confirmed the transaction and store into database
func (tx *RemoveNodeRegistration) EscrowApplyConfirmed(blockTimestamp int64) error {

	var (
		nodeRegistration model.NodeRegistration
		queries          [][]interface{}
		row              *sql.Row
		err              error
	)

	row, err = tx.QueryExecutor.ExecuteSelectRow(
		tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(),
		false,
		tx.Body.NodePublicKey,
	)
	if err != nil {
		return err
	}
	err = tx.NodeRegistrationQuery.Scan(&nodeRegistration, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.AppErr, "NodeNotRegistered")
	}

	// sender balance by refunding the locked balance
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		nodeRegistration.GetLockedBalance()-(tx.Fee+tx.Escrow.GetCommission()),
		map[string]interface{}{
			"account_address": tx.Escrow.GetSenderAddress(),
			"block_height":    tx.Height,
		},
	)
	queries = append(queries, accountBalanceSenderQ...)

	// sender ledger
	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.Escrow.GetSenderAddress(),
		BalanceChange:  nodeRegistration.GetLockedBalance() - (tx.Fee + tx.Escrow.GetCommission()),
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventRemoveNodeRegistrationTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	senderAccountLedgerArgs = append([]interface{}{senderAccountLedgerQ}, senderAccountLedgerArgs...)
	queries = append(queries, senderAccountLedgerArgs)

	// Insert Escrow
	escrowArgs := tx.EscrowQuery.InsertEscrowTransaction(tx.Escrow)
	queries = append(queries, escrowArgs...)

	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}

	return nil
}

/*
EscrowApproval handle approval an escrow transaction, execute tasks that was skipped when escrow pending.
like: spreading commission and fee, and also more pending tasks
*/
func (tx *RemoveNodeRegistration) EscrowApproval(
	blockTimestamp int64,
	txBody *model.ApprovalEscrowTransactionBody,
) error {
	var (
		nodeRegistration model.NodeRegistration
		queries          [][]interface{}
		escrow           = tx.Escrow
		row              *sql.Row
		err              error
	)

	switch txBody.GetApproval() {
	case model.EscrowApproval_Reject:
		escrow.Status = model.EscrowStatus_Rejected
		// give back sender balance
		accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
			nodeRegistration.GetLockedBalance()-(tx.Fee+escrow.GetCommission()),
			map[string]interface{}{
				"account_address": escrow.GetSenderAddress(),
				"block_height":    tx.Height,
			},
		)
		queries = append(queries, accountBalanceSenderQ...)

		// sender ledger
		senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
			AccountAddress: escrow.GetSenderAddress(),
			BalanceChange:  nodeRegistration.GetLockedBalance() - (tx.Fee + escrow.GetCommission()),
			TransactionID:  tx.ID,
			BlockHeight:    tx.Height,
			EventType:      model.EventType_EventRemoveNodeRegistrationTransaction,
			Timestamp:      uint64(blockTimestamp),
		})
		senderAccountLedgerArgs = append([]interface{}{senderAccountLedgerQ}, senderAccountLedgerArgs...)
		queries = append(queries, senderAccountLedgerArgs)
	default:
		escrow.Status = model.EscrowStatus_Approved
		row, err = tx.QueryExecutor.ExecuteSelectRow(
			tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(),
			false,
			tx.Body.NodePublicKey,
		)
		if err != nil {
			return err
		}
		return blocker.NewBlocker(blocker.AppErr, "NodeNotRegistered")
	}

	// Rebuild node registration
	nodeRegistration.Height = tx.Height
	nodeRegistration.LockedBalance = 0
	// nodeRegistration.NodeAddress = nil
	nodeRegistration.NodePublicKey = tx.Body.GetNodePublicKey()
	nodeRegistration.Latest = true
	nodeRegistration.RegistrationStatus = uint32(model.NodeRegistrationState_NodeDeleted)

	// Node registration
	queries = append(queries, tx.NodeRegistrationQuery.UpdateNodeRegistration(&nodeRegistration)...)

		// Rebuild node registration
		nodeRegistration.Height = tx.Height
		nodeRegistration.LockedBalance = 0
		nodeRegistration.NodeAddress = nil
		nodeRegistration.NodePublicKey = tx.Body.GetNodePublicKey()
		nodeRegistration.Latest = true
		nodeRegistration.RegistrationStatus = uint32(model.NodeRegistrationState_NodeDeleted)

		// Node registration
		queries = append(queries, tx.NodeRegistrationQuery.UpdateNodeRegistration(&nodeRegistration)...)

		// approver balance
		approverBalanceQ := tx.AccountBalanceQuery.AddAccountBalance(
			escrow.GetCommission(),
			map[string]interface{}{
				"account_address": escrow.GetApproverAddress(),
				"block_height":    tx.Height,
			},
		)
		queries = append(queries, approverBalanceQ...)
		// approver ledger
		approverLedgerQ, approverLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
			AccountAddress: escrow.GetApproverAddress(),
			BalanceChange:  escrow.GetCommission(),
			BlockHeight:    tx.Height,
			TransactionID:  tx.ID,
			Timestamp:      uint64(blockTimestamp),
			EventType:      model.EventType_EventRemoveNodeRegistrationTransaction,
		})
		approverLedgerArgs = append([]interface{}{approverLedgerQ}, approverLedgerArgs...)
		queries = append(queries, approverLedgerArgs)
	}

	// Insert Escrow
	escrowArgs := tx.EscrowQuery.InsertEscrowTransaction(escrow)
	queries = append(queries, escrowArgs...)

	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}

	return nil
}
