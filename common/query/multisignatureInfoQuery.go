package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	MultisignatureInfoQueryInterface interface {
		GetMultisignatureInfoByAddress(multisigAddress string) (str string, args []interface{})
		InsertMultisignatureInfo(multisigInfo *model.MultiSignatureInfo) (str string, args []interface{})
		Scan(multisigInfo *model.MultiSignatureInfo, row *sql.Row) error
		ExtractModel(multisigInfo *model.MultiSignatureInfo) []interface{}
		BuildModel(multisigInfos []*model.MultiSignatureInfo, rows *sql.Rows) ([]*model.MultiSignatureInfo, error)
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
			"addresses",
			"block_height",
		},
		TableName: "multisignature_info",
	}
}

func (msi *MultisignatureInfoQuery) getTableName() string {
	return msi.TableName
}

func (msi *MultisignatureInfoQuery) GetMultisignatureInfoByAddress(multisigAddress string) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE multisig_address = ?", strings.Join(msi.Fields, ", "), msi.getTableName())
	return query, []interface{}{
		multisigAddress,
	}
}

// InsertPendingSignature inserts a new pending transaction into DB
func (msi *MultisignatureInfoQuery) InsertMultisignatureInfo(multisigInfo *model.MultiSignatureInfo) (str string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		msi.getTableName(),
		strings.Join(msi.Fields, ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ? ", len(msi.Fields)-1)),
	), msi.ExtractModel(multisigInfo)
}

func (*MultisignatureInfoQuery) Scan(multisigInfo *model.MultiSignatureInfo, row *sql.Row) error {
	var addresses string
	err := row.Scan(
		&multisigInfo.MultisigAddress,
		&multisigInfo.MinimumSignatures,
		&multisigInfo.Nonce,
		&addresses,
		&multisigInfo.BlockHeight,
	)
	multisigInfo.Addresses = strings.Split(addresses, ", ")
	return err
}

func (*MultisignatureInfoQuery) ExtractModel(multisigInfo *model.MultiSignatureInfo) []interface{} {
	addresses := strings.Join(multisigInfo.Addresses, ", ")
	return []interface{}{
		&multisigInfo.MultisigAddress,
		&multisigInfo.MinimumSignatures,
		&multisigInfo.Nonce,
		addresses,
		&multisigInfo.BlockHeight,
	}
}

func (msi *MultisignatureInfoQuery) BuildModel(
	mss []*model.MultiSignatureInfo, rows *sql.Rows,
) ([]*model.MultiSignatureInfo, error) {
	for rows.Next() {
		var (
			multisigInfo model.MultiSignatureInfo
			addresses    string
		)
		err := rows.Scan(
			&multisigInfo.MultisigAddress,
			&multisigInfo.MinimumSignatures,
			&multisigInfo.Nonce,
			&addresses,
			&multisigInfo.BlockHeight,
		)
		multisigInfo.Addresses = strings.Split(addresses, ", ")
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
	}
}
