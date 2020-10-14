package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	MultisignatureInfoQueryInterface interface {
		GetMultisignatureInfoByAddressWithParticipants(
			multisigAddress []byte,
			currentHeight, limit uint32,
		) (str string, args []interface{})
		InsertMultisignatureInfo(multisigInfo *model.MultiSignatureInfo) [][]interface{}
		InsertMultiSignatureInfos(multiSignatureInfos []*model.MultiSignatureInfo) [][]interface{}
		Scan(multisigInfo *model.MultiSignatureInfo, row *sql.Row) error
		ExtractModel(multisigInfo *model.MultiSignatureInfo) []interface{}
		BuildModel(multisigInfos []*model.MultiSignatureInfo, rows *sql.Rows) ([]*model.MultiSignatureInfo, error)
		BuildModelWithParticipant(multisigInfos []*model.MultiSignatureInfo, rows *sql.Rows) ([]*model.MultiSignatureInfo, error)
	}

	MultisignatureInfoQuery struct {
		Fields    []string
		TableName string
	}
)

// NewMultisignatureInfoQuery returns PendingTransactionQuery instance
func NewMultisignatureInfoQuery() *MultisignatureInfoQuery {
	return &MultisignatureInfoQuery{
		Fields: []string{
			"multisig_address",
			"minimum_signatures",
			"nonce",
			"block_height",
			"latest",
		},
		TableName: "multisignature_info",
	}
}

func (msi *MultisignatureInfoQuery) getTableName() string {
	return msi.TableName
}

// GetMultisignatureInfoByAddressWithParticipants
func (msi *MultisignatureInfoQuery) GetMultisignatureInfoByAddressWithParticipants(
	multisigAddress []byte,
	currentHeight, limit uint32,
) (str string, args []interface{}) {
	var (
		blockHeight   uint32
		t1Fields      []string
		msParticipant = NewMultiSignatureParticipantQuery()
	)
	if currentHeight > limit {
		blockHeight = currentHeight - limit
	}
	for _, msiField := range msi.Fields {
		t1Fields = append(t1Fields, fmt.Sprintf("t1.%s", msiField))
	}
	queryMultisigInfo := fmt.Sprintf(
		"SELECT %s, t2.account_address FROM %s t1 LEFT JOIN %s t2 ON t1.multisig_address = t2.multisig_address "+
			"WHERE t1.multisig_address = ? AND t1.block_height >= ? AND t1.latest = true AND t2.latest = true "+
			"ORDER BY t2.account_address_index DESC",
		strings.Join(t1Fields, ", "),
		msi.getTableName(),
		msParticipant.getTableName(),
	)
	return queryMultisigInfo, []interface{}{
		multisigAddress,
		multisigAddress,
		blockHeight,
	}
}

