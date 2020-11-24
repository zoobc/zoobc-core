package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	PendingSignatureQueryInterface interface {
		GetPendingSignatureByHash(
			txHash []byte,
			currentHeight, limit uint32,
		) (str string, args []interface{})
		InsertPendingSignature(pendingSig *model.PendingSignature) [][]interface{}
		InsertPendingSignatures(pendingSigs []*model.PendingSignature) (str string, args []interface{})
		Scan(pendingSig *model.PendingSignature, row *sql.Row) error
		ExtractModel(pendingSig *model.PendingSignature) []interface{}
		BuildModel(pendingSigs []*model.PendingSignature, rows *sql.Rows) ([]*model.PendingSignature, error)
	}

	PendingSignatureQuery struct {
		Fields    []string
		TableName string
	}
)

// NewPendingSignatureQuery returns PendingTransactionQuery instance
func NewPendingSignatureQuery() *PendingSignatureQuery {
	return &PendingSignatureQuery{
		Fields: []string{
			"transaction_hash",
			"account_address",
			"multisig_address",
			"signature",
			"block_height",
			"latest",
		},
		TableName: "pending_signature",
	}
}

func (psq *PendingSignatureQuery) getTableName() string {
	return psq.TableName
}

func (psq *PendingSignatureQuery) GetPendingSignatureByHash(
	txHash []byte,
	currentHeight, limit uint32,
) (str string, args []interface{}) {
	var (
		blockHeight uint32
	)
	if currentHeight > limit {
		blockHeight = currentHeight - limit
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE transaction_hash = ? AND block_height >= ? AND latest = true",
		strings.Join(psq.Fields, ", "), psq.getTableName())
	return query, []interface{}{
		txHash,
		blockHeight,
	}
}

// InsertPendingSignature inserts a new pending transaction into DB
func (psq *PendingSignatureQuery) InsertPendingSignature(pendingSig *model.PendingSignature) [][]interface{} {
	var queries [][]interface{}
	insertQuery := fmt.Sprintf("INSERT OR REPLACE INTO %s (%s) VALUES(%s)",
		psq.getTableName(),
		strings.Join(psq.Fields, ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ? ", len(psq.Fields)-1)),
	)
	updateQuery := fmt.Sprintf("UPDATE %s SET latest = false WHERE account_address = ? AND transaction_hash = ? "+
		"AND block_height != %d AND latest = true",
		psq.getTableName(),
		pendingSig.BlockHeight,
	)
	queries = append(queries,
		append([]interface{}{insertQuery}, psq.ExtractModel(pendingSig)...),
		[]interface{}{
			updateQuery, pendingSig.AccountAddress, pendingSig.TransactionHash,
		},
	)
	return queries
}

// InsertPendingSignatures represents query builder to insert multiple record in single query
func (psq *PendingSignatureQuery) InsertPendingSignatures(pendingSigs []*model.PendingSignature) (str string, args []interface{}) {
	if len(pendingSigs) > 0 {
		str = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			psq.getTableName(),
			strings.Join(psq.Fields, ", "),
		)
		for k, pendingSig := range pendingSigs {
			str += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(psq.Fields)-1),
			)
			if k < len(pendingSigs)-1 {
				str += ", "
			}
			args = append(args, psq.ExtractModel(pendingSig)...)
		}
	}
	return str, args
}

