package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	FeeVoteRevealVoteQueryInterface interface {
		GetFeeVoteRevealByAccountAddress(accountAddress string) (string, []interface{})
		InsertRevealVote(revealVote *model.FeeVoteRevealVote) (string, []interface{})
		Scan(vote *model.FeeVoteRevealVote, row *sql.Row) error
	}
	FeeVoteRevealVoteQuery struct {
		Fields    []string
		TableName string
	}
)

func NewFeeVoteRevealVoteQuery() *FeeVoteRevealVoteQuery {
	return &FeeVoteRevealVoteQuery{
		Fields: []string{
			"recent_block_hash",
			"recent_block_height",
			"fee_vote",
			"voter_address",
			"voter_signature",
			"block_height",
		},
		TableName: "fee_vote_reveal_vote",
	}
}

func (fvr *FeeVoteRevealVoteQuery) getTableName() string {
	return fvr.TableName
}
func (fvr *FeeVoteRevealVoteQuery) GetFeeVoteRevealByAccountAddress(accountAddress string) (qry string, args []interface{}) {
	return fmt.Sprintf(
		"SELECT (%s) FROM %s WHERE voter_address = ? ORDER BY block_height DESC LIMIT 1",
		strings.Join(fvr.Fields, ", "),
		fvr.getTableName(),
	), []interface{}{accountAddress}
}

func (fvr *FeeVoteRevealVoteQuery) InsertRevealVote(revealVote *model.FeeVoteRevealVote) (qry string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		fvr.getTableName(),
		strings.Join(fvr.Fields, ", "),
		fmt.Sprintf("?%s", strings.Repeat(", ?", len(fvr.Fields)-1)),
	), fvr.ExtractModel(revealVote)
}

func (*FeeVoteRevealVoteQuery) ExtractModel(revealVote *model.FeeVoteRevealVote) []interface{} {
	return []interface{}{
		revealVote.VoteInfo.GetRecentBlockHash(),
		revealVote.VoteInfo.GetRecentBlockHeight(),
		revealVote.VoteInfo.GetFeeVote(),
		revealVote.GetVoterAddress(),
		revealVote.GetVoterSignature(),
		revealVote.GetBlockHeight(),
	}
}

func (fvr *FeeVoteRevealVoteQuery) Scan(vote *model.FeeVoteRevealVote, row *sql.Row) error {
	var (
		voterAddress   string
		blockHeight    uint32
		voterSignature []byte
		feeVoteInfo    model.FeeVoteInfo
	)
	err := row.Scan(
		&feeVoteInfo.RecentBlockHash,
		&feeVoteInfo.RecentBlockHeight,
		&feeVoteInfo.FeeVote,
		&voterAddress,
		&voterSignature,
		&blockHeight,
	)
	vote.VoterAddress = voterAddress
	vote.BlockHeight = blockHeight
	vote.VoterSignature = voterSignature
	vote.VoteInfo = &feeVoteInfo
	fmt.Println(vote)
	return err
}

// Rollback delete records `WHERE block_height > "block_height"`
func (fvr *FeeVoteRevealVoteQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", fvr.getTableName()),
			height,
		},
	}
}
