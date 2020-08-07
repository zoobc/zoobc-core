package query

import (
	"database/sql"
	"fmt"
	"github.com/zoobc/zoobc-core/common/blocker"
	"strings"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// LiquidPaymentTransactionQuery fields must have
	LiquidPaymentTransactionQuery struct {
		Fields    []string
		TableName string
	}

	// LiquidPaymentTransactionQueryInterface methods must have
	LiquidPaymentTransactionQueryInterface interface {
		InsertLiquidPaymentTransaction(liquidPayment *model.LiquidPayment) [][]interface{}
		InsertLiquidPaymentTransactions(liquidPayments []*model.LiquidPayment) (str string, args []interface{})
		GetPendingLiquidPaymentTransactionByID(id int64, status model.LiquidPaymentStatus) (str string, args []interface{})
		GetPassedTimePendingLiquidPaymentTransactions(timestamp int64) (qStr string, args []interface{})
		CompleteLiquidPaymentTransaction(id int64, causedFields map[string]interface{}) [][]interface{}
		ExtractModel(*model.LiquidPayment) []interface{}
		BuildModels(*sql.Rows) ([]*model.LiquidPayment, error)
		Scan(liquidPayment *model.LiquidPayment, row *sql.Row) error
	}
)

// NewLiquidPaymentTransactionQuery build a LiquidPaymentTransactionQuery
func NewLiquidPaymentTransactionQuery() *LiquidPaymentTransactionQuery {
	return &LiquidPaymentTransactionQuery{
		Fields: []string{
			"id",
			"sender_address",
			"recipient_address",
			"amount",
			"applied_time",
			"complete_minutes",
			"status",
			"block_height",
			"latest",
		},
		TableName: "liquid_payment_transaction",
	}
}

func (lpt *LiquidPaymentTransactionQuery) getTableName() string {
	return lpt.TableName
}

func (lpt *LiquidPaymentTransactionQuery) InsertLiquidPaymentTransaction(liquidPayment *model.LiquidPayment) [][]interface{} {
	liquidPaymentTobeWritten := liquidPayment
	liquidPaymentTobeWritten.Latest = true
	return [][]interface{}{
		{
			fmt.Sprintf(
				"UPDATE %s set latest = ? WHERE id = ?",
				lpt.getTableName(),
			),
			false,
			liquidPaymentTobeWritten.GetID(),
		},
		append(
			[]interface{}{
				fmt.Sprintf(
					"INSERT INTO %s (%s) VALUES(%s)",
					lpt.getTableName(),
					strings.Join(lpt.Fields, ","),
					fmt.Sprintf("? %s", strings.Repeat(", ?", len(lpt.Fields)-1))),
			},
			lpt.ExtractModel(liquidPaymentTobeWritten)...,
		),
	}
}

func (lpt *LiquidPaymentTransactionQuery) InsertLiquidPaymentTransactions(liquidPayments []*model.LiquidPayment) (str string, args []interface{}) {
	if len(liquidPayments) > 0 {
		str = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			lpt.getTableName(),
			strings.Join(lpt.Fields, ", "),
		)
		for k, liquidPayment := range liquidPayments {
			str += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(lpt.Fields)-1),
			)
			if k < len(liquidPayments)-1 {
				str += ","
			}
			args = append(args, lpt.ExtractModel(liquidPayment)...)
		}

	}
	return str, args
}

