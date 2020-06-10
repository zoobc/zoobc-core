package transaction

import (
	"bytes"
	"database/sql"
	"strings"

	"golang.org/x/crypto/sha3"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	FeeVoteRevealTransaction struct {
		ID                     int64
		Fee                    int64
		SenderAddress          string
		Height                 uint32
		Timestamp              int64
		Body                   *model.FeeVoteRevealTransactionBody
		FeeScaleService        fee.FeeScaleServiceInterface
		SignatureInterface     crypto.SignatureInterface
		BlockQuery             query.BlockQueryInterface
		NodeRegistrationQuery  query.NodeRegistrationQueryInterface
		FeeVoteCommitVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		FeeVoteRevealVoteQuery query.FeeVoteRevealVoteQueryInterface
		AccountBalanceHelper   AccountBalanceHelperInterface
		AccountLedgerHelper    AccountLedgerHelperInterface
		QueryExecutor          query.ExecutorInterface
	}
)

// Validate for validating the transaction concerned
func (tx *FeeVoteRevealTransaction) Validate(dbTx bool) error {
	var (
		accountBalance model.AccountBalance
		feeVotePhase   model.FeeVotePhase
		recentBlock    model.Block
		commitVote     model.FeeVoteCommitmentVote
		nodeReg        model.NodeRegistration
		lastFeeScale   model.FeeScale
		args           []interface{}
		row            *sql.Row
		qry            string
		err            error
	)

	// check the transaction submitted on reveal-phase
	feeVotePhase, _, err = tx.FeeScaleService.GetCurrentPhase(tx.Timestamp, true)
	if err != nil {
		return err
	}
	if feeVotePhase != model.FeeVotePhase_FeeVotePhaseReveal {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidPhasePeriod")
	}

	// get last fee scale height
	err = tx.FeeScaleService.GetLatestFeeScale(&lastFeeScale)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	// must match the previously submitted in CommitmentVote
	qry, args = tx.FeeVoteCommitVoteQuery.GetVoteCommitByAccountAddressAndHeight(
		tx.SenderAddress,
		lastFeeScale.BlockHeight,
	)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return err
	}
	err = tx.FeeVoteCommitVoteQuery.Scan(&commitVote, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return blocker.NewBlocker(blocker.ValidationErr, "CommitVoteNotFound")
		}
		return err
	}

	digest := sha3.New256()
	_, err = digest.Write(tx.GetFeeVoteInfoBytes())
	if err != nil {
		return err
	}

	if res := bytes.Compare(commitVote.GetVoteHash(), digest.Sum([]byte{})); res != 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "NotMatchVoteHashed")
	}

	// VoteObject.Signature must be a valid signature from node-owner on bytes(VoteInfo)
	err = tx.SignatureInterface.VerifySignature(
		tx.GetFeeVoteInfoBytes(),
		tx.Body.GetVoterSignature(),
		tx.SenderAddress,
	)
	if err != nil {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidSignature")
	}
	row, err = tx.QueryExecutor.ExecuteSelectRow(
		tx.BlockQuery.GetBlockByHeight(tx.Body.GetFeeVoteInfo().GetRecentBlockHeight()),
		dbTx,
	)
	if err != nil {
		return err
	}
	err = tx.BlockQuery.Scan(&recentBlock, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "BlockNotFound")
	}
	if res := bytes.Compare(tx.Body.GetFeeVoteInfo().GetRecentBlockHash(), recentBlock.GetBlockHash()); res != 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidRecentBlock")
	}

	// sender must be as node owner
	qry, args = tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.SenderAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return err
	}
	err = tx.NodeRegistrationQuery.Scan(&nodeReg, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "SenderAccountNotNodeOwner")
	}

	// check duplicated reveal to database, once per node owner per period
	err = tx.checkDuplicateVoteReveal(dbTx)
	if err != nil {
		return err
	}
	// check account balance sender
	err = tx.AccountBalanceHelper.GetBalanceByAccountID(&accountBalance, tx.SenderAddress, dbTx)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "AccountBalanceNotFound")
	}
	if accountBalance.GetSpendableBalance() < tx.Fee {
		return blocker.NewBlocker(blocker.ValidationErr, "BalanceNotEnough")
	}
	return nil
}

func (tx *FeeVoteRevealTransaction) checkDuplicateVoteReveal(dbTx bool) error {
	var (
		revealVote model.FeeVoteRevealVote
		qry, args  = tx.FeeVoteRevealVoteQuery.GetFeeVoteRevealByAccountAddressAndRecentBlockHeight(
			tx.SenderAddress,
			tx.Body.GetFeeVoteInfo().GetRecentBlockHeight(),
		)
		row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	)
	if err != nil {
		return err
	}
	err = tx.FeeVoteRevealVoteQuery.Scan(&revealVote, row)
	if err != nil {
		// it means don't have previous vote
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	return blocker.NewBlocker(blocker.ValidationErr, "DuplicatedFeeVoteReveal")
}

// ApplyUnconfirmed to apply unconfirmed transaction
func (tx *FeeVoteRevealTransaction) ApplyUnconfirmed() error {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, -tx.Fee)
}

