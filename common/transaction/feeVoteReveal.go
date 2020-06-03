package transaction

import (
	"database/sql"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	FeeVoteRevealTransaction struct {
		ID                     int64
		Fee                    int64
		SenderAddress          string
		Height                 uint32
		Timestamp              int64
		Body                   *model.FeeVoteRevealTransactionBody
		AccountBalanceQuery    query.AccountBalanceQueryInterface
		NodeRegistrationQuery  query.NodeRegistrationQueryInterface
		FeeVoteRevealVoteQuery query.FeeVoteRevealVoteQueryInterface
		FeeVoteCommitVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		AccountBalanceHelper   AccountBalanceHelperInterface
		AccountLedgerHelper    AccountLedgerHelperInterface
		QueryExecutor          query.ExecutorInterface
	}
)

// Validate for validating the transaction concerned
func (tx *FeeVoteRevealTransaction) Validate(dbTx bool) error {
	var (
		err        error
		row        *sql.Row
		qry        string
		args       []interface{}
		nodeReg    model.NodeRegistration
		revealVote model.FeeVoteRevealVote
	)

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
	// check duplicated reveal to database, once per node owner
	qry, args = tx.FeeVoteRevealVoteQuery.GetFeeVoteRevealByAccountAddress(tx.SenderAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return err
	}
	err = tx.FeeVoteRevealVoteQuery.Scan(&revealVote, row)
	if err == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "DuplicatedFeeVoteReveal")
	} else {
		if err != sql.ErrNoRows {
			return err
		}
		// check the transaction submitted on reveal-phase

		// VoteObject.Signature must be a valid signature from node-owner on bytes(VoteInfo)
		// VoteObject.RecentBlockHash must be within the timeframe of current voting period.
		// must match the previously submitted in commit_votes table || commit_vote transaction before
		// qry, args = tx.FeeVoteCommitVoteQuery.

	}
	return nil
}