// InsertMultisignatureInfo inserts a new pending transaction into DB
func (msi *MultisignatureInfoQuery) InsertMultisignatureInfo(multisigInfo *model.MultiSignatureInfo) [][]interface{} {
	var queries [][]interface{}
	insertQuery := fmt.Sprintf("INSERT OR REPLACE INTO %s (%s) VALUES(%s)",
		msi.getTableName(),
		strings.Join(msi.Fields, ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ? ", len(msi.Fields)-1)),
	)
	updateQuery := fmt.Sprintf("UPDATE %s SET latest = false WHERE multisig_address = ? "+
		"AND block_height != %d AND latest = true",
		msi.getTableName(),
		multisigInfo.BlockHeight,
	)
	queries = append(queries,
		append([]interface{}{insertQuery}, msi.ExtractModel(multisigInfo)...),
		[]interface{}{
			updateQuery, multisigInfo.MultisigAddress,
		},
	)
	return queries
}

// InsertMultiSignatureInfos represents query builder to insert multiple records into multisignature_info and multisignature_participant table
// without updating the version
func (msi *MultisignatureInfoQuery) InsertMultiSignatureInfos(multiSignatureInfos []*model.MultiSignatureInfo) [][]interface{} {
	var (
		participantQueryInterface      = NewMultiSignatureParticipantQuery()
		queries                        [][]interface{}
		musigInfoArgs, participantArgs []interface{}
		musigInfoQ                     = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			msi.getTableName(),
			strings.Join(msi.Fields, ", "),
		)
		participantQ = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES",
			participantQueryInterface.getTableName(),
			strings.Join(participantQueryInterface.Fields, ", "),
		)
	)

	if len(multiSignatureInfos) > 0 {
		for m, musig := range multiSignatureInfos {
			musigInfoQ += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(msi.Fields)-1),
			)
			if m < len(multiSignatureInfos)-1 {
				musigInfoQ += ","
			}
			musigInfoArgs = append(musigInfoArgs, msi.ExtractModel(musig)...)

			for a, address := range musig.GetAddresses() {
				participantQ += fmt.Sprintf("(?%s)", strings.Repeat(", ?", len(participantQueryInterface.Fields)-1))

				if !(a == len(musig.GetAddresses())-1 && m == len(multiSignatureInfos)-1) {
					participantQ += ","
				}
				participantArgs = append(participantArgs, participantQueryInterface.ExtractModel(&model.MultiSignatureParticipant{
					MultiSignatureAddress: musig.GetMultisigAddress(),
					AccountAddress:        address,
					AccountAddressIndex:   uint32(a),
					BlockHeight:           musig.GetBlockHeight(),
					Latest:                musig.GetLatest(),
				})...)
			}
		}

		queries = append(
			queries,
			append([]interface{}{musigInfoQ}, musigInfoArgs...),
			append([]interface{}{participantQ}, participantArgs...),
		)
	}
	return queries
}

// ImportSnapshot takes payload from downloaded snapshot and insert them into database
func (msi *MultisignatureInfoQuery) ImportSnapshot(payload interface{}) ([][]interface{}, error) {
	var (
		queries [][]interface{}
	)
	musigInfos, ok := payload.([]*model.MultiSignatureInfo)
	if !ok {
		return nil, blocker.NewBlocker(blocker.DBErr, "ImportSnapshotCannotCastTo"+msi.TableName)
	}
	if len(musigInfos) > 0 {
		recordsPerPeriod, rounds, remaining := CalculateBulkSize(len(msi.Fields), len(musigInfos))
		for i := 0; i < rounds; i++ {
			qry := msi.InsertMultiSignatureInfos(musigInfos[i*recordsPerPeriod : (i*recordsPerPeriod)+recordsPerPeriod])
			queries = append(queries, qry...)
		}
		if remaining > 0 {
			qry := msi.InsertMultiSignatureInfos(musigInfos[len(musigInfos)-remaining:])
			queries = append(queries, qry...)
		}
	}
	return queries, nil
}

// RecalibrateVersionedTable recalibrate table to clean up multiple latest rows due to import function
func (msi *MultisignatureInfoQuery) RecalibrateVersionedTable() []string {
	return []string{
		fmt.Sprintf(
			"update %s set latest = false where latest = true AND (multisig_address, block_height) NOT IN "+
				"(select t2.multisig_address, max(t2.block_height) from %s t2 group by t2.multisig_address)",
			msi.getTableName(), msi.getTableName()),
		fmt.Sprintf(
			"update %s set latest = true where latest = false AND (multisig_address, block_height) IN "+
				"(select t2.multisig_address, max(t2.block_height) from %s t2 group by t2.multisig_address)",
			msi.getTableName(), msi.getTableName()),
	}
}

// Scan will build model from *sql.Row
func (*MultisignatureInfoQuery) Scan(multisigInfo *model.MultiSignatureInfo, row *sql.Row) error {
	err := row.Scan(
		&multisigInfo.MultisigAddress,
		&multisigInfo.MinimumSignatures,
		&multisigInfo.Nonce,
		&multisigInfo.BlockHeight,
		&multisigInfo.Latest,
	)
	return err
}