/*
UndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *FeeVoteRevealTransaction) UndoApplyUnconfirmed() error {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, tx.Fee)
}

// ApplyConfirmed applying transaction, will store ledger, account balance update, and also the transaction it self
func (tx *FeeVoteRevealTransaction) ApplyConfirmed(blockTimestamp int64) error {
	var (
		err error
	)

	err = tx.AccountBalanceHelper.AddAccountBalance(tx.SenderAddress, -tx.Fee, tx.Height)
	if err != nil {
		return err
	}

	err = tx.AccountLedgerHelper.InsertLedgerEntry(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  -tx.Fee,
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventFeeVoteRevealTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	if err != nil {
		return err
	}
	qry, args := tx.FeeVoteRevealVoteQuery.InsertRevealVote(&model.FeeVoteRevealVote{
		VoteInfo:       tx.Body.GetFeeVoteInfo(),
		VoterSignature: tx.Body.GetVoterSignature(),
		VoterAddress:   tx.SenderAddress,
		BlockHeight:    tx.Height,
	})
	err = tx.QueryExecutor.ExecuteTransaction(qry, args...)
	if err != nil {
		return err
	}
	return nil
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (*FeeVoteRevealTransaction) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	var (
		buff    = bytes.NewBuffer(txBodyBytes)
		chunked []byte
		err     error
	)

	chunked, err = util.ReadTransactionBytes(buff, int(constant.RecentBlockHashLength))
	if err != nil {
		return nil, err
	}
	recentBlockHas, err := util.ReadTransactionBytes(buff, int(util.ConvertBytesToUint32(chunked)))
	if err != nil {
		return nil, err
	}

	chunked, err = util.ReadTransactionBytes(buff, int(constant.RecentBlockHeight))
	if err != nil {
		return nil, err
	}
	recentBlockHeight := util.ConvertBytesToUint32(chunked)

	chunked, err = util.ReadTransactionBytes(buff, int(constant.FeeVote))
	if err != nil {
		return nil, err
	}
	feeVote := util.ConvertBytesToUint64(chunked)

	chunked, err = util.ReadTransactionBytes(buff, int(constant.VoterSignatureLength))
	if err != nil {
		return nil, err
	}
	voterSignature, err := util.ReadTransactionBytes(buff, int(util.ConvertBytesToUint32(chunked)))
	if err != nil {
		return nil, err
	}
	return &model.FeeVoteRevealTransactionBody{
		FeeVoteInfo: &model.FeeVoteInfo{
			RecentBlockHash:   recentBlockHas,
			RecentBlockHeight: recentBlockHeight,
			FeeVote:           int64(feeVote),
		},
		VoterSignature: voterSignature,
	}, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *FeeVoteRevealTransaction) GetBodyBytes() []byte {
	buff := bytes.NewBuffer([]byte{})
	buff.Write(util.ConvertUint32ToBytes(uint32(len(tx.Body.FeeVoteInfo.RecentBlockHash))))
	buff.Write(tx.Body.FeeVoteInfo.RecentBlockHash)
	buff.Write(util.ConvertUint32ToBytes(tx.Body.FeeVoteInfo.RecentBlockHeight))
	buff.Write(util.ConvertUint64ToBytes(uint64(tx.Body.FeeVoteInfo.FeeVote)))
	buff.Write(util.ConvertUint32ToBytes(uint32(len(tx.Body.VoterSignature))))
	buff.Write(tx.Body.VoterSignature)
	return buff.Bytes()
}

// GetTransactionBody append isTransaction_TransactionBody oneOf
func (tx *FeeVoteRevealTransaction) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_FeeVoteRevealTransactionBody{
		FeeVoteRevealTransactionBody: tx.Body,
	}
}

// GetFeeVoteInfoBytes will build bytes from model.FeeVoteInfo
func (tx *FeeVoteRevealTransaction) GetFeeVoteInfoBytes() []byte {
	buff := bytes.NewBuffer([]byte{})
	buff.Write(util.ConvertUint32ToBytes(uint32(len(tx.Body.FeeVoteInfo.RecentBlockHash))))
	buff.Write(tx.Body.FeeVoteInfo.RecentBlockHash)
	buff.Write(util.ConvertUint32ToBytes(tx.Body.FeeVoteInfo.RecentBlockHeight))
	buff.Write(util.ConvertUint64ToBytes(uint64(tx.Body.FeeVoteInfo.FeeVote)))
	return buff.Bytes()
}

// Escrowable will check the transaction is escrow or not. Curently doesn't have ecrow option
func (*FeeVoteRevealTransaction) Escrowable() (EscrowTypeAction, bool) {
	return nil, false
}

// GetAmount return Amount from TransactionBody
func (tx *FeeVoteRevealTransaction) GetAmount() int64 {
	return 0
}

// GetMinimumFee calculate fee
func (tx *FeeVoteRevealTransaction) GetMinimumFee() (int64, error) {
	return 0, nil
}

// SkipMempoolTransaction this tx type has no mempool filter
func (tx *FeeVoteRevealTransaction) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	blockTimestamp int64,
) (bool, error) {
	// check tx is still valid for reveal vote phase based on new block timestamp
	var feeVotePhase, _, err = tx.FeeScaleService.GetCurrentPhase(blockTimestamp, true)
	if err != nil {
		return true, err
	}
	if feeVotePhase != model.FeeVotePhase_FeeVotePhaseReveal {
		return true, nil
	}
	// check duplicate vote on mempool
	for _, selectedTx := range selectedTransactions {
		// if we find another fee reveal tx in currently selected transactions, filter current one out of selection
		sameTxType := model.TransactionType_FeeVoteRevealVoteTransaction == model.TransactionType(selectedTx.GetTransactionType())
		if sameTxType && tx.SenderAddress == selectedTx.SenderAccountAddress {
			return true, nil
		}
	}
	// check previouds vote
	err = tx.checkDuplicateVoteReveal(false)
	if err != nil {
		if strings.Contains(err.Error(), string(blocker.ValidationErr)) {
			return true, nil
		}
		return true, err
	}
	return false, nil
}

// GetSize send money Amount should be 8
func (tx *FeeVoteRevealTransaction) GetSize() uint32 {
	// only amount
	return uint32(len(tx.GetBodyBytes()))
}
