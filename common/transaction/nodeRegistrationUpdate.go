package transaction

import (
	"bytes"
	"net"
	"net/url"

	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"

	"github.com/zoobc/zoobc-core/common/model"
)

// UpdateNodeRegistration Implement service layer for (new) node registration's transaction
type UpdateNodeRegistration struct {
	Body                  *model.UpdateNodeRegistrationTransactionBody
	Fee                   int64
	SenderAddress         string
	Height                uint32
	AccountBalanceQuery   query.AccountBalanceQueryInterface
	NodeRegistrationQuery query.NodeRegistrationQueryInterface
	BlockQuery            query.BlockQueryInterface
	QueryExecutor         query.ExecutorInterface
	AuthPoown             auth.ProofOfOwnershipValidationInterface
}

func (tx *UpdateNodeRegistration) ApplyConfirmed() error {
	var (
		nodeQueries          [][]interface{}
		prevNodeRegistration *model.NodeRegistration
	)
	// get the latest noderegistration by owner (sender account)
	qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.SenderAddress)
	rows, err := tx.QueryExecutor.ExecuteSelect(qry, false, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	if nr := tx.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows); len(nr) > 0 {
		prevNodeRegistration = nr[0]
	} else {
		return blocker.NewBlocker(blocker.AppErr, "NodeNotFoundWithAccountAddress")
	}

	var lockedBalance int64
	if tx.Body.LockedBalance > 0 {
		lockedBalance = tx.Body.LockedBalance
	} else {
		lockedBalance = prevNodeRegistration.LockedBalance
	}
	var nodeAddress string
	if tx.Body.NodeAddress != "" {
		nodeAddress = tx.Body.NodeAddress
	} else {
		nodeAddress = prevNodeRegistration.NodeAddress
	}
	var nodePublicKey []byte
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
		Queued:             prevNodeRegistration.Queued,
		AccountAddress:     prevNodeRegistration.AccountAddress,
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
		err                  error
		prevNodeRegistration *model.NodeRegistration
	)

	// update sender balance by reducing his spendable balance of the tx fee + new balance to be lock
	// (delta between old locked balance and updatee locked balance)
	var effectiveBalanceToLock int64
	if tx.Body.LockedBalance > 0 {
		// get the latest noderegistration by owner (sender account)
		qry, args := tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.SenderAddress)
		rows, err := tx.QueryExecutor.ExecuteSelect(qry, false, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		if nr := tx.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows); len(nr) > 0 {
			prevNodeRegistration = nr[0]
		} else {
			return blocker.NewBlocker(blocker.AppErr, "NodeNotFoundWithAccountAddress")
		}
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
	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		tx.Body.LockedBalance+tx.Fee,
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
func (tx *UpdateNodeRegistration) Validate(dbTx bool) error {
	var (
		accountBalance       model.AccountBalance
		prevNodeRegistration model.NodeRegistration
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
	if !rows.Next() {
		// sender doesn't own any node
		// note: any account can own exactly one node at the time, meaning that, if this query returns no rows,
		return blocker.NewBlocker(blocker.ValidationErr, "SenderAccountNotNodeOwner")
	}
	_ = rows.Scan(
		&prevNodeRegistration.NodeID,
		&prevNodeRegistration.NodePublicKey,
		&prevNodeRegistration.AccountAddress,
		&prevNodeRegistration.RegistrationHeight,
		&prevNodeRegistration.NodeAddress,
		&prevNodeRegistration.LockedBalance,
		&prevNodeRegistration.Queued,
		&prevNodeRegistration.Latest,
		&prevNodeRegistration.Height)
	defer rows.Close()

	// validate node public key, if we are updating that field
	// note: node pub key must be not already registered for another node
	if len(tx.Body.NodePublicKey) > 0 && !bytes.Equal(prevNodeRegistration.NodePublicKey, tx.Body.NodePublicKey) {
		rows, err := tx.QueryExecutor.ExecuteSelect(tx.NodeRegistrationQuery.
			GetNodeRegistrationByNodePublicKey(), false, tx.Body.NodePublicKey)
		if err != nil {
			return err
		}
		if rows.Next() {
			// public key already registered
			return blocker.NewBlocker(blocker.ValidationErr, "NodePublicKeyAlredyRegistered")
		}
	}

	if tx.Body.LockedBalance > 0 {
		// delta amount to be locked
		effectiveBalanceToLock := tx.Body.LockedBalance - prevNodeRegistration.LockedBalance
		if effectiveBalanceToLock < 0 {
			// cannot lock less than what previously locked
			return blocker.NewBlocker(blocker.ValidationErr, "LockedBalanceLessThenPreviouslyLocked")
		}

		// check balance
		qry, args := tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
		rows, err := tx.QueryExecutor.ExecuteSelect(qry, dbTx, args...)
		if err != nil {
			return blocker.NewBlocker(blocker.DBErr, err.Error())
		} else if rows.Next() {
			_ = rows.Scan(
				&accountBalance.AccountAddress,
				&accountBalance.BlockHeight,
				&accountBalance.SpendableBalance,
				&accountBalance.Balance,
				&accountBalance.PopRevenue,
				&accountBalance.Latest,
			)
		}
		defer rows.Close()

		if accountBalance.SpendableBalance < tx.Fee+effectiveBalanceToLock {
			return blocker.NewBlocker(blocker.ValidationErr, "UserBalanceNotEnough")
		}
		// TODO: check minimum amount to be locked (at current height the min amount is = 0, but in future may change)
	} else if accountBalance.SpendableBalance < tx.Fee {
		return blocker.NewBlocker(blocker.ValidationErr, "UserBalanceNotEnough")
	}

	if tx.Body.NodeAddress != "" {
		_, err := url.ParseRequestURI(tx.Body.NodeAddress)
		if err != nil {
			if net.ParseIP(tx.Body.NodeAddress) == nil {
				return blocker.NewBlocker(blocker.ValidationErr, "InvalidAddress")
			}
		}
	}

	return nil
}

func (tx *UpdateNodeRegistration) GetAmount() int64 {
	return tx.Body.LockedBalance
}

func (tx *UpdateNodeRegistration) GetSize() uint32 {
	// note: the first 4 bytes (uint32) of nodeAddress contain the field length
	// (necessary to parse the bytes into tx body struct)
	nodeAddress := constant.NodeAddressLength + uint32(len([]byte(tx.Body.NodeAddress)))
	// ProofOfOwnership (message + signature)
	poown := util.GetProofOfOwnershipSize(true)
	return constant.NodePublicKey + constant.Balance + poown + nodeAddress
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (*UpdateNodeRegistration) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	// read body bytes
	buffer := bytes.NewBuffer(txBodyBytes)
	nodePublicKey, err := util.ReadTransactionBytes(buffer, int(constant.NodePublicKey))
	if err != nil {
		return nil, err
	}
	nodeAddressLengthBytes, err := util.ReadTransactionBytes(buffer, int(constant.NodeAddressLength))
	if err != nil {
		return nil, err
	}
	nodeAddressLength := util.ConvertBytesToUint32(nodeAddressLengthBytes)        // uint32 length of next bytes to read
	nodeAddress, err := util.ReadTransactionBytes(buffer, int(nodeAddressLength)) // based on nodeAddressLength
	if err != nil {
		return nil, err
	}
	lockedBalanceBytes, err := util.ReadTransactionBytes(buffer, int(constant.Balance))
	if err != nil {
		return nil, err
	}
	lockedBalance := util.ConvertBytesToUint64(lockedBalanceBytes)
	// parse ProofOfOwnership (message + signature) bytes
	poownBytes, err := util.ReadTransactionBytes(buffer, int(util.GetProofOfOwnershipSize(true)))
	if err != nil {
		return nil, err
	}
	poown, err := util.ParseProofOfOwnershipBytes(poownBytes)
	if err != nil {
		return nil, err
	}
	return &model.UpdateNodeRegistrationTransactionBody{
		NodePublicKey: nodePublicKey,
		NodeAddress:   string(nodeAddress),
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
	addressLengthBytes := util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.NodeAddress))))
	buffer.Write(addressLengthBytes)
	buffer.Write([]byte(tx.Body.NodeAddress))
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.LockedBalance)))
	// convert ProofOfOwnership (message + signature) to bytes
	buffer.Write(util.GetProofOfOwnershipBytes(tx.Body.Poown))
	return buffer.Bytes()
}
