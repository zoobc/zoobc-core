package transaction

import (
	"bytes"
	"database/sql"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

// CreateBlockchainObjectTransaction is Transaction Type that implemented TypeAction
type CreateBlockchainObjectTransaction struct {
	ID                            int64
	Fee                           int64
	SenderAddress                 []byte
	Height                        uint32
	TransactionHash               []byte
	Body                          *model.CreateBlockchainObjectTransactionBody
	Escrow                        *model.Escrow
	AccountBalanceHelper          AccountBalanceHelperInterface
	QueryExecutor                 query.ExecutorInterface
	EscrowQuery                   query.EscrowTransactionQueryInterface
	BlockchainObjectQuery         query.BlockchainObjectQueryInterface
	BlockchainObjectPropertyQuery query.BlockchainObjectPropertyQueryInterface
	AccountDatasetQuery           query.AccountDatasetQueryInterface
	EscrowFee                     fee.FeeModelInterface
	NormalFee                     fee.FeeModelInterface
}

// ApplyConfirmed func that for applying Transaction CreateBlockchainObjectTransaction type.
func (tx *CreateBlockchainObjectTransaction) ApplyConfirmed(blockTimestamp int64) error {
	var err error
	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.SenderAddress,
		-(tx.Fee + tx.Body.BlockchainObjectBalance),
		model.EventType_EventCreateBlockchainObjectTransaction,
		tx.Height,
		tx.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}
	// add new account address (blockchain Object ID) on account balance
	var blockchainObjectID = make([]byte, constant.AccountAddressTypeLength+uint32(sha3.New256().Size()))
	blockchainObjectID = append(tx.SenderAddress[:4], tx.TransactionHash...)
	err = tx.AccountBalanceHelper.AddAccountBalance(
		blockchainObjectID,
		tx.Body.BlockchainObjectBalance,
		model.EventType_EventCreateBlockchainObjectTransaction,
		tx.Height,
		tx.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}
	// insert blockchain object
	blockchainObject := &model.BlockchainObject{
		ID:                  blockchainObjectID,
		OwnerAccountAddress: tx.SenderAddress,
		BlockHeight:         tx.Height,
	}
	var qry, args = tx.BlockchainObjectQuery.InsertBlockcahinObject(blockchainObject)
	err = tx.QueryExecutor.ExecuteTransaction(qry, args)
	if err != nil {
		return err
	}
	// insert immutable properties
	var boImmutableProperties []*model.BlockchainObjectProperty
	for key := range tx.Body.BlockchainObjectImmutableProperties {
		boImmutableProperties = append(boImmutableProperties, &model.BlockchainObjectProperty{
			BlockchainObjectID: blockchainObjectID,
			Key:                key,
			Value:              tx.Body.BlockchainObjectImmutableProperties[key],
			BlockHeight:        tx.Height,
		})
	}
	qry, args = tx.BlockchainObjectPropertyQuery.InsertBlockcahinObjectProperties(boImmutableProperties)
	err = tx.QueryExecutor.ExecuteTransaction(qry, args)
	if err != nil {
		return err
	}
	// insert mutable properties
	if len(tx.Body.BlockchainObjectMutableProperties) > 0 {
		var boMutableProperties []*model.AccountDataset
		// SetterAccountAddress should be blockchainObjectID to make blockchain object can change it's property
		for key := range tx.Body.BlockchainObjectMutableProperties {
			boMutableProperties = append(boMutableProperties, &model.AccountDataset{
				SetterAccountAddress:    blockchainObjectID,
				RecipientAccountAddress: blockchainObjectID,
				Property:                tx.Body.BlockchainObjectMutableProperties[key],
				Value:                   key,
				Height:                  tx.Height,
				IsActive:                true,
				Latest:                  true,
			})
		}
		qry, args = tx.AccountDatasetQuery.InsertAccountDatasets(boMutableProperties)
		err = tx.QueryExecutor.ExecuteTransaction(qry, args)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `CreateBlockchainObjectTransaction` type
*/
func (tx *CreateBlockchainObjectTransaction) ApplyUnconfirmed() error {
	// update sender balance by reducing his spendable balance of the tx fee
	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, -(tx.Body.BlockchainObjectBalance + tx.Fee))
	if err != nil {
		return err
	}
	return nil
}

/*
UndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *CreateBlockchainObjectTransaction) UndoApplyUnconfirmed() error {
	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, tx.Body.BlockchainObjectBalance+tx.Fee)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return nil
}

/*
Validate is func that for validating to Transaction CreateBlockchainObjectTransaction type
*/
func (tx *CreateBlockchainObjectTransaction) Validate(dbTx bool) error {
	if tx.Body.GetBlockchainObjectBalance() <= 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "BalanceMustMoreThan0")
	}
	if len(tx.TransactionHash) != sha3.New256().Size() {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidTransactionHash")
	}
	if len(tx.Body.BlockchainObjectImmutableProperties) == 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "ShouldHaveImmutablePropertyAtLeastOne")
	}
	// check existing account
	var (
		accountBalance     model.AccountBalance
		blockchainObjectID = append(tx.SenderAddress[:4], tx.TransactionHash...)
		err                = tx.AccountBalanceHelper.GetBalanceByAccountAddress(&accountBalance, blockchainObjectID, dbTx)
	)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}
	return nil
}