// ExtractModel will get values exclude addresses, perfectly used while inserting new record.
func (*MultisignatureInfoQuery) ExtractModel(multisigInfo *model.MultiSignatureInfo) []interface{} {
	return []interface{}{
		&multisigInfo.MultisigAddress,
		&multisigInfo.MinimumSignatures,
		&multisigInfo.Nonce,
		&multisigInfo.BlockHeight,
		&multisigInfo.Latest,
	}
}

// BuildModel will build model from *sql.Rows
func (msi *MultisignatureInfoQuery) BuildModel(
	mss []*model.MultiSignatureInfo, rows *sql.Rows,
) ([]*model.MultiSignatureInfo, error) {
	for rows.Next() {
		var (
			multisigInfo model.MultiSignatureInfo
		)
		rows.Columns()
		err := rows.Scan(
			&multisigInfo.MultisigAddress,
			&multisigInfo.MinimumSignatures,
			&multisigInfo.Nonce,
			&multisigInfo.BlockHeight,
			&multisigInfo.Latest,
		)
		if err != nil {
			return nil, err
		}
		mss = append(mss, &multisigInfo)
	}
	return mss, nil
}

// BuildModelWithParticipant will build model from *sql.Rows that expect has addresses column
// which is result from sub query of multisignature_participant
func (msi *MultisignatureInfoQuery) BuildModelWithParticipant(
	mss []*model.MultiSignatureInfo, rows *sql.Rows,
) ([]*model.MultiSignatureInfo, error) {
	for rows.Next() {
		var (
			multisigInfo       model.MultiSignatureInfo
			participantAddress []byte
		)
		rows.Columns()
		err := rows.Scan(
			&multisigInfo.MultisigAddress,
			&multisigInfo.MinimumSignatures,
			&multisigInfo.Nonce,
			&multisigInfo.BlockHeight,
			&multisigInfo.Latest,
			&participantAddress,
		)
		multisigInfo.Addresses = [][]byte{participantAddress}
		if err != nil {
			return nil, err
		}
		mss = append(mss, &multisigInfo)
	}
	return mss, nil
}

// Rollback delete records `WHERE block_height > "height"`
func (msi *MultisignatureInfoQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", msi.getTableName()),
			height,
		},
		{
			fmt.Sprintf("UPDATE %s SET latest = ? WHERE latest = ? AND (multisig_address, block_height"+
				") IN (SELECT t2.multisig_address, MAX(t2.block_height) "+
				"FROM %s as t2 GROUP BY t2.multisig_address)",
				msi.getTableName(),
				msi.getTableName(),
			),
			1, 0,
		},
	}
}

func (msi *MultisignatureInfoQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE (multisig_address, block_height) IN "+
			"(SELECT t2.multisig_address, MAX(t2.block_height) FROM %s t2 "+
			"WHERE t2.block_height >= %d AND t2.block_height <= %d AND t2.block_height != 0 GROUP BY t2.multisig_address) "+
			"ORDER BY block_height",
		strings.Join(msi.Fields, ", "),
		msi.getTableName(),
		msi.getTableName(),
		fromHeight,
		toHeight,
	)
	//
	// return fmt.Sprintf(
	// 	"SELECT %s, %s FROM %s WHERE (multisig_address, block_height) IN (SELECT t2.multisig_address, MAX(t2.block_height) FROM %s t2 WHERE t2.block_height >= %d AND t2.block_height <= %d AND t2.block_height != 0 GROUP BY t2.multisig_address) ORDER BY block_height, account_address_index",
	// 	strings.Join(msi.Fields, ", "),
	// 	"(SELECT GROUP_CONCAT(account_address, ',') FROM multisignature_participant GROUP BY multisig_address, block_height "+
	// 		"ORDER BY account_address_index ASC) as addresses",
	// 	msi.getTableName(),
	// 	msi.getTableName(),
	// 	fromHeight,
	// 	toHeight,
	// )
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (msi *MultisignatureInfoQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0`,
		msi.getTableName(), fromHeight, toHeight)
}
