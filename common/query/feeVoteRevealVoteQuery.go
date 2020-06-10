package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	FeeVoteRevealVoteQueryInterface interface {
		GetFeeVoteRevealByAccountAddressAndRecentBlockHeight(accountAddress string, blockHeight uint32) (string, []interface{})
		GetFeeVoteRevealsInPeriod(
			lowerBlockHeight, upperBlockHeight uint32,
		) (string, []interface{})
		InsertRevealVote(revealVote *model.FeeVoteRevealVote) (string, []interface{})
		Scan(vote *model.FeeVoteRevealVote, row *sql.Row) error
		BuildModel(votes []*model.FeeVoteRevealVote, rows *sql.Rows) ([]*model.FeeVoteRevealVote, error)
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

// GetFeeVoteRevealByAccountAddressAndRecentBlockHeight represents getting fee_vote_reveal by account address
func (fvr *FeeVoteRevealVoteQuery) GetFeeVoteRevealByAccountAddressAndRecentBlockHeight(
	accountAddress string,
	blockHeight uint32,
) (qry string, args []interface{}) {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE voter_address = ? AND recent_block_height = ? ORDER BY block_height DESC LIMIT 1",
		strings.Join(fvr.Fields, ", "),
		fvr.getTableName(),
	), []interface{}{accountAddress, blockHeight}
}

// GetFeeVoteRevealsInPeriod select reveals within block-height range
// blockHeight limit are inclusive
func (fvr *FeeVoteRevealVoteQuery) GetFeeVoteRevealsInPeriod(
	lowerBlockHeight, upperBlockHeight uint32,
) (qry string, args []interface{}) {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE block_height between ? AND ? ORDER BY fee_vote ASC",
		strings.Join(fvr.Fields, ", "),
		fvr.getTableName(),
	), []interface{}{lowerBlockHeight, upperBlockHeight}
}

// InsertRevealVote represents insert new record to fee_vote_reveal
func (fvr *FeeVoteRevealVoteQuery) InsertRevealVote(revealVote *model.FeeVoteRevealVote) (qry string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		fvr.getTableName(),
		strings.Join(fvr.Fields, ", "),
		fmt.Sprintf("?%s", strings.Repeat(", ?", len(fvr.Fields)-1)),
	), fvr.ExtractModel(revealVote)
}

// ExtractModel extracting model.FeeVoteRevealVote values
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

// Scan build model.FeeVoteRevealVote from sql.Row
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
	return err
}

func (fvr *FeeVoteRevealVoteQuery) BuildModel(
	votes []*model.FeeVoteRevealVote, rows *sql.Rows,
) ([]*model.FeeVoteRevealVote, error) {
	for rows.Next() {
		var (
			revealVote = model.FeeVoteRevealVote{
				VoteInfo:       &model.FeeVoteInfo{},
				VoterSignature: nil,
				VoterAddress:   "",
				BlockHeight:    0,
			}
			err error
		)
		err = rows.Scan(
			&revealVote.VoteInfo.RecentBlockHash,
			&revealVote.VoteInfo.RecentBlockHeight,
			&revealVote.VoteInfo.FeeVote,
			&revealVote.VoterAddress,
			&revealVote.VoterSignature,
			&revealVote.BlockHeight,
		)
		if err != nil {
			return nil, err
		}
		votes = append(votes, &revealVote)
	}
	return votes, nil
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

// SelectDataForSnapshot select only the block at snapshot block_height
func (fvr *FeeVoteRevealVoteQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE block_height >= %d AND block_height <= %d`,
		strings.Join(fvr.Fields, ", "), fvr.getTableName(), fromHeight, toHeight)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (fvr *FeeVoteRevealVoteQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d`,
		fvr.getTableName(), fromHeight, toHeight)
}
