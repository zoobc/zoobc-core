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
func GetDerivedQuery(ct chaintype.ChainType) (derivedQuery []DerivedQuery) {
	derivedQuery = []DerivedQuery{
		NewBlockQuery(ct),
	}
	switch ct.(type) {
	case *chaintype.MainChain:
		mainchainDerivedQuery := []DerivedQuery{
			NewTransactionQuery(ct),
			NewNodeRegistrationQuery(),
			NewAccountBalanceQuery(),
			NewAccountDatasetsQuery(),
			NewMempoolQuery(ct),
			NewSkippedBlocksmithQuery(),
			NewParticipationScoreQuery(),
			NewPublishedReceiptQuery(),
			NewAccountLedgerQuery(),
		}
		derivedQuery = append(derivedQuery, mainchainDerivedQuery...)
	case *chaintype.SpineChain:
		spinechainDerivedQuery := []DerivedQuery{
			NewSpinePublicKeyQuery(),
		}
		derivedQuery = append(derivedQuery, spinechainDerivedQuery...)
	}
	return derivedQuery
}
