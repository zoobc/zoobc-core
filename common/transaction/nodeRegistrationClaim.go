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

// ClaimNodeRegistration Implement service layer for claim node registration's transaction
type ClaimNodeRegistration struct {
	ID                    int64
	Fee                   int64
	SenderAddress         string
	Height                uint32
	Body                  *model.ClaimNodeRegistrationTransactionBody
	Escrow                *model.Escrow
	AccountBalanceQuery   query.AccountBalanceQueryInterface
	NodeRegistrationQuery query.NodeRegistrationQueryInterface
	BlockQuery            query.BlockQueryInterface
	QueryExecutor         query.ExecutorInterface
	AuthPoown             auth.NodeAuthValidationInterface
	AccountLedgerQuery    query.AccountLedgerQueryInterface
	EscrowQuery           query.EscrowTransactionQueryInterface
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
		if _, ok := authorizedType[model.TransactionType(sel.GetTransactionType())]; ok && tx.SenderAddress == sel.SenderAccountAddress {
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

	// update sender balance by claiming the locked balance
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
		NodePublicKey:      tx.Body.NodePublicKey,
		Latest:             true,
		RegistrationStatus: uint32(model.NodeRegistrationState_NodeDeleted),
		// We can't just set accountAddress to an empty string,
		// otherwise it could trigger an error when parsing the transaction from its bytes
		AccountAddress: nodeReg.GetAccountAddress(),
	})
	queries = append(queries, nodeQueries...)

	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  nodeReg.GetLockedBalance() - tx.Fee,
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventClaimNodeRegistrationTransaction,
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
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `ClaimNodeRegistration` type:
	- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *ClaimNodeRegistration) ApplyUnconfirmed() error {

	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-(tx.Fee),
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

func (tx *ClaimNodeRegistration) UndoApplyUnconfirmed() error {
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
func (tx *ClaimNodeRegistration) Validate(dbTx bool) error {
	var (
		nodeRegistrations []*model.NodeRegistration
		accountBalance    model.AccountBalance
	)

	// validate proof of ownership
	if tx.Body.Poown == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "PoownRequired")
	}
	if err := tx.AuthPoown.ValidateProofOfOwnership(
		tx.Body.Poown, tx.Body.NodePublicKey,
		tx.QueryExecutor,
		tx.BlockQuery); err != nil {
		return err
	}

	rows2, err := tx.QueryExecutor.ExecuteSelect(tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(), false, tx.Body.NodePublicKey)
	if err != nil {
		return err
	}
	defer rows2.Close()
	// cannot claim a deleted node
	nodeRegistrations, err = tx.NodeRegistrationQuery.BuildModel(nodeRegistrations, rows2)
	if (len(nodeRegistrations) == 0) || (err != nil) {
		// public key must be already registered
		return blocker.NewBlocker(blocker.ValidationErr, "NodePublicKeyNotRegistered")
	}
	if nodeRegistrations[0].RegistrationStatus == uint32(model.NodeRegistrationState_NodeDeleted) {
		return blocker.NewBlocker(blocker.ValidationErr, "NodeAlreadyClaimedOrDeleted")
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

func (tx *ClaimNodeRegistration) GetAmount() int64 {
	return 0
}

func (*ClaimNodeRegistration) GetMinimumFee() (int64, error) {
	return 0, nil
}

func (*ClaimNodeRegistration) GetSize() uint32 {
	// ProofOfOwnership (message + signature)
	poown := util.GetProofOfOwnershipSize(true)
	return constant.AccountAddress + constant.NodePublicKey + poown
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (tx *ClaimNodeRegistration) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	// read body bytes todo: add accountAddressLength to body
	buffer := bytes.NewBuffer(txBodyBytes)
	nodePublicKey, err := util.ReadTransactionBytes(buffer, int(constant.NodePublicKey))
	if err != nil {
		return nil, err
	}
	// parse ProofOfOwnership (message + signature) bytes
	poown, err := util.ParseProofOfOwnershipBytes(buffer.Next(int(util.GetProofOfOwnershipSize(true))))
	if err != nil {
		return nil, err
	}
	return &model.ClaimNodeRegistrationTransactionBody{
		NodePublicKey: nodePublicKey,
		Poown:         poown,
	}, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *ClaimNodeRegistration) GetBodyBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(tx.Body.NodePublicKey)
	// convert ProofOfOwnership (message + signature) to bytes
	buffer.Write(util.GetProofOfOwnershipBytes(tx.Body.Poown))
	return buffer.Bytes()
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
func (tx *ClaimNodeRegistration) EscrowValidate(bool) error {
	var (
		nodeRegistration model.NodeRegistration
		row              *sql.Row
		err              error
	)

	// validate proof of ownership
	if tx.Body.Poown == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "PoownRequired")
	}
	if tx.Escrow.GetApproverAddress() == "" {
		return blocker.NewBlocker(blocker.ValidationErr, "ApproverAddressRequired")
	}
	if tx.Escrow.GetCommission() <= 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "CommissionNotEnough")
	}

	err = tx.AuthPoown.ValidateProofOfOwnership(
		tx.Body.Poown, tx.Body.NodePublicKey,
		tx.QueryExecutor,
		tx.BlockQuery,
	)
	if err != nil {
		return err
	}

	row, err = tx.QueryExecutor.ExecuteSelectRow(
		tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(),
		false,
		tx.Body.NodePublicKey,
	)
	if err != nil {
		return err
	}

	// cannot claim a deleted node
	err = tx.NodeRegistrationQuery.Scan(&nodeRegistration, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		// public key must be already registered
		return blocker.NewBlocker(blocker.ValidationErr, "NodePublicKeyNotRegistered")
	}

	if nodeRegistration.RegistrationStatus == uint32(model.NodeRegistrationState_NodeDeleted) {
		return blocker.NewBlocker(blocker.ValidationErr, "NodeAlreadyClaimedOrDeleted")
	}

	return nil
}