func (tx *CreateBlockchainObjectTransaction) GetMinimumFee() (int64, error) {
	if tx.Escrow != nil && tx.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.Escrow.GetApproverAddress(), []byte{}) {
		return tx.EscrowFee.CalculateTxMinimumFee(tx.Body, tx.Escrow)
	}
	return tx.NormalFee.CalculateTxMinimumFee(tx.Body, tx.Escrow)
}

func (tx *CreateBlockchainObjectTransaction) GetAmount() int64 {
	return tx.Body.BlockchainObjectBalance
}

func (tx *CreateBlockchainObjectTransaction) GetSize() (uint32, error) {
	txBodyBytes, err := tx.GetBodyBytes()
	if err != nil {
		return 0, err
	}
	return uint32(len(txBodyBytes)), nil
}

func (tx *CreateBlockchainObjectTransaction) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	var (
		buffer       = bytes.NewBuffer(txBodyBytes)
		dataLength   uint32
		chunkedBytes []byte
		txBody       model.CreateBlockchainObjectTransactionBody
		err          error
	)

	// get balance of blockchain object
	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.Balance))
	if err != nil {
		return nil, err
	}
	txBody.BlockchainObjectBalance = int64(util.ConvertBytesToUint64(chunkedBytes))
	// get number of immutable properties
	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.BlockchainObjectNumberOfProperties))
	if err != nil {
		return nil, err
	}
	var (
		totalProperty       = util.ConvertBytesToUint32(chunkedBytes)
		immutableProperties = make(map[string]string, totalProperty)
		key, value          string
	)
	for i := 0; uint32(i) < totalProperty; i++ {
		// get key of property
		chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.BlockchainObjectPropertyKeyLength))
		if err != nil {
			return nil, err
		}
		dataLength = util.ConvertBytesToUint32(chunkedBytes)
		chunkedBytes, err = util.ReadTransactionBytes(buffer, int(dataLength))
		if err != nil {
			return nil, err
		}
		key = string(chunkedBytes)

		// get value of property
		chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.BlockchainObjectPropertyValueLength))
		if err != nil {
			return nil, err
		}
		dataLength = util.ConvertBytesToUint32(chunkedBytes)
		chunkedBytes, err = util.ReadTransactionBytes(buffer, int(dataLength))
		if err != nil {
			return nil, err
		}
		value = string(chunkedBytes)
		immutableProperties[key] = value
	}
	txBody.BlockchainObjectImmutableProperties = immutableProperties

	// get number of mutable properties
	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.BlockchainObjectNumberOfProperties))
	if err != nil {
		return nil, err
	}
	totalProperty = util.ConvertBytesToUint32(chunkedBytes)
	if totalProperty > 0 {
		var mutableProperties = make(map[string]string)
		for i := 0; uint32(i) < totalProperty; i++ {
			// get key of property
			chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.BlockchainObjectPropertyKeyLength))
			if err != nil {
				return nil, err
			}
			dataLength = util.ConvertBytesToUint32(chunkedBytes)
			chunkedBytes, err = util.ReadTransactionBytes(buffer, int(dataLength))
			if err != nil {
				return nil, err
			}
			key = string(chunkedBytes)

			// get value of property
			chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.BlockchainObjectPropertyValueLength))
			if err != nil {
				return nil, err
			}
			dataLength = util.ConvertBytesToUint32(chunkedBytes)
			chunkedBytes, err = util.ReadTransactionBytes(buffer, int(dataLength))
			if err != nil {
				return nil, err
			}
			value = string(chunkedBytes)
			mutableProperties[key] = value
		}
		txBody.BlockchainObjectMutableProperties = mutableProperties
	}
	return &txBody, nil
}

