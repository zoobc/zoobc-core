package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	SpinePublicKeyQueryInterface interface {
		InsertSpinePublicKey(spinePublicKey *model.SpinePublicKey) [][]interface{}
		GetValidSpinePublicKeysByHeight(height uint32) string
		GetSpinePublicKeyByNodePublicKey(nodePublicKey []byte) (str string, args []interface{})
		ExtractModel(spk *model.SpinePublicKey) []interface{}
		BuildModel(spinePublicKeys []*model.SpinePublicKey, rows *sql.Rows) ([]*model.SpinePublicKey, error)
		BuildBlocksmith(blocksmiths []*model.Blocksmith, rows *sql.Rows) ([]*model.Blocksmith, error)
		Scan(spk *model.SpinePublicKey, row *sql.Row) error
	}

	SpinePublicKeyQuery struct {
		Fields    []string
		TableName string
	}
)

func NewSpinePublicKeyQuery() *SpinePublicKeyQuery {
	return &SpinePublicKeyQuery{
		Fields: []string{
			"node_public_key",
			"public_key_action",
			"latest",
			"height",
		},
		TableName: "spine_public_key",
	}
}

func (spkq *SpinePublicKeyQuery) getTableName() string {
	return spkq.TableName
}

// InsertSpinePublicKey
func (spkq *SpinePublicKeyQuery) InsertSpinePublicKey(spinePublicKey *model.SpinePublicKey) [][]interface{} {
	var (
		queries [][]interface{}
	)
	qryUpdate := fmt.Sprintf("UPDATE %s SET latest = 0 WHERE node_public_key = ?", spkq.getTableName())
	qryInsert := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		spkq.getTableName(),
		strings.Join(spkq.Fields, ","),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(spkq.Fields)-1)),
	)

	queries = append(queries,
		append([]interface{}{qryUpdate}, spinePublicKey.NodePublicKey),
		append([]interface{}{qryInsert}, spkq.ExtractModel(spinePublicKey)...),
	)

	return queries
}

func (spkq *SpinePublicKeyQuery) GetValidSpinePublicKeysByHeight(height uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE height <= %d AND public_key_action=%d AND latest=1",
		strings.Join(spkq.Fields, ", "), spkq.getTableName(), height, uint32(model.SpinePublicKeyAction_AddKey))
}

// GetSpinePublicKeyByNodePublicKey returns query string to get Node Registration by node public key
func (spkq *SpinePublicKeyQuery) GetSpinePublicKeyByNodePublicKey(nodePublicKey []byte) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1",
		strings.Join(spkq.Fields, ", "), spkq.getTableName())
	return query, []interface{}{
		nodePublicKey,
	}
}

// ExtractModel extract the model struct fields to the order of SpinePublicKeyQuery.Fields
func (spkq *SpinePublicKeyQuery) ExtractModel(spk *model.SpinePublicKey) []interface{} {
	return []interface{}{
		spk.NodePublicKey,
		spk.PublicKeyAction,
		spk.Latest,
		spk.Height,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (spkq *SpinePublicKeyQuery) BuildModel(
	spinePublicKeys []*model.SpinePublicKey,
	rows *sql.Rows,
) ([]*model.SpinePublicKey, error) {
	for rows.Next() {
		var (
			spk model.SpinePublicKey
			err error
		)
		err = rows.Scan(
			&spk.NodePublicKey,
			&spk.PublicKeyAction,
			&spk.Latest,
			&spk.Height,
		)
		if err != nil {
			return nil, err
		}
		spinePublicKeys = append(spinePublicKeys, &spk)
	}
	return spinePublicKeys, nil
}

func (spkq *SpinePublicKeyQuery) BuildBlocksmith(
	blocksmiths []*model.Blocksmith, rows *sql.Rows,
) ([]*model.Blocksmith, error) {
	for rows.Next() {
		var (
			blocksmith model.Blocksmith
			nodeStatus int64
			height     uint32
			latest     bool
		)
		err := rows.Scan(
			&blocksmith.NodePublicKey,
			&nodeStatus,
			&latest,
			&height,
		)
		if err != nil {
			return nil, err
		}
		blocksmith.Chaintype = &chaintype.SpineChain{}
		blocksmiths = append(blocksmiths, &blocksmith)
	}
	return blocksmiths, nil
}

// Rollback delete records `WHERE block_height > `height`
// and UPDATE latest of the `account_address` clause by `block_height`
func (spkq *SpinePublicKeyQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE height > ?", spkq.TableName),
			height,
		},
		{
			fmt.Sprintf(`
			UPDATE %s SET latest = ?
			WHERE latest = ? AND (height || '_' || node_public_key) IN (
				SELECT (MAX(height) || '_' || node_public_key) as con
				FROM %s
				GROUP BY node_public_key
			)`,
				spkq.TableName,
				spkq.TableName,
			),
			1, 0,
		},
	}
}

// Scan represents `sql.Scan`
func (spkq *SpinePublicKeyQuery) Scan(spk *model.SpinePublicKey, row *sql.Row) error {
	var (
		err error
	)
	err = row.Scan(
		&spk.NodePublicKey,
		&spk.PublicKeyAction,
		&spk.Latest,
		&spk.Height,
	)
	if err != nil {
		return err
	}
	return nil
}