// ImportSnapshot takes payload from downloaded snapshot and insert them into database
func (lpt *LiquidPaymentTransactionQuery) ImportSnapshot(payload interface{}) ([][]interface{}, error) {
	var (
		queries [][]interface{}
	)
	liquidPayments, ok := payload.([]*model.LiquidPayment)
	if !ok {
		return nil, blocker.NewBlocker(blocker.DBErr, "ImportSnapshotCannotCastTo"+lpt.TableName)
	}
	if len(liquidPayments) > 0 {
		recordsPerPeriod, rounds, remaining := CalculateBulkSize(len(lpt.Fields), len(liquidPayments))
		for i := 0; i < rounds; i++ {
			qry, args := lpt.InsertLiquidPaymentTransactions(liquidPayments[i*recordsPerPeriod : (i*recordsPerPeriod)+recordsPerPeriod])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
		if remaining > 0 {
			qry, args := lpt.InsertLiquidPaymentTransactions(liquidPayments[len(liquidPayments)-remaining:])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
	}
	return queries, nil
}

// RecalibrateVersionedTable recalibrate table to clean up multiple latest rows due to import function
func (lpt *LiquidPaymentTransactionQuery) RecalibrateVersionedTable() []string {
	return []string{
		fmt.Sprintf(
			"update %s set latest = false where latest = true AND (id, block_height) NOT IN "+
				"(select t2.id, max(t2.block_height) from %s t2 group by t2.id)",
			lpt.getTableName(), lpt.getTableName()),
		fmt.Sprintf(
			"update %s set latest = true where latest = false AND (id, block_height) IN "+
				"(select t2.id, max(t2.block_height) from %s t2 group by t2.id)",
			lpt.getTableName(), lpt.getTableName()),
	}
}

func (lpt *LiquidPaymentTransactionQuery) CompleteLiquidPaymentTransaction(id int64, causedFields map[string]interface{}) [][]interface{} {
	return [][]interface{}{
		{
			fmt.Sprintf(
				"INSERT INTO %s (id, sender_address, recipient_address, amount, applied_time, complete_minutes, status, block_height, latest)"+
					" SELECT id, sender_address, recipient_address, amount, applied_time, complete_minutes, ?, %d, true FROM %s WHERE id = %d AND latest = 1"+
					" ON CONFLICT(id, block_height) DO UPDATE SET status = ?",
				lpt.getTableName(),
				causedFields["block_height"],
				lpt.getTableName(),
				id,
			),
			model.LiquidPaymentStatus_LiquidPaymentCompleted,
			model.LiquidPaymentStatus_LiquidPaymentCompleted,
		},
		{
			fmt.Sprintf(
				"UPDATE %s set latest = ? WHERE id = ? AND block_height != %d and latest = true",
				lpt.getTableName(),
				causedFields["block_height"],
			),
			false,
			id,
		},
	}
}

// GetPendingLiquidPaymentTransactionByID fetches the latest Liquid payment record that matches with the ID and have pending status
func (lpt *LiquidPaymentTransactionQuery) GetPendingLiquidPaymentTransactionByID(id int64,
	status model.LiquidPaymentStatus) (str string, args []interface{}) {
	return fmt.Sprintf(
			"SELECT %s FROM %s WHERE id = ? AND status = ? AND latest = ?",
			strings.Join(lpt.Fields, ", "),
			lpt.getTableName(),
		),
		[]interface{}{id, status, true}
}

func (lpt *LiquidPaymentTransactionQuery) GetPassedTimePendingLiquidPaymentTransactions(timestamp int64) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"SELECT %s FROM %s WHERE applied_time+(complete_minutes*%d) <= ? AND status = ? AND latest = ?",
			strings.Join(lpt.Fields, ", "),
			lpt.getTableName(),
			constant.CompleteMinutesUnit,
		),
		[]interface{}{timestamp, model.LiquidPaymentStatus_LiquidPaymentPending, true}
}

// ExtractModel will extract values of LiquidPayment as []interface{}
func (lpt *LiquidPaymentTransactionQuery) ExtractModel(liquidPayment *model.LiquidPayment) []interface{} {
	return []interface{}{
		liquidPayment.GetID(),
		liquidPayment.GetSenderAddress(),
		liquidPayment.GetRecipientAddress(),
		liquidPayment.GetAmount(),
		liquidPayment.GetAppliedTime(),
		liquidPayment.GetCompleteMinutes(),
		liquidPayment.GetStatus(),
		liquidPayment.GetBlockHeight(),
		liquidPayment.GetLatest(),
	}
}

// BuildModels extract sqlRaw into []*model.LiquidPayment
func (lpt *LiquidPaymentTransactionQuery) BuildModels(rows *sql.Rows) ([]*model.LiquidPayment, error) {
	var (
		liquidPayments []*model.LiquidPayment
		err            error
	)

	for rows.Next() {
		var liquidPayment model.LiquidPayment
		err = rows.Scan(
			&liquidPayment.ID,
			&liquidPayment.SenderAddress,
			&liquidPayment.RecipientAddress,
			&liquidPayment.Amount,
			&liquidPayment.AppliedTime,
			&liquidPayment.CompleteMinutes,
			&liquidPayment.Status,
			&liquidPayment.BlockHeight,
			&liquidPayment.Latest,
		)
		if err != nil {
			return nil, err
		}
		liquidPayments = append(liquidPayments, &liquidPayment)
	}
	return liquidPayments, nil
}

// Scan extract sqlRaw *sql.Row into model.LiquidPayment
func (lpt *LiquidPaymentTransactionQuery) Scan(liquidPayment *model.LiquidPayment, row *sql.Row) error {
	return row.Scan(
		&liquidPayment.ID,
		&liquidPayment.SenderAddress,
		&liquidPayment.RecipientAddress,
		&liquidPayment.Amount,
		&liquidPayment.AppliedTime,
		&liquidPayment.CompleteMinutes,
		&liquidPayment.Status,
		&liquidPayment.BlockHeight,
		&liquidPayment.Latest,
	)
}

// Rollback delete records `WHERE height > "height"
func (lpt *LiquidPaymentTransactionQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", lpt.getTableName()),
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
				lpt.TableName,
				lpt.TableName,
			),
			1,
			0,
		},
	}
}

func (lpt *LiquidPaymentTransactionQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE (id, block_height) IN (SELECT t2.id, MAX(t2.block_height) FROM %s as t2 "+
			"WHERE t2.block_height >= %d AND t2.block_height <= %d AND t2.block_height != 0 GROUP BY t2.id) ORDER BY block_height",
		strings.Join(lpt.Fields, ","),
		lpt.getTableName(),
		lpt.getTableName(),
		fromHeight,
		toHeight,
	)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (lpt *LiquidPaymentTransactionQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0`,
		lpt.TableName, fromHeight, toHeight)
}
