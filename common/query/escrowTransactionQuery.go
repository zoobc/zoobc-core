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
		InsertEscrowTransaction(*model.Escrow) (string, []interface{})
		ExtractModel(*model.Escrow) []interface{}
		BuildModels(*sql.Rows) ([]*model.Escrow, error)
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
		},
	}
}

func (et *EscrowTransactionQuery) getTableName() string {
	return et.TableName
}

// InsertEscrowTransaction represents insert query for escrow_transaction table
func (et *EscrowTransactionQuery) InsertEscrowTransaction(escrow *model.Escrow) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES(%s)",
			et.getTableName(),
			strings.Join(et.Fields, ","),
			fmt.Sprintf("? %s", strings.Repeat(", ?", len(et.Fields)-1)),
		),
		et.ExtractModel(escrow)
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
		)
		if err != nil {
			return nil, err
		}
		escrows = append(escrows, &escrow)
	}
	return escrows, nil
}
