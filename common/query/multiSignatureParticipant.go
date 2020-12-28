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
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	MultiSignatureParticipantQueryInterface interface {
		ExtractModel(participant *model.MultiSignatureParticipant) []interface{}
		InsertMultisignatureParticipants(participants []*model.MultiSignatureParticipant) (queries [][]interface{})
		Scan(participant *model.MultiSignatureParticipant, row *sql.Row) error
		BuildModel(rows *sql.Rows) (participants []*model.MultiSignatureParticipant, err error)
		GetMultiSignatureParticipantsByMultisigAddressAndHeightRange(
			multisigAddress []byte,
			fromHeight,
			toHeight uint32,
		) (str string, args []interface{})
	}
	MultiSignatureParticipantQuery struct {
		Fields    []string
		TableName string
	}
)

func NewMultiSignatureParticipantQuery() *MultiSignatureParticipantQuery {
	return &MultiSignatureParticipantQuery{
		Fields: []string{
			"multisig_address",
			"account_address",
			"account_address_index",
			// TODO: multisig participants should not have latest field. once they are added to a multisig address they can never been updated
			"latest",
			"block_height",
		},
		TableName: "multisignature_participant",
	}
}

func (msq *MultiSignatureParticipantQuery) getTableName() string {
	return msq.TableName
}

func (msq *MultiSignatureParticipantQuery) ExtractModel(participant *model.MultiSignatureParticipant) []interface{} {
	return []interface{}{
		participant.GetMultiSignatureAddress(),
		participant.GetAccountAddress(),
		participant.GetAccountAddressIndex(),
		participant.GetLatest(),
		participant.GetBlockHeight(),
	}
}
func (msq *MultiSignatureParticipantQuery) BuildModel(rows *sql.Rows) (participants []*model.MultiSignatureParticipant, err error) {
	for rows.Next() {

		var participant model.MultiSignatureParticipant
		err = rows.Scan(
			&participant.MultiSignatureAddress,
			&participant.AccountAddress,
			&participant.AccountAddressIndex,
			&participant.Latest,
			&participant.BlockHeight,
		)
		if err != nil {
			return participants, err
		}
		participants = append(participants, &participant)
	}
	return participants, nil
}
func (msq *MultiSignatureParticipantQuery) Scan(participant *model.MultiSignatureParticipant, row *sql.Row) error {
	return row.Scan(
		&participant.MultiSignatureAddress,
		&participant.AccountAddress,
		&participant.AccountAddressIndex,
		&participant.Latest,
		&participant.BlockHeight,
	)
}
func (msq *MultiSignatureParticipantQuery) InsertMultisignatureParticipants(
	participants []*model.MultiSignatureParticipant,
) (queries [][]interface{}) {
	var (
		qStr = fmt.Sprintf(
			"INSERT OR REPLACE INTO %s (%s) VALUES ",
			msq.getTableName(),
			strings.Join(msq.Fields, ", "),
		)
		args []interface{}
	)
	if participants != nil {
		for k, participant := range participants {
			qStr += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(msq.Fields)-1),
			)
			if k < len(participants)-1 {
				qStr += ", "
			}
			args = append(args, msq.ExtractModel(participant)...)
		}
		queries = append(
			queries,
			append([]interface{}{qStr}, args...),
			[]interface{}{
				fmt.Sprintf(
					"UPDATE %s SET latest = ? WHERE multisig_address = ? AND block_height != ? AND latest = ?",
					msq.getTableName(),
				),
				false,
				participants[0].GetMultiSignatureAddress(),
				participants[0].GetBlockHeight(),
				true,
			},
		)
		return queries
	}
	return nil
}

func (msq *MultiSignatureParticipantQuery) GetMultiSignatureParticipantsByMultisigAddressAndHeightRange(
	multisigAddress []byte,
	fromHeight,
	toHeight uint32,
) (str string, args []interface{}) {
	qry := fmt.Sprintf("SELECT %s FROM %s WHERE multisig_address = ? AND block_height >= ? AND block_height <= ? "+
		"ORDER BY account_address_index",
		strings.Join(msq.Fields, ","), msq.getTableName())
	return qry, []interface{}{
		multisigAddress,
		fromHeight,
		toHeight,
	}
}

func (msq *MultiSignatureParticipantQuery) Rollback(blockHeight uint32) [][]interface{} {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", msq.getTableName()),
			blockHeight,
		},
		{
			fmt.Sprintf("UPDATE %s SET latest = ? WHERE latest = ? AND (multisig_address, block_height"+
				") IN (SELECT t2.multisig_address, MAX(t2.block_height) "+
				"FROM %s as t2 GROUP BY t2.multisig_address)",
				msq.getTableName(),
				msq.getTableName(),
			),
			1, 0,
		},
	}
}

func (msq *MultiSignatureParticipantQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE (multisig_address, block_height) IN ("+
		"SELECT t2.multisig_address, MAX(t2.block_height) FROM %s as t2 "+
		"WHERE t2.block_height >= %d AND t2.block_height <= %d "+
		"GROUP BY t2.multisig_address) ORDER BY block_height",
		strings.Join(msq.Fields, ","), msq.getTableName(), msq.getTableName(), fromHeight, toHeight)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (msq *MultiSignatureParticipantQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d`,
		msq.getTableName(), fromHeight, toHeight)
}
