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
		InsertEscrowTransactions(escrows []*model.Escrow) (str string, args []interface{})
		GetLatestEscrowTransactionByID(int64) (string, []interface{})
		GetEscrowTransactions(fields map[string]interface{}) (string, []interface{})
		GetExpiredEscrowTransactionsAtCurrentBlock(blockHeight uint32) string
		GetEscrowTransactionsByTransactionIdsAndStatus(
			transactionIds []string, status model.EscrowStatus,
		) string
		ExpiringEscrowTransactions(blockHeight uint32) (string, []interface{})
		ExtractModel(*model.Escrow) []interface{}
		BuildModels(*sql.Rows) ([]*model.Escrow, error)
		Scan(escrow *model.Escrow, row *sql.Row) error
		GetFields() []string
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

// InsertEscrowTransactions represents query builder to insert multiple record in single query
func (et *EscrowTransactionQuery) InsertEscrowTransactions(escrows []*model.Escrow) (str string, args []interface{}) {
	if len(escrows) > 0 {
		str = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			et.getTableName(),
			strings.Join(et.Fields, ","),
		)
		for k, escrow := range escrows {
			str += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(et.Fields)-1),
			)

			if k < len(escrows)-1 {
				str += ","
			}
			args = append(args, et.ExtractModel(escrow)...)
		}
	}

	return str, args
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

// GetExpiredEscrowTransactionsAtCurrentBlock fetch provided block height expired escrow transaction
func (et *EscrowTransactionQuery) GetExpiredEscrowTransactionsAtCurrentBlock(blockHeight uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE timeout + block_height = %d AND latest = true AND status = %d",
		strings.Join(et.Fields, ", "), et.getTableName(), blockHeight, model.EscrowStatus_Pending)
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

func (et *EscrowTransactionQuery) GetEscrowTransactionsByTransactionIdsAndStatus(
	transactionIds []string, status model.EscrowStatus,
) string {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE id IN (%s) AND status = %d",
		strings.Join(et.Fields, ", "),
		et.getTableName(),
		strings.Join(transactionIds, ", "),
		status,
	)
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

func (et *EscrowTransactionQuery) GetFields() []string {
	return et.Fields
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
			WHERE latest = ? AND (id, block_height) IN (
				SELECT t2.id, MAX(t2.block_height)
				FROM %s as t2
				GROUP BY t2.id
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
	return fmt.Sprintf("SELECT %s FROM %s WHERE (id, block_height) IN (SELECT t2.id, MAX("+
		"t2.block_height) FROM %s as t2 WHERE t2.block_height >= %d AND t2.block_height <= %d GROUP BY t2.id) ORDER BY block_height",
		strings.Join(et.Fields, ","),
		et.getTableName(),
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