// ImportSnapshot takes payload from downloaded snapshot and insert them into database
func (psq *PendingSignatureQuery) ImportSnapshot(payload interface{}) ([][]interface{}, error) {
	var (
		queries [][]interface{}
	)
	pendingSigs, ok := payload.([]*model.PendingSignature)
	if !ok {
		return nil, blocker.NewBlocker(blocker.DBErr, "ImportSnapshotCannotCastTo"+psq.TableName)
	}
	if len(pendingSigs) > 0 {
		recordsPerPeriod, rounds, remaining := CalculateBulkSize(len(psq.Fields), len(pendingSigs))
		for i := 0; i < rounds; i++ {
			qry, args := psq.InsertPendingSignatures(pendingSigs[i*recordsPerPeriod : (i*recordsPerPeriod)+recordsPerPeriod])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
		if remaining > 0 {
			qry, args := psq.InsertPendingSignatures(pendingSigs[len(pendingSigs)-remaining:])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
	}
	return queries, nil
}

// RecalibrateVersionedTable recalibrate table to clean up multiple latest rows due to import function
func (psq *PendingSignatureQuery) RecalibrateVersionedTable() []string {
	return []string{
		fmt.Sprintf(
			"update %s set latest = false where latest = true AND (account_address, transaction_hash, block_height) NOT IN "+
				"(select t2.account_address, t2.transaction_hash, max(t2.block_height) from %s t2 group by t2.account_address, t2.transaction_hash)",
			psq.getTableName(), psq.getTableName()),
		fmt.Sprintf(
			"update %s set latest = true where latest = false AND (account_address, transaction_hash, block_height) IN "+
				"(select t2.account_address, t2.transaction_hash, max(t2.block_height) from %s t2 group by t2.account_address, t2.transaction_hash)",
			psq.getTableName(), psq.getTableName()),
	}
}

func (*PendingSignatureQuery) Scan(pendingSig *model.PendingSignature, row *sql.Row) error {
	err := row.Scan(
		&pendingSig.TransactionHash,
		&pendingSig.MultiSignatureAddress,
		&pendingSig.AccountAddress,
		&pendingSig.Signature,
		&pendingSig.BlockHeight,
		&pendingSig.Latest,
	)
	return err
}

func (*PendingSignatureQuery) ExtractModel(pendingSig *model.PendingSignature) []interface{} {
	return []interface{}{
		&pendingSig.TransactionHash,
		&pendingSig.MultiSignatureAddress,
		&pendingSig.AccountAddress,
		&pendingSig.Signature,
		&pendingSig.BlockHeight,
		&pendingSig.Latest,
	}
}

func (psq *PendingSignatureQuery) BuildModel(
	pss []*model.PendingSignature, rows *sql.Rows,
) ([]*model.PendingSignature, error) {
	for rows.Next() {
		var pendingSig model.PendingSignature
		err := rows.Scan(
			&pendingSig.TransactionHash,
			&pendingSig.MultiSignatureAddress,
			&pendingSig.AccountAddress,
			&pendingSig.Signature,
			&pendingSig.BlockHeight,
			&pendingSig.Latest,
		)
		if err != nil {
			return nil, err
		}
		pss = append(pss, &pendingSig)
	}
	return pss, nil
}

// Rollback delete records `WHERE block_height > "height"`
func (psq *PendingSignatureQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", psq.TableName),
			height,
		},
		{
			fmt.Sprintf("UPDATE %s SET latest = ? WHERE latest = ? AND (account_address, transaction_hash, "+
				"block_height) IN (SELECT t2.account_address, t2.transaction_hash, "+
				"MAX(t2.block_height) FROM %s as t2 GROUP BY t2.account_address, t2.transaction_hash)",
				psq.TableName,
				psq.TableName,
			),
			1, 0,
		},
	}
}

func (psq *PendingSignatureQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE (account_address, transaction_hash, block_height) "+
			"IN (SELECT t2.account_address, t2.transaction_hash, MAX(t2.block_height) FROM %s as t2 "+
			"WHERE t2.block_height >= %d AND t2.block_height <= %d AND t2.block_height != 0 "+
			"GROUP BY t2.account_address, t2.transaction_hash) ORDER BY block_height",
		strings.Join(psq.Fields, ","),
		psq.TableName,
		psq.TableName,
		fromHeight,
		toHeight,
	)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (psq *PendingSignatureQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0`,
		psq.TableName, fromHeight, toHeight)
}
