package query

import "github.com/zoobc/zoobc-core/common/chaintype"

type (
	// DerivedQuery represent query that can be rolled back
	DerivedQuery interface {
		// Rollback return query string to rollback table to `height`
		Rollback(height uint32) (multiQueries [][]interface{})
	}
)

// GetDerivedQuery func to get the whole queries has has rollback method
func GetDerivedQuery(chainType chaintype.ChainType) []DerivedQuery {
	return []DerivedQuery{
		NewBlockQuery(chainType),
		NewTransactionQuery(chainType),
		NewNodeRegistrationQuery(),
		NewAccountBalanceQuery(),
		NewAccountDatasetsQuery(),
		NewSkippedBlocksmithQuery(),
		NewParticipationScoreQuery(),
	}
}
