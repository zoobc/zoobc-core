package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// FeeScaleVoteCommitQueryInterface interface that implemented by FeeScaleVoteCommitQuery
	FeeScaleVoteCommitQueryInterface interface {
		GetVoteCommitByAccountAddress(accountAddress string) (qStr string, args []interface{})
		InsertCommitVote(voteCommit *model.FeeScaleVoteCommit) (str string, args []interface{})
		ExtractModel(voteCommit *model.FeeScaleVoteCommit) []interface{}
		Scan(voteCommit *model.FeeScaleVoteCommit, row *sql.Row) error
		Rollback(height uint32) (multiQueries [][]interface{})
	}
	// FeeScaleVoteCommitQuery struct that have string  query for FeeScaleVoteCommits
	FeeScaleVoteCommitQuery struct {
		Fields    []string
		TableName string
	}
)

// NewFeeScaleVoteCommitsQuery returns FeeScaleVoteCommitsQuery instance
func NewFeeScaleVoteCommitsQuery() *FeeScaleVoteCommitQuery {
	return &FeeScaleVoteCommitQuery{
		Fields: []string{
			"vote_hash",
			"voter_address",
			"block_height",
		},
		TableName: "fee_scale_vote_commits",
	}
}

func (fsvc *FeeScaleVoteCommitQuery) getTableName() string {
	return fsvc.TableName
}

// InsertCommitVote to build insert query for `fee_scale_vote_commits` table
func (fsvc *FeeScaleVoteCommitQuery) InsertCommitVote(voteCommit *model.FeeScaleVoteCommit) (
	str string, args []interface{},
) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		fsvc.getTableName(),
		strings.Join(fsvc.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(fsvc.Fields)-1)),
	), fsvc.ExtractModel(voteCommit)
}

// GetVoteCommitByAccountAddress to get vote commit by account address & block height
func (fsvc *FeeScaleVoteCommitQuery) GetVoteCommitByAccountAddress(accountAddress string) (
	qStr string, args []interface{},
) {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE voter_address = ? ORDER BY block_height DESC LIMIT 1`,
		strings.Join(fsvc.Fields, ","), fsvc.getTableName()), []interface{}{accountAddress}
}

// ExtractModel to  extract FeeScaleVoteCommit model to []interface
func (*FeeScaleVoteCommitQuery) ExtractModel(voteCommit *model.FeeScaleVoteCommit) []interface{} {
	return []interface{}{
		voteCommit.VoteHash,
		voteCommit.VoterAddress,
		voteCommit.BlockHeight,
	}
}

// Scan similar with `sql.Scan`
func (*FeeScaleVoteCommitQuery) Scan(voteCommit *model.FeeScaleVoteCommit, row *sql.Row) error {
	err := row.Scan(
		&voteCommit.VoteHash,
		&voteCommit.VoterAddress,
		&voteCommit.BlockHeight,
	)
	return err
}

// Rollback delete records `WHERE block_height > "block_height"`
func (fsvc *FeeScaleVoteCommitQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", fsvc.getTableName()),
			height,
		},
	}
}

// SelectDataForSnapshot select only the block at snapshot block_height (fromHeight is unused)
func (fsvc *FeeScaleVoteCommitQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE block_height >= %d AND block_height <= %d`,
		strings.Join(fsvc.Fields, ","), fsvc.getTableName(), fromHeight, toHeight)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (fsvc *FeeScaleVoteCommitQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	// do not delete genesis block
	if fromHeight == 0 {
		fromHeight++
	}
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d`,
		fsvc.getTableName(), fromHeight, toHeight)
}
