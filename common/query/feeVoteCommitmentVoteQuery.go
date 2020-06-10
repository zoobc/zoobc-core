package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// FeeVoteCommitmentVoteQueryInterface interface that implemented by FeeVoteCommitmentVoteQuery
	FeeVoteCommitmentVoteQueryInterface interface {
		GetVoteCommitByAccountAddressAndHeight(
			accountAddress string,
			height uint32,
		) (qStr string, args []interface{})
		InsertCommitVote(voteCommit *model.FeeVoteCommitmentVote) (qStr string, args []interface{})
		ExtractModel(voteCommit *model.FeeVoteCommitmentVote) []interface{}
		Scan(voteCommit *model.FeeVoteCommitmentVote, row *sql.Row) error
		Rollback(height uint32) (multiQueries [][]interface{})
	}
	// FeeVoteCommitmentVoteQuery struct that have string  query for FeeVoteCommitmentVotes
	FeeVoteCommitmentVoteQuery struct {
		Fields    []string
		TableName string
	}
)

// NewFeeVoteCommitmentVoteQuery returns FeeVoteCommitmentVotesQuery instance
func NewFeeVoteCommitmentVoteQuery() *FeeVoteCommitmentVoteQuery {
	return &FeeVoteCommitmentVoteQuery{
		Fields: []string{
			"vote_hash",
			"voter_address",
			"block_height",
		},
		TableName: "fee_vote_commitment_vote",
	}
}

func (fsvc *FeeVoteCommitmentVoteQuery) getTableName() string {
	return fsvc.TableName
}

// InsertCommitVote to build insert query for `fee_vote_commitment_vote` table
func (fsvc *FeeVoteCommitmentVoteQuery) InsertCommitVote(voteCommit *model.FeeVoteCommitmentVote) (
	qStr string, args []interface{},
) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		fsvc.getTableName(),
		strings.Join(fsvc.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(fsvc.Fields)-1)),
	), fsvc.ExtractModel(voteCommit)
}

// GetVoteCommitByAccountAddressAndHeight to get vote commit by account address & block height
func (fsvc *FeeVoteCommitmentVoteQuery) GetVoteCommitByAccountAddressAndHeight(
	accountAddress string, height uint32,
) (
	qStr string, args []interface{},
) {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE voter_address = ? AND block_height>= ?`,
		strings.Join(fsvc.Fields, ","), fsvc.getTableName()), []interface{}{accountAddress, height}
}

// ExtractModel to  extract FeeVoteCommitmentVote model to []interface
func (*FeeVoteCommitmentVoteQuery) ExtractModel(voteCommit *model.FeeVoteCommitmentVote) []interface{} {
	return []interface{}{
		voteCommit.VoteHash,
		voteCommit.VoterAddress,
		voteCommit.BlockHeight,
	}
}

// Scan similar with `sql.Scan`
func (*FeeVoteCommitmentVoteQuery) Scan(voteCommit *model.FeeVoteCommitmentVote, row *sql.Row) error {
	err := row.Scan(
		&voteCommit.VoteHash,
		&voteCommit.VoterAddress,
		&voteCommit.BlockHeight,
	)
	return err
}

// Rollback delete records `WHERE block_height > "block_height"`
func (fsvc *FeeVoteCommitmentVoteQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", fsvc.getTableName()),
			height,
		},
	}
}

// SelectDataForSnapshot select only the block at snapshot block_height (fromHeight is unused)
func (fsvc *FeeVoteCommitmentVoteQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE block_height >= %d AND block_height <= %d`,
		strings.Join(fsvc.Fields, ","), fsvc.getTableName(), fromHeight, toHeight)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (fsvc *FeeVoteCommitmentVoteQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	// do not delete genesis block
	if fromHeight == 0 {
		fromHeight++
	}
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d`,
		fsvc.getTableName(), fromHeight, toHeight)
}
