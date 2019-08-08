package transaction

import (
	"bytes"
	"errors"
	"net"
	"net/url"

	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"

	"github.com/zoobc/zoobc-core/common/model"
)

// UpdateNodeRegistration Implement service layer for (new) node registration's transaction
type UpdateNodeRegistration struct {
	Body                  *model.UpdateNodeRegistrationTransactionBody
	Fee                   int64
	SenderAddress         string
	SenderAccountType     uint32
	Height                uint32
	AccountBalanceQuery   query.AccountBalanceQueryInterface
	AccountQuery          query.AccountQueryInterface
	NodeRegistrationQuery query.NodeRegistrationQueryInterface
	QueryExecutor         query.ExecutorInterface
}

func (tx *UpdateNodeRegistration) ApplyConfirmed() error {
	var (
		queries              [][]interface{}
		prevNodeRegistration *model.NodeRegistration
	)
	if err := tx.Validate(); err != nil {
		return err
	}

	if tx.Height > 0 {
		err := tx.UndoApplyUnconfirmed()
		if err != nil {
			return err
		}
	}
	// get the latest noderegistration by owner (sender account)
	senderID := util.CreateAccountIDFromAddress(tx.SenderAccountType, tx.SenderAddress)
	rows, err := tx.QueryExecutor.ExecuteSelect(tx.NodeRegistrationQuery.GetNodeRegistrationByAccountID(senderID))
	if err != nil {
		// no nodes registered with this accountID
		return err
	} else if rows.Next() {
		nr := tx.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
		prevNodeRegistration = nr[0]
	}
	defer rows.Close()
	if prevNodeRegistration == nil {
		return errors.New("NodeNotFoundWithAccountID")
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
		AccountId:          senderID,
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
			"account_id": util.CreateAccountIDFromAddress(
				tx.SenderAccountType,
				tx.SenderAddress,
			),
			"block_height": tx.Height,
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
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `UpdateNodeRegistration` type:
	- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *UpdateNodeRegistration) ApplyUnconfirmed() error {

	var (
		err                  error
		prevNodeRegistration *model.NodeRegistration
	)

	if err := tx.Validate(); err != nil {
		return err
	}

	// update sender balance by reducing his spendable balance of the tx fee
	var effectiveBalanceToLock int64
	if tx.Body.LockedBalance > 0 {
		// get the latest noderegistration by owner (sender account)
		senderID := util.CreateAccountIDFromAddress(tx.SenderAccountType, tx.SenderAddress)
		rows, err := tx.QueryExecutor.ExecuteSelect(tx.NodeRegistrationQuery.GetNodeRegistrationByAccountID(senderID))
		if err != nil {
			return err
		} else if rows.Next() {
			nr := tx.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
			prevNodeRegistration = nr[0]
		}
		defer rows.Close()
		if prevNodeRegistration == nil {
			return errors.New("NodeNotFoundWithAccountID")
		}
		// delta amount to be locked
		effectiveBalanceToLock = tx.Body.LockedBalance - prevNodeRegistration.LockedBalance
	}

	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-(effectiveBalanceToLock + tx.Fee),
		map[string]interface{}{
			"account_id": util.CreateAccountIDFromAddress(
				tx.SenderAccountType,
				tx.SenderAddress,
			),
		},
	)
	// add row to node_registry table
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
			"account_id": util.CreateAccountIDFromAddress(
				tx.SenderAccountType,
				tx.SenderAddress,
			),
		},
	)
	err := tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	if err != nil {
		return err
	}
	return nil
}

