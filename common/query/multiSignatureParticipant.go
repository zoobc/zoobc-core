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