func (tx *CreateBlockchainObjectTransaction) GetBodyBytes() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.BlockchainObjectBalance)))
	buffer.Write(util.ConvertUint32ToBytes(uint32(len(tx.Body.BlockchainObjectImmutableProperties))))
	for key := range tx.Body.BlockchainObjectImmutableProperties {
		buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(key)))))
		buffer.Write([]byte(key))
		buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.BlockchainObjectImmutableProperties[key])))))
		buffer.Write([]byte(tx.Body.BlockchainObjectImmutableProperties[key]))
	}
	buffer.Write(util.ConvertUint32ToBytes(uint32(len(tx.Body.BlockchainObjectMutableProperties))))
	for key := range tx.Body.BlockchainObjectMutableProperties {
		buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(key)))))
		buffer.Write([]byte(key))
		buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.BlockchainObjectMutableProperties[key])))))
		buffer.Write([]byte(tx.Body.BlockchainObjectMutableProperties[key]))
	}
	return buffer.Bytes(), nil
}

func (tx *CreateBlockchainObjectTransaction) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_CreateBlockchainObjectTransactionBody{
		CreateBlockchainObjectTransactionBody: tx.Body,
	}
}

func (tx *CreateBlockchainObjectTransaction) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	blockTimestamp int64,
	blockHeight uint32,
) (bool, error) {
	return false, nil
}

// Escrowable check if transaction type has escrow part and it will refill escrow part
func (tx *CreateBlockchainObjectTransaction) Escrowable() (EscrowTypeAction, bool) {
	if tx.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.Escrow.GetApproverAddress(), []byte{}) {
		tx.Escrow = &model.Escrow{
			ID:              tx.ID,
			SenderAddress:   tx.SenderAddress,
			ApproverAddress: tx.Escrow.GetApproverAddress(),
			Commission:      tx.Escrow.GetCommission(),
			Timeout:         tx.Escrow.GetTimeout(),
			Status:          tx.Escrow.GetStatus(),
			BlockHeight:     tx.Height,
			Latest:          true,
			Instruction:     tx.Escrow.GetInstruction(),
		}
		return EscrowTypeAction(tx), true
	}
	return nil, false
}

/*
EscrowApplyConfirmed is applyConfirmed specific for Escrow's transaction
*/
func (tx *CreateBlockchainObjectTransaction) EscrowApplyConfirmed(blockTimestamp int64) (err error) {
	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.SenderAddress,
		-(tx.Fee + tx.Escrow.GetCommission()),
		model.EventType_EventEscrowedTransaction,
		tx.Height,
		tx.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}
	return nil
}

func (tx *CreateBlockchainObjectTransaction) EscrowApplyUnconfirmed() (err error) {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, -(tx.Fee + tx.Escrow.GetCommission()))
}

func (tx *CreateBlockchainObjectTransaction) EscrowUndoApplyUnconfirmed() error {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, tx.Fee+tx.Escrow.GetCommission())
}

func (tx *CreateBlockchainObjectTransaction) EscrowValidate(dbTx bool) (err error) {
	if tx.Escrow.GetApproverAddress() == nil || bytes.Equal(tx.Escrow.GetApproverAddress(), []byte{}) {
		return blocker.NewBlocker(blocker.ValidationErr, "ApproverAddressRequired")
	}
	if tx.Escrow.GetCommission() <= 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "CommissionNotEnough")
	}
	if tx.Escrow.GetTimeout() > uint64(constant.MinRollbackBlocks) {
		return blocker.NewBlocker(blocker.ValidationErr, "TimeoutLimitExceeded")
	}
	err = tx.Validate(dbTx)
	if err != nil {
		return err
	}
	var enough bool
	enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTx, tx.SenderAddress, tx.Fee+tx.Escrow.GetCommission())
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

func (tx *CreateBlockchainObjectTransaction) EscrowApproval(
	blockTimestamp int64,
	txBody *model.ApprovalEscrowTransactionBody,
) (err error) {
	switch txBody.GetApproval() {
	case model.EscrowApproval_Approve:
		tx.Escrow.Status = model.EscrowStatus_Approved
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.SenderAddress,
			tx.Fee,
			model.EventType_EventEscrowedTransaction,
			tx.Height,
			tx.ID,
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
			tx.Escrow.GetApproverAddress(),
			tx.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.Height,
			tx.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	case model.EscrowApproval_Reject:
		tx.Escrow.Status = model.EscrowStatus_Rejected
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.Escrow.GetApproverAddress(),
			tx.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.Height,
			tx.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	default:
		tx.Escrow.Status = model.EscrowStatus_Expired
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.SenderAddress,
			tx.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.Height,
			tx.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	}
	escrowQ := tx.EscrowQuery.InsertEscrowTransaction(tx.Escrow)
	err = tx.QueryExecutor.ExecuteTransactions(escrowQ)
	if err != nil {
		return err
	}
	return nil
}
