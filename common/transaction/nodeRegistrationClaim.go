package transaction

import (
	"bytes"

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
	AuthPoown             auth.ProofOfOwnershipValidationInterface
	AccountLedgerQuery    query.AccountLedgerQueryInterface
	EscrowQuery           query.EscrowTransactionQueryInterface
}

// SkipMempoolTransaction filter out of the mempool a node registration tx if there are other node registration tx in mempool
// to make sure only one node registration tx at the time (the one with highest fee paid) makes it to the same block
func (tx *ClaimNodeRegistration) SkipMempoolTransaction(selectedTransactions []*model.Transaction) (bool, error) {
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
		nodeQueries          [][]interface{}
		prevNodeRegistration *model.NodeRegistration
	)

	rows, err := tx.QueryExecutor.ExecuteSelect(tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(), false, tx.Body.NodePublicKey)
	if err != nil {
		return err
	}
	defer rows.Close()

	nr, err := tx.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if (err != nil) || (len(nr) == 0) {
		return blocker.NewBlocker(blocker.AppErr, "NodePublicKeyNotRegistered")
	}
	prevNodeRegistration = nr[0]

	// tag the node as deleted
	nodeRegistration := &model.NodeRegistration{
		NodeID:             prevNodeRegistration.NodeID,
		LockedBalance:      0,
		Height:             tx.Height,
		NodeAddress:        nil,
		RegistrationHeight: prevNodeRegistration.RegistrationHeight,
		NodePublicKey:      tx.Body.NodePublicKey,
		Latest:             true,
		RegistrationStatus: uint32(model.NodeRegistrationState_NodeDeleted),
		// We can't just set accountAddress to an empty string,
		// otherwise it could trigger an error when parsing the transaction from its bytes
		AccountAddress: prevNodeRegistration.AccountAddress,
	}
	// update sender balance by claiming the locked balance
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		prevNodeRegistration.LockedBalance-tx.Fee,
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)
	nodeQueries = tx.NodeRegistrationQuery.UpdateNodeRegistration(nodeRegistration)
	queries := append(accountBalanceSenderQ, nodeQueries...)

	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  prevNodeRegistration.LockedBalance - tx.Fee,
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

	return nil
}

func (tx *ClaimNodeRegistration) GetAmount() int64 {
	return 0
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

	return nil, false
}

// EscrowValidate validate node registration transaction and tx body
func (tx *ClaimNodeRegistration) EscrowValidate(bool) error {

	return nil
}

/*
EscrowUndoApplyUnconfirmed func that perform on apply confirm preparation
*/
func (tx *ClaimNodeRegistration) EscrowUndoApplyUnconfirmed() error {

	return nil
}

/*
EscrowApplyConfirmed func that for applying pending escrow transaction.
*/
func (tx *ClaimNodeRegistration) EscrowApplyConfirmed(blockTimestamp int64) error {

	return nil
}

/*
EscrowApproval handle approval an escrow transaction, execute tasks that was skipped when escrow pending.
like: spreading commission and fee, and also more pending tasks
*/
func (tx *ClaimNodeRegistration) EscrowApproval(int64, *model.ApprovalEscrowTransactionBody) error {
	return nil
}
