package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// EscrowTransactionQuery fields must have
	EscrowTransactionQuery struct {
		Fields    []string
		TableName string
	}

	// EscrowTransactionQueryInterface methods must have
	EscrowTransactionQueryInterface interface {
		InsertEscrowTransaction(escrow *model.Escrow) [][]interface{}
		GetLatestEscrowTransactionByID(int64) (string, []interface{})
		GetEscrowTransactions(fields map[string]interface{}) (string, []interface{})
		ExpiringEscrowTransactions(blockHeight uint32) (string, []interface{})
		ExtractModel(*model.Escrow) []interface{}
		BuildModels(*sql.Rows) ([]*model.Escrow, error)
		Scan(escrow *model.Escrow, row *sql.Row) error
		TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string
	}
)

// NewEscrowTransactionQuery build an EscrowTransactionQuery
func NewEscrowTransactionQuery() *EscrowTransactionQuery {
	return &EscrowTransactionQuery{
		Fields: []string{
			"id",
			"sender_address",
			"recipient_address",
			"approver_address",
			"amount",
			"commission",
			"timeout",
			"status",
			"block_height",
			"latest",
			"instruction",
		},
		TableName: "escrow_transaction",
	}
}

func (et *EscrowTransactionQuery) getTableName() string {
	return et.TableName
}

/*
InsertEscrowTransaction represents insert query for escrow_transaction table.
There 2 queries result:
		1. Update the previous record to latest = false
		2. Insert new record which is the newest
*/
func (et *EscrowTransactionQuery) InsertEscrowTransaction(escrow *model.Escrow) [][]interface{} {
	return [][]interface{}{
		{
			fmt.Sprintf(
				"UPDATE %s set latest = ? WHERE id = ?",
				et.getTableName(),
			),
			false,
			escrow.GetID(),
		},
		append(
			[]interface{}{
				fmt.Sprintf(
					"INSERT INTO %s (%s) VALUES(%s)",
					et.getTableName(),
					strings.Join(et.Fields, ","),
					fmt.Sprintf("? %s", strings.Repeat(", ?", len(et.Fields)-1))),
			},
			et.ExtractModel(escrow)...,
		),
	}
}

// GetLatestEscrowTransactionByID represents getting latest escrow by id
func (et *EscrowTransactionQuery) GetLatestEscrowTransactionByID(id int64) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"SELECT %s FROM %s WHERE id = ? AND latest = ?",
			strings.Join(et.Fields, ", "),
			et.getTableName(),
		),
		[]interface{}{id, true}
}

// GetEscrowTransactions represents SELECT with multiple clauses connected via AND operand
func (et *EscrowTransactionQuery) GetEscrowTransactions(fields map[string]interface{}) (qStr string, args []interface{}) {
	qStr = fmt.Sprintf("SELECT %s FROM %s ", strings.Join(et.Fields, ", "), et.getTableName())

	if len(fields) > 0 {
		qStr += "WHERE "
		i := 1
		for k, v := range fields {
			qStr += fmt.Sprintf("%s = ? ", k)
			if i < len(fields) {
				qStr += "AND "
			}
			args = append(args, v)
			i++
		}
	}

	return qStr, args
}

// ExpiringEscrowTransactions represents update escrows status to expired where that has been expired by blockHeight
func (et *EscrowTransactionQuery) ExpiringEscrowTransactions(blockHeight uint32) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"UPDATE %s SET latest = ?, status = ? WHERE timeout < ? AND status = 0",
			et.getTableName(),
		),
		[]interface{}{
			1,
			model.EscrowStatus_Expired,
			blockHeight,
		}
}

// ExtractModel will extract values of escrow as []interface{}
func (et *EscrowTransactionQuery) ExtractModel(escrow *model.Escrow) []interface{} {
	return []interface{}{
		escrow.GetID(),
		escrow.GetSenderAddress(),
		escrow.GetRecipientAddress(),
		escrow.GetApproverAddress(),
		escrow.GetAmount(),
		escrow.GetCommission(),
		escrow.GetTimeout(),
		escrow.GetStatus(),
		escrow.GetBlockHeight(),
		escrow.GetLatest(),
		escrow.GetInstruction(),
	}
}

// BuildModels extract sqlRaw into []*model.Escrow
func (et *EscrowTransactionQuery) BuildModels(rows *sql.Rows) ([]*model.Escrow, error) {
	var (
		escrows []*model.Escrow
		err     error
	)

	for rows.Next() {
		var escrow model.Escrow
		err = rows.Scan(
			&escrow.ID,
			&escrow.SenderAddress,
			&escrow.RecipientAddress,
			&escrow.ApproverAddress,
			&escrow.Amount,
			&escrow.Commission,
			&escrow.Timeout,
			&escrow.Status,
			&escrow.BlockHeight,
			&escrow.Latest,
			&escrow.Instruction,
		)
		if err != nil {
			return nil, err
		}
		escrows = append(escrows, &escrow)
	}
	return escrows, nil
}

// Scan extract sqlRaw *sql.Row into model.Escrow
func (et *EscrowTransactionQuery) Scan(escrow *model.Escrow, row *sql.Row) error {
	return row.Scan(
		&escrow.ID,
		&escrow.SenderAddress,
		&escrow.RecipientAddress,
		&escrow.ApproverAddress,
		&escrow.Amount,
		&escrow.Commission,
		&escrow.Timeout,
		&escrow.Status,
		&escrow.BlockHeight,
		&escrow.Latest,
		&escrow.Instruction,
	)
}

// Rollback delete records `WHERE height > "height"
func (et *EscrowTransactionQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", et.getTableName()),
			height,
		},
		{
			fmt.Sprintf(`
			UPDATE %s SET latest = ?
			WHERE latest = ? AND (block_height || '_' || id) IN (
				SELECT (MAX(block_height) || '_' || id) as prev
				FROM %s
				GROUP BY id
			)`,
				et.TableName,
				et.TableName,
			),
			1,
			0,
		},
	}
}

func (et *EscrowTransactionQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE block_height >= %d AND block_height <= %d AND latest = 1 ORDER BY block_height DESC",
		strings.Join(et.Fields, ", "),
		et.getTableName(),
		fromHeight,
		toHeight,
	)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (et *EscrowTransactionQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d`,
		et.TableName, fromHeight, toHeight)
}
