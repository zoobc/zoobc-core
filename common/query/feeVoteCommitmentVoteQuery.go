// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package query

import (
	"database/sql"
	"fmt"
	"github.com/zoobc/zoobc-core/common/blocker"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// FeeVoteCommitmentVoteQueryInterface interface that implemented by FeeVoteCommitmentVoteQuery
	FeeVoteCommitmentVoteQueryInterface interface {
		GetVoteCommitByAccountAddressAndHeight(
			accountAddress []byte,
			height uint32,
		) (qStr string, args []interface{})
		InsertCommitVote(voteCommit *model.FeeVoteCommitmentVote) (qStr string, args []interface{})
		InsertCommitVotes(voteCommits []*model.FeeVoteCommitmentVote) (qStr string, args []interface{})
		ExtractModel(voteCommit *model.FeeVoteCommitmentVote) []interface{}
		Scan(voteCommit *model.FeeVoteCommitmentVote, row *sql.Row) error
		BuildModel(
			feeVoteCommitmentVotes []*model.FeeVoteCommitmentVote,
			rows *sql.Rows,
		) ([]*model.FeeVoteCommitmentVote, error)
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

func (fsvc *FeeVoteCommitmentVoteQuery) InsertCommitVotes(voteCommits []*model.FeeVoteCommitmentVote) (str string, args []interface{}) {
	if len(voteCommits) > 0 {
		str = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			fsvc.getTableName(),
			strings.Join(fsvc.Fields, ", "),
		)
		for k, voteCommit := range voteCommits {
			str += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(fsvc.Fields)-1),
			)
			if k < len(voteCommits)-1 {
				str += ","
			}
			args = append(args, fsvc.ExtractModel(voteCommit)...)
		}
	}
	return str, args
}

// ImportSnapshot takes payload from downloaded snapshot and insert them into database
func (fsvc *FeeVoteCommitmentVoteQuery) ImportSnapshot(payload interface{}) ([][]interface{}, error) {
	var (
		queries [][]interface{}
	)
	commits, ok := payload.([]*model.FeeVoteCommitmentVote)
	if !ok {
		return nil, blocker.NewBlocker(blocker.DBErr, "ImportSnapshotCannotCastTo"+fsvc.TableName)
	}
	if len(commits) > 0 {
		recordsPerPeriod, rounds, remaining := CalculateBulkSize(len(fsvc.Fields), len(commits))
		for i := 0; i < rounds; i++ {
			qry, args := fsvc.InsertCommitVotes(commits[i*recordsPerPeriod : (i*recordsPerPeriod)+recordsPerPeriod])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
		if remaining > 0 {
			qry, args := fsvc.InsertCommitVotes(commits[len(commits)-remaining:])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
	}
	return queries, nil
}

// RecalibrateVersionedTable recalibrate table to clean up multiple latest rows due to import function
func (fsvc *FeeVoteCommitmentVoteQuery) RecalibrateVersionedTable() []string {
	return []string{} // only table with `latest` column need this
}

// GetVoteCommitByAccountAddressAndHeight to get vote commit by account address & block height
func (fsvc *FeeVoteCommitmentVoteQuery) GetVoteCommitByAccountAddressAndHeight(
	accountAddress []byte, height uint32,
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

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (*FeeVoteCommitmentVoteQuery) BuildModel(
	feeVoteCommitmentVotes []*model.FeeVoteCommitmentVote,
	rows *sql.Rows,
) ([]*model.FeeVoteCommitmentVote, error) {
	for rows.Next() {
		var (
			feeCommit model.FeeVoteCommitmentVote
			err       error
		)
		err = rows.Scan(
			&feeCommit.VoteHash,
			&feeCommit.VoterAddress,
			&feeCommit.BlockHeight,
		)
		if err != nil {
			return nil, err
		}
		feeVoteCommitmentVotes = append(feeVoteCommitmentVotes, &feeCommit)
	}
	return feeVoteCommitmentVotes, nil
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
	return fmt.Sprintf(`SELECT %s FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0`,
		strings.Join(fsvc.Fields, ","), fsvc.getTableName(), fromHeight, toHeight)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (fsvc *FeeVoteCommitmentVoteQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0`,
		fsvc.getTableName(), fromHeight, toHeight)
}
