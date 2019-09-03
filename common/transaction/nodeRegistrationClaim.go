package transaction

import (
	"bytes"

	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"

	"github.com/zoobc/zoobc-core/common/model"
)

// ClaimNodeRegistration Implement service layer for claim node registration's transaction
type ClaimNodeRegistration struct {
	Body                  *model.ClaimNodeRegistrationTransactionBody
	Fee                   int64
	SenderAddress         string
	Height                uint32
	AccountBalanceQuery   query.AccountBalanceQueryInterface
	NodeRegistrationQuery query.NodeRegistrationQueryInterface
	BlockQuery            query.BlockQueryInterface
	QueryExecutor         query.ExecutorInterface
	AuthPoown             auth.ProofOfOwnershipValidationInterface
}

func (tx *ClaimNodeRegistration) ApplyConfirmed() error {
	var (
		queries              [][]interface{}
		prevNodeRegistration *model.NodeRegistration
	)

	if tx.Height > 0 {
		err := tx.UndoApplyUnconfirmed()
		if err != nil {
			return err
		}
	}

	qry1, args1 := tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(tx.Body.NodePublicKey)
	rows, err := tx.QueryExecutor.ExecuteSelect(qry1, args1)
	if err != nil {
		return err
	}
	defer rows.Close()
	if nr := tx.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows); len(nr) > 0 {
		prevNodeRegistration = nr[0]
	} else {
		return blocker.NewBlocker(blocker.AppErr, "NodePublicKeyNotRegistered")
	}

	nodeRegistration := &model.NodeRegistration{
		NodeID:             prevNodeRegistration.NodeID,
		NodePublicKey:      tx.Body.NodePublicKey,
		AccountAddress:     tx.Body.AccountAddress,
		LockedBalance:      prevNodeRegistration.LockedBalance,
		Height:             tx.Height,
		NodeAddress:        prevNodeRegistration.NodeAddress,
		RegistrationHeight: prevNodeRegistration.RegistrationHeight,
		Latest:             true,
		Queued:             prevNodeRegistration.Queued,
	}

	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		-(tx.Fee),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)
	updateNodeQ, updateNodeArg := tx.NodeRegistrationQuery.UpdateNodeRegistration(nodeRegistration)
	queries = append(append([][]interface{}{}, accountBalanceSenderQ...),
		append([]interface{}{updateNodeQ}, updateNodeArg...),
	)
	// update node_registry entry
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
func (tx *ClaimNodeRegistration) Validate() error {
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

	// check that sender is node's owner
	if tx.Body.AccountAddress == "" {
		return blocker.NewBlocker(blocker.ValidationErr, "AccountAddressRequired")
	}
	qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.Body.AccountAddress)
	rows, err := tx.QueryExecutor.ExecuteSelect(qry, args)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		// account address has already an active node registration (either queued or not)
		return blocker.NewBlocker(blocker.ValidationErr, "AccountAlreadyNodeOwner")
	}

	qry1, args1 := tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(tx.Body.NodePublicKey)
	rows, err = tx.QueryExecutor.ExecuteSelect(qry1, args1)
	if err != nil {
		return err
	}
	if !rows.Next() {
		// public key must be already registered
		return blocker.NewBlocker(blocker.ValidationErr, "NodePublicKeyNotRegistered")
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
func (*ClaimNodeRegistration) ParseBodyBytes(txBodyBytes []byte) model.TransactionBodyInterface {
	buffer := bytes.NewBuffer(txBodyBytes)
	nodePublicKey := buffer.Next(int(constant.NodePublicKey))
	accountAddress := buffer.Next(int(constant.AccountAddress))
	// parse ProofOfOwnership (message + signature) bytes
	poown := util.ParseProofOfOwnershipBytes(buffer.Next(int(util.GetProofOfOwnershipSize(true))))
	return &model.ClaimNodeRegistrationTransactionBody{
		NodePublicKey:  nodePublicKey,
		AccountAddress: string(accountAddress),
		Poown:          poown,
	}
}

// GetBodyBytes translate tx body to bytes representation
func (tx *ClaimNodeRegistration) GetBodyBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(tx.Body.NodePublicKey)
	buffer.Write([]byte(tx.Body.AccountAddress))
	// convert ProofOfOwnership (message + signature) to bytes
	buffer.Write(util.GetProofOfOwnershipBytes(tx.Body.Poown))
	return buffer.Bytes()
}
