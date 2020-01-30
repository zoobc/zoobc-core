package transaction

import (
	"bytes"
	"database/sql"

	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

// UpdateNodeRegistration Implement service layer for (new) node registration's transaction
type UpdateNodeRegistration struct {
	ID                    int64
	Fee                   int64
	SenderAddress         string
	Height                uint32
	Timestamp             int64
	Body                  *model.UpdateNodeRegistrationTransactionBody
	Escrow                *model.Escrow
	AccountBalanceQuery   query.AccountBalanceQueryInterface
	NodeRegistrationQuery query.NodeRegistrationQueryInterface
	BlockQuery            query.BlockQueryInterface
	QueryExecutor         query.ExecutorInterface
	AuthPoown             auth.ProofOfOwnershipValidationInterface
	AccountLedgerQuery    query.AccountLedgerQueryInterface
	EscrowQuery           query.EscrowTransactionQueryInterface
}

// SkipMempoolTransaction filter out of the mempool a node registration tx if there are other node registration tx in mempool
// to make sure only one node registration tx at the time (the one with highest fee paid) makes it to the same block
func (tx *UpdateNodeRegistration) SkipMempoolTransaction(selectedTransactions []*model.Transaction) (bool, error) {
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
func (tx *UpdateNodeRegistration) ApplyConfirmed(blockTimestamp int64) error {
	var (
		nodeQueries          [][]interface{}
		prevNodeRegistration *model.NodeRegistration
		lockedBalance        int64
		nodeAddress          *model.NodeAddress
		nodePublicKey        []byte
	)
	// get the latest node registration by owner (sender account)
	qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.SenderAddress)
	rows, err := tx.QueryExecutor.ExecuteSelect(qry, false, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	nr, err := tx.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if (err != nil) || len(nr) == 0 {
		return blocker.NewBlocker(blocker.AppErr, "NodeNotFoundWithAccountAddress")
	}
	prevNodeRegistration = nr[0]

	if tx.Body.LockedBalance > 0 {
		lockedBalance = tx.Body.LockedBalance
	} else {
		lockedBalance = prevNodeRegistration.LockedBalance
	}

	if tx.Body.NodeAddress != nil {
		nodeAddress = tx.Body.GetNodeAddress()
	} else {
		nodeAddress = prevNodeRegistration.GetNodeAddress()
	}

	if len(tx.Body.NodePublicKey) != 0 {
		nodePublicKey = tx.Body.NodePublicKey
	} else {
		nodePublicKey = prevNodeRegistration.NodePublicKey
	}
	nodeRegistration := &model.NodeRegistration{
		NodeID:             prevNodeRegistration.NodeID,
		LockedBalance:      lockedBalance,
		Height:             tx.Height,
		NodeAddress:        nodeAddress,
		RegistrationHeight: prevNodeRegistration.RegistrationHeight,
		NodePublicKey:      nodePublicKey,
		Latest:             true,
		RegistrationStatus: prevNodeRegistration.RegistrationStatus,
		// account address is the only field that can't be updated via update node registration
		AccountAddress: prevNodeRegistration.AccountAddress,
	}

	var effectiveBalanceToLock int64
	if tx.Body.LockedBalance > 0 {
		// delta amount to be locked
		effectiveBalanceToLock = tx.Body.LockedBalance - prevNodeRegistration.LockedBalance
	}

	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		-(effectiveBalanceToLock + tx.Fee),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)

	nodeQueries = tx.NodeRegistrationQuery.UpdateNodeRegistration(nodeRegistration)
	queries := append(accountBalanceSenderQ, nodeQueries...)

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

	return nil
}

/*
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `UpdateNodeRegistration` type:
	- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *UpdateNodeRegistration) ApplyUnconfirmed() error {

	var (
		err                    error
		prevNodeRegistration   *model.NodeRegistration
		effectiveBalanceToLock int64
	)

	// update sender balance by reducing his spendable balance of the tx fee + new balance to be lock
	// (delta between old locked balance and updatee locked balance)
	if tx.Body.LockedBalance > 0 {
		// get the latest noderegistration by owner (sender account)
		qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.SenderAddress)
		rows, err := tx.QueryExecutor.ExecuteSelect(qry, false, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		nr, err := tx.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
		if (err != nil) || len(nr) == 0 {
			return blocker.NewBlocker(blocker.AppErr, "NodeNotFoundWithAccountAddress")
		}
		prevNodeRegistration = nr[0]

		// delta amount to be locked
		effectiveBalanceToLock = tx.Body.LockedBalance - prevNodeRegistration.LockedBalance
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

	// validate node public key, if we are updating that field
	// note: node pub key must be not already registered for another node
	if len(tx.Body.NodePublicKey) > 0 && !bytes.Equal(prevNodeRegistration.NodePublicKey, tx.Body.NodePublicKey) {
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
	}

	// delta amount to be locked
	var effectiveBalanceToLock = tx.Body.LockedBalance - prevNodeRegistration.LockedBalance
	if effectiveBalanceToLock < 0 {
		// cannot lock less than what previously locked
		return blocker.NewBlocker(blocker.ValidationErr, "LockedBalanceLessThenPreviouslyLocked")
	}

	// check balance
	qry, args = tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
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

	nodeAddress := tx.Body.GetNodeAddress()
	if nodeAddress == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "NodeAddressEmpty")
	}

	return nil
}

func (tx *UpdateNodeRegistration) GetAmount() int64 {
	return tx.Body.LockedBalance
}

func (tx *UpdateNodeRegistration) GetSize() uint32 {
	// note: the first 4 bytes (uint32) of nodeAddress contain the field length
	// (necessary to parse the bytes into tx body struct)
	nodeAddress := constant.NodeAddressLength + uint32(len([]byte(
		tx.NodeRegistrationQuery.ExtractNodeAddress(tx.Body.GetNodeAddress()),
	)))
	// ProofOfOwnership (message + signature)
	poown := util.GetProofOfOwnershipSize(true)
	return constant.NodePublicKey + constant.Balance + poown + nodeAddress
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (tx *UpdateNodeRegistration) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {

	var (
		nodePublicKey []byte
		lockedBalance uint64
		nodeAddress   *model.NodeAddress
		poown         *model.ProofOfOwnership
		err           error
	)
	// read body bytes
	buffer := bytes.NewBuffer(txBodyBytes)

	nodePublicKey, err = util.ReadTransactionBytes(buffer, int(constant.NodePublicKey))
	if err != nil {
		return nil, err
	}

	nodeAddressLengthBytes, err := util.ReadTransactionBytes(buffer, int(constant.NodeAddressLength))
	if err != nil {
		return nil, err
	}
	nodeAddressLength := util.ConvertBytesToUint32(nodeAddressLengthBytes)             // uint32 length of next bytes to read
	nodeAddressBytes, err := util.ReadTransactionBytes(buffer, int(nodeAddressLength)) // based on nodeAddressLength
	if err != nil {
		return nil, err
	}
	nodeAddress = tx.NodeRegistrationQuery.BuildNodeAddress(string(nodeAddressBytes))

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
		NodeAddress:   nodeAddress,
		LockedBalance: int64(lockedBalance),
		Poown:         poown,
	}, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *UpdateNodeRegistration) GetBodyBytes() []byte {

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(tx.Body.NodePublicKey)
	// note: the first 4 bytes (uint32) of nodeAddress contain the field length
	// (necessary to parse the bytes into tx body struct)
	fullNodeAddress := tx.NodeRegistrationQuery.ExtractNodeAddress(tx.Body.GetNodeAddress())
	addressLengthBytes := util.ConvertUint32ToBytes(uint32(len([]byte(
		fullNodeAddress,
	))))
	buffer.Write(addressLengthBytes)
	buffer.Write([]byte(fullNodeAddress))
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
	if tx.Escrow != nil {
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
func (tx *UpdateNodeRegistration) EscrowValidate(dbTx bool) error {
	var (
		nodeRegistration, tempNodeRegistration model.NodeRegistration
		effectiveBalanceToLock                 int64
		accountBalance                         model.AccountBalance
		row                                    *sql.Row
		err                                    error
	)
	// formally validate tx body fields
	if tx.Body.Poown == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "PoownRequired")
	}
	if tx.Escrow.GetApproverAddress() == "" {
		return blocker.NewBlocker(blocker.ValidationErr, "ApproverAddressRequired")
	}
	if tx.Escrow.GetCommission() <= 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "CommissionNotEnough")
	}

	if tx.Body.GetNodeAddress() == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "NodeAddressEmpty")
	}

	// validate proof of ownership
	err = tx.AuthPoown.ValidateProofOfOwnership(
		tx.Body.Poown, tx.Body.GetNodePublicKey(),
		tx.QueryExecutor,
		tx.BlockQuery,
	)
	if err != nil {
		return err
	}

	// check that sender is node's owner
	qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.SenderAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return err
	}
	err = tx.NodeRegistrationQuery.Scan(&nodeRegistration, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}

		return blocker.NewBlocker(blocker.ValidationErr, "SenderAccountNotNodeOwner")
	}

	if nodeRegistration.GetRegistrationStatus() == uint32(model.NodeRegistrationState_NodeDeleted) {
		return blocker.NewBlocker(blocker.AuthErr, "NodeDeleted")

	}

	// validate node public key, if we are updating that field
	// note: node pub key must be not already registered for another node
	if len(tx.Body.GetNodePublicKey()) > 0 && !bytes.Equal(nodeRegistration.GetNodePublicKey(), tx.Body.GetNodePublicKey()) {
		row, err = tx.QueryExecutor.ExecuteSelectRow(
			tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(),
			false,
			tx.Body.GetNodePublicKey(),
		)
		if err != nil {
			return err
		}
		err = tx.NodeRegistrationQuery.Scan(&tempNodeRegistration, row)
		if err != nil {
			if err != sql.ErrNoRows {
				return err

			}
			return blocker.NewBlocker(blocker.ValidationErr, "NodePublicKeyAlreadyRegistered")
		}

	}

	// delta amount to be locked
	effectiveBalanceToLock = tx.Body.LockedBalance - nodeRegistration.GetLockedBalance()
	if effectiveBalanceToLock < 0 {
		// cannot lock less than what previously locked
		return blocker.NewBlocker(blocker.ValidationErr, "LockedBalanceLessThenPreviouslyLocked")
	}

	// check balance
	qry, args = tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return err
	}
	err = tx.AccountBalanceQuery.Scan(&accountBalance, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "SenderAccountAddressNotFound")
	}

	if accountBalance.GetSpendableBalance() < tx.Fee+effectiveBalanceToLock+tx.Escrow.GetCommission() {
		return blocker.NewBlocker(blocker.ValidationErr, "UserBalanceNotEnough")
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

	// skip when not pending
	if tx.Escrow.GetStatus() != model.EscrowStatus_Pending {
		return nil
	}

	// update sender balance by reducing his spendable balance of the tx fee + new balance to be lock
	// (delta between old locked balance and update locked balance)
	if tx.Body.LockedBalance > 0 {
		// get the latest node registration by owner (sender account)
		qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.SenderAddress)
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

	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-(effectiveBalanceToLock + tx.Fee + tx.Escrow.GetCommission()),
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

	// skip when status not pending
	if tx.Escrow.GetStatus() != model.EscrowStatus_Pending {
		return nil
	}

	// get the latest node registration by owner (sender account)
	qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.SenderAddress)
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

	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		effectiveBalanceToLock+tx.Fee+tx.Escrow.GetCommission(),
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

// EscrowApplyConfirmed method for confirmed the transaction and store into database
func (tx *UpdateNodeRegistration) EscrowApplyConfirmed(blockTimestamp int64) error {
	var (
		effectiveBalanceToLock int64
		nodeRegistration       model.NodeRegistration
		queries                [][]interface{}
		row                    *sql.Row
		err                    error
	)

	// skip when status not pending
	if tx.Escrow.GetStatus() != model.EscrowStatus_Pending {
		return nil
	}

	// get the latest node registration by owner (sender account)
	qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.SenderAddress)
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

	// Rebuild node registration
	if tx.Body.GetLockedBalance() > 0 {
		nodeRegistration.LockedBalance = tx.Body.GetLockedBalance()
		// delta amount to be locked
		effectiveBalanceToLock = tx.Body.GetLockedBalance() - nodeRegistration.GetLockedBalance()
	}

	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		-(effectiveBalanceToLock+tx.Fee)+tx.Escrow.GetCommission(),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)
	queries = append(queries, accountBalanceSenderQ...)

	// sender account ledger log
	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  -(effectiveBalanceToLock + tx.Fee + tx.Escrow.GetCommission()),
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

	return nil
}

/*
EscrowApproval handle approval an escrow transaction, execute tasks that was skipped when escrow pending.
like: spreading commission and fee, and also more pending tasks
*/
func (tx *UpdateNodeRegistration) EscrowApproval(blockTimestamp int64) error {
	var (
		nodeRegistration model.NodeRegistration
		queries          [][]interface{}
		err              error
		row              *sql.Row
	)

	// get the latest node registration by owner (sender account)
	qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.SenderAddress)
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

	// Rebuild node registration
	if tx.Body.GetLockedBalance() > 0 {
		nodeRegistration.LockedBalance = tx.Body.GetLockedBalance()

	}

	if tx.Body.GetNodeAddress() != nil {
		nodeRegistration.NodeAddress = tx.Body.GetNodeAddress()
	}

	if len(tx.Body.GetNodePublicKey()) != 0 {
		nodeRegistration.NodePublicKey = tx.Body.GetNodePublicKey()
	}
	nodeRegistration.Height = tx.Height
	nodeRegistration.Latest = true

	// Node registration Query
	queries = append(
		queries,
		tx.NodeRegistrationQuery.UpdateNodeRegistration(&nodeRegistration)...,
	)

	// approver
	approverBalanceQ := tx.AccountBalanceQuery.AddAccountBalance(
		tx.Escrow.GetCommission(),
		map[string]interface{}{
			"account_address": tx.Escrow.GetApproverAddress(),
			"block_height":    tx.Height,
		},
	)
	queries = append(queries, approverBalanceQ...)

	// approver account ledger log
	approverLedgerQ, approverLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.Escrow.GetApproverAddress(),
		BalanceChange:  tx.Escrow.GetCommission(),
		BlockHeight:    tx.Height,
		TransactionID:  tx.ID,
		Timestamp:      uint64(blockTimestamp),
		EventType:      model.EventType_EventUpdateNodeRegistrationTransaction,
	})
	approverLedgerArgs = append([]interface{}{approverLedgerQ}, approverLedgerArgs...)
	queries = append(queries, approverLedgerArgs)

	// Insert Escrow
	escrowArgs := tx.EscrowQuery.InsertEscrowTransaction(tx.Escrow)
	queries = append(queries, escrowArgs...)

	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}
	return nil
}
