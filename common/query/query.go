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
func GetDerivedQuery(ct chaintype.ChainType) []DerivedQuery {
	switch ct.(type) {
	case *chaintype.MainChain:
		return []DerivedQuery{
			NewBlockQuery(ct),
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
	case *chaintype.SpineChain:
		return []DerivedQuery{
			NewSpinePublicKeyQuery(),
			NewMegablockQuery(),
		}
	}
	return nil
}
