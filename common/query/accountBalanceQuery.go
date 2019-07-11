package query

import (
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	AccountBalanceQuery struct {
		Fields    []string
		TableName string
		ChainType contract.ChainType
	}
	AccountBalanceInt interface {
		GetAccountBalanceByAccountID(accountID []byte) (*model.AccountBalance, error)
	}
)

func NewAccountBalanceQuery(chaintype contract.ChainType) *AccountBalanceQuery {
	return &AccountBalanceQuery{
		Fields: []string{
			"account_id",
			"block_height",
			"spendable_balance",
			"balance",
			"pop_revenue",
		},
		TableName: "account_balance",
		ChainType: chaintype,
	}
}
func (q *AccountBalanceQuery) GetAccountBalanceByAccountID() string {
	return fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE account_id = ? 
	`, strings.Join(q.Fields, ","), q.TableName)
}
