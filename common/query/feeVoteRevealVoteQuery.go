package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	FeeVoteRevealVoteQueryInterface interface {
		GetFeeVoteRevealByAccountAddress(accountAddress string) (qStr string, args []interface{})
		InsertRevealVote(revealVote *model.FeeVoteRevealVote) (qStr string, args []interface{})
		Scan(vote *model.FeeVoteRevealVote, row *sql.Row) error
	}
	FeeVoteRevealVoteQuery struct {
		Fields    []string
		TableName string
	}
)

func NewFeeVoteRevealVoteQuery() *FeeVoteCommitmentVoteQuery {
	return &FeeVoteCommitmentVoteQuery{
		Fields: []string{
			"voter_address",
			"recent_block_hash",
			"recent_block_height",
			"fee_vote",
			"voter_signature",
			"block_height",
		},
		TableName: "fee_vote_reveal_vote",
	}
}

func (fvr *FeeVoteRevealVoteQuery) getTableName() string {
	return fvr.TableName
}
func (fvr *FeeVoteRevealVoteQuery) GetFeeVoteRevealByAccountAddress(accountAddress string) (qStr string, args []interface{}) {
	return fmt.Sprintf(
		"SELECT (%s) FROM %s WHERE voter_address = ? ORDER BY block_height DESC LIMIT 1",
		strings.Join(fvr.Fields, ", "),
		fvr.getTableName(),
	), []interface{}{accountAddress}
}

func (fvr *FeeVoteRevealVoteQuery) InsertRevealVote(revealVote *model.FeeVoteRevealVote) (qStr string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		fvr.getTableName(),
		strings.Join(fvr.Fields, ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(fvr.Fields)-1)),
	), fvr.ExtractModel(revealVote)
}

func (*FeeVoteRevealVoteQuery) ExtractModel(revealVote *model.FeeVoteRevealVote) []interface{} {
	return []interface{}{
		revealVote.VoteInfo.GetRecentBlockHash(),
		revealVote.VoteInfo.GetRecentBlockHeight(),
		revealVote.VoteInfo.GetFeeVote(),
		revealVote.GetVoterSignature(),
		revealVote.GetVoterAddress(),
		revealVote.GetBlockHeight(),
	}
}

func (fvr *FeeVoteRevealVoteQuery) Scan(vote *model.FeeVoteRevealVote, row *sql.Row) error {
	return row.Scan(
		&vote.VoterAddress,
		&vote.VoteInfo.RecentBlockHash,
		&vote.VoteInfo.RecentBlockHeight,
		&vote.VoteInfo.FeeVote,
		&vote.VoterSignature,
		&vote.BlockHeight,
	)
}