// Validate validate node registration transaction and tx body
func (tx *UpdateNodeRegistration) Validate() error {
	var (
		accountBalance       model.AccountBalance
		prevNodeRegistration model.NodeRegistration
	)
	// formally validate tx body fields
	if tx.Body.Poown == nil {
		return errors.New("PoownRequired")
	}
	// TODO: validate poown when implemented
	//
	//

	// check that sender is node's owner
	senderID := util.CreateAccountIDFromAddress(tx.SenderAccountType, tx.SenderAddress)
	rows, err := tx.QueryExecutor.ExecuteSelect(tx.NodeRegistrationQuery.GetNodeRegistrationByAccountID(senderID))
	if err != nil {
		return err
	}
	if !rows.Next() {
		// sender doesn't own any node
		// note: any account can own exactly one node at the time, meaning that, if this query returns no rows,
		return errors.New("NodeNotFoundWithAccountID")
	}
	_ = rows.Scan(
		&prevNodeRegistration.NodeID,
		&prevNodeRegistration.NodePublicKey,
		&prevNodeRegistration.AccountId,
		&prevNodeRegistration.RegistrationHeight,
		&prevNodeRegistration.NodeAddress,
		&prevNodeRegistration.LockedBalance,
		&prevNodeRegistration.Queued,
		&prevNodeRegistration.Latest,
		&prevNodeRegistration.Height)
	defer rows.Close()

	// validate node public key, if we are updating that field
	// note: node pub key must be not already registered
	if len(tx.Body.NodePublicKey) == 32 {
		rows, err := tx.QueryExecutor.ExecuteSelect(tx.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(tx.Body.NodePublicKey))
		if err != nil {
			return err
		}
		if rows.Next() {
			// public key already registered
			return errors.New("NodePublicKeyAlredyRegistered")
		}
	}

	rows, err = tx.QueryExecutor.ExecuteSelect(tx.AccountBalanceQuery.GetAccountBalanceByAccountID(senderID))

	if err != nil {
		return err
	} else if rows.Next() {
		_ = rows.Scan(
			&accountBalance.AccountID,
			&accountBalance.BlockHeight,
			&accountBalance.SpendableBalance,
			&accountBalance.Balance,
			&accountBalance.PopRevenue,
			&accountBalance.Latest,
		)
	}
	defer rows.Close()

	if tx.Body.LockedBalance > 0 {
		// delta amount to be locked
		effectiveBalanceToLock := tx.Body.LockedBalance - prevNodeRegistration.LockedBalance
		if effectiveBalanceToLock < 0 {
			// cannot lock less than what previously locked
			return errors.New("LockedBalanceLessThenPreviouslyLocked")
		}
		if accountBalance.SpendableBalance < tx.Fee+effectiveBalanceToLock {
			return errors.New("UserBalanceNotEnough")
		}
		// TODO: check minimum amount to be locked (at current height the min amount is = 0, but in future may change)
	} else if accountBalance.SpendableBalance < tx.Fee {
		return errors.New("UserBalanceNotEnough")
	}

	if tx.Body.NodeAddress != "" {
		if tx.Body.NodeAddressLength == 0 {
			return errors.New("NodeAddressLengthRequired")
		}
		if len(tx.Body.NodeAddress) != int(tx.Body.NodeAddressLength) {
			return errors.New("InvalidNodeAddressLength")
		}
		if net.ParseIP(tx.Body.NodeAddress) == nil {
			// not a valid ipv4 or ipv6 address. let's check if is a valid domain name
			if _, err := url.Parse(tx.Body.NodeAddress); err != nil {
				return errors.New("InvalidAddress")
			}
		}
	}

	return nil
}

func (tx *UpdateNodeRegistration) GetAmount() int64 {
	return tx.Body.LockedBalance
}

func (tx *UpdateNodeRegistration) GetSize() uint32 {
	nodePublicKey := 32
	nodeAddressLength := 1
	nodeAddress := tx.Body.NodeAddressLength
	lockedBalance := 8
	//TODO: return bytes of ProofOfOwnership (message + signature) when implemented
	poown := 256
	return uint32(nodePublicKey+nodeAddressLength+lockedBalance+poown) + nodeAddress + nodeAddress
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (*UpdateNodeRegistration) ParseBodyBytes(txBodyBytes []byte) *model.UpdateNodeRegistrationTransactionBody {
	buffer := bytes.NewBuffer(txBodyBytes)
	nodePublicKey := buffer.Next(32)
	nodeAddressLength := util.ConvertBytesToUint32([]byte{buffer.Next(1)[0], 0, 0, 0}) // uint32 length of next bytes to read
	nodeAddress := buffer.Next(int(nodeAddressLength))                                 // based on nodeAddressLength
	lockedBalance := util.ConvertBytesToUint64(buffer.Next(8))
	//TODO: parse ProofOfOwnership (message + signature) bytes when implemented
	poown := new(model.ProofOfOwnership)
	return &model.UpdateNodeRegistrationTransactionBody{
		NodePublicKey:     nodePublicKey,
		NodeAddressLength: nodeAddressLength,
		NodeAddress:       string(nodeAddress),
		LockedBalance:     int64(lockedBalance),
		Poown:             poown,
	}
}

// GetBodyBytes translate tx body to bytes representation
func (*UpdateNodeRegistration) GetBodyBytes(txBody *model.UpdateNodeRegistrationTransactionBody) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(txBody.NodePublicKey)
	addressLengthBytes := util.ConvertUint32ToBytes(txBody.NodeAddressLength)
	buffer.Write([]byte{addressLengthBytes[0]})
	buffer.Write([]byte(txBody.NodeAddress))
	buffer.Write(util.ConvertUint64ToBytes(uint64(txBody.LockedBalance)))
	//TODO: convert ProofOfOwnership (message + signature) to bytes
	buffer.Write([]byte{})
	return buffer.Bytes()
}