/*
EscrowApplyUnconfirmed is func that for applying to unconfirmed Transaction `ClaimNodeRegistration` type:
	- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *ClaimNodeRegistration) EscrowApplyUnconfirmed() error {

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
func (tx *ClaimNodeRegistration) EscrowUndoApplyUnconfirmed() error {
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

/*
EscrowApplyConfirmed func that for applying pending escrow transaction.
*/
func (tx *ClaimNodeRegistration) EscrowApplyConfirmed(blockTimestamp int64) error {
	var (
		prevNodeRegistration *model.NodeRegistration
		err                  error
		row                  *sql.Row
		queries              [][]interface{}
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

	// tag the node as deleted
	// update sender balance by claiming the locked balance
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		prevNodeRegistration.LockedBalance-(tx.Fee+tx.Escrow.GetCommission()),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)
	queries = append(queries, accountBalanceSenderQ...)
	// sender Account Ledger
	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  prevNodeRegistration.LockedBalance - (tx.Fee + tx.Escrow.GetCommission()),
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventClaimNodeRegistrationTransaction,
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
func (tx *ClaimNodeRegistration) EscrowApproval(
	blockTimestamp int64,
	txBody *model.ApprovalEscrowTransactionBody,
) error {
	var (
		prevNodeRegistration model.NodeRegistration
		queries              [][]interface{}
		escrow               = tx.Escrow
		row                  *sql.Row
		err                  error
	)

	switch txBody.GetApproval() {
	case model.EscrowApproval_Reject:
		escrow.Status = model.EscrowStatus_Rejected
	default:
		escrow.Status = model.EscrowStatus_Approved
		// approver balance
		approverBalanceQ := tx.AccountBalanceQuery.AddAccountBalance(
			tx.Escrow.GetCommission(),
			map[string]interface{}{
				"account_address": tx.Escrow.GetApproverAddress(),
				"block_height":    tx.Height,
			},
		)
		queries = append(queries, approverBalanceQ...)

		// approver account ledger log
		approverAccountLedgerQ, approverAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
			AccountAddress: tx.Escrow.GetApproverAddress(),
			BalanceChange:  tx.Escrow.GetCommission(),
			BlockHeight:    tx.Height,
			TransactionID:  tx.ID,
			Timestamp:      uint64(blockTimestamp),
			EventType:      model.EventType_EventClaimNodeRegistrationTransaction,
		})
		approverAccountLedgerArgs = append([]interface{}{approverAccountLedgerQ}, approverAccountLedgerArgs...)
		queries = append(queries, approverAccountLedgerArgs)

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
		// Node registration
		nodeQueries := tx.NodeRegistrationQuery.UpdateNodeRegistration(&model.NodeRegistration{
			NodeID:        prevNodeRegistration.GetNodeID(),
			LockedBalance: 0,
			Height:        tx.Height,
			// NodeAddress:        nil,
			RegistrationHeight: prevNodeRegistration.GetRegistrationHeight(),
			NodePublicKey:      tx.Body.NodePublicKey,
			Latest:             true,
			RegistrationStatus: uint32(model.NodeRegistrationState_NodeDeleted),
			// We can't just set accountAddress to an empty string,
			// otherwise it could trigger an error when parsing the transaction from its bytes
			AccountAddress: prevNodeRegistration.GetAccountAddress(),
		})
		queries = append(queries, nodeQueries...)
	}
	// Node registration
	nodeQueries := tx.NodeRegistrationQuery.UpdateNodeRegistration(&model.NodeRegistration{
		NodeID:        prevNodeRegistration.GetNodeID(),
		LockedBalance: 0,
		Height:        tx.Height,
		// NodeAddress:        nil,
		RegistrationHeight: prevNodeRegistration.GetRegistrationHeight(),
		NodePublicKey:      tx.Body.NodePublicKey,
		Latest:             true,
		RegistrationStatus: uint32(model.NodeRegistrationState_NodeDeleted),
		// We can't just set accountAddress to an empty string,
		// otherwise it could trigger an error when parsing the transaction from its bytes
		AccountAddress: prevNodeRegistration.GetAccountAddress(),
	})
	queries = append(queries, nodeQueries...)

	// Insert Escrow
	escrowArgs := tx.EscrowQuery.InsertEscrowTransaction(tx.Escrow)
	queries = append(queries, escrowArgs...)

	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}

	return nil
}
