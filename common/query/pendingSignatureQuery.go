package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	PendingSignatureQueryInterface interface {
		GetPendingSignatureByHash(txHash []byte) (str string, args []interface{})
		InsertPendingSignature(pendingSig *model.PendingSignature) (str string, args []interface{})
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
			"signature",
			"block_height",
		},
		TableName: "pending_signature",
	}
}

func (psq *PendingSignatureQuery) getTableName() string {
	return psq.TableName
}

func (psq *PendingSignatureQuery) GetPendingSignatureByHash(txHash []byte) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE transaction_hash = ?", strings.Join(psq.Fields, ", "), psq.getTableName())
	return query, []interface{}{
		txHash,
	}
}

// InsertPendingSignature inserts a new pending transaction into DB
func (psq *PendingSignatureQuery) InsertPendingSignature(pendingSig *model.PendingSignature) (str string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		psq.getTableName(),
		strings.Join(psq.Fields, ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ? ", len(psq.Fields)-1)),
	), psq.ExtractModel(pendingSig)
}

func (*PendingSignatureQuery) Scan(pendingSig *model.PendingSignature, row *sql.Row) error {
	err := row.Scan(
		&pendingSig.TransactionHash,
		&pendingSig.AccountAddress,
		&pendingSig.Signature,
		&pendingSig.BlockHeight,
	)
	return err
}

func (*PendingSignatureQuery) ExtractModel(pendingSig *model.PendingSignature) []interface{} {
	return []interface{}{
		&pendingSig.TransactionHash,
		&pendingSig.AccountAddress,
		&pendingSig.Signature,
		&pendingSig.BlockHeight,
	}
}

func (psq *PendingSignatureQuery) BuildModel(
	pss []*model.PendingSignature, rows *sql.Rows,
) ([]*model.PendingSignature, error) {
	for rows.Next() {
		var pendingSig model.PendingSignature
		err := rows.Scan(
			&pendingSig.TransactionHash,
			&pendingSig.AccountAddress,
			&pendingSig.Signature,
			&pendingSig.BlockHeight,
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
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", psq.getTableName()),
			height,
		},
	}
}
