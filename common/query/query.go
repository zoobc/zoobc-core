package query

import "github.com/zoobc/zoobc-core/common/chaintype"

type (
	// DerivedQuery represent query that can be rolled back
	DerivedQuery interface {
		// Rollback return query string to rollback table to `height`
		Rollback(height uint32) (multiQueries [][]interface{})
	}
	SnapshotQuery interface {
		SelectDataForSnapshot(fromHeight, toHeight uint32) string
		TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string
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
			NewEscrowTransactionQuery(),
			NewPendingTransactionQuery(),
			NewPendingSignatureQuery(),
			NewMultisignatureInfoQuery(),
			NewSpineBlockManifestQuery(),
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

// GetSnapshotQuery func to get all query repos that have a SelectDataForSnapshot method
func GetSnapshotQuery(ct chaintype.ChainType) (snapshotQuery map[string]SnapshotQuery) {
	switch ct.(type) {
	case *chaintype.MainChain:
		snapshotQuery = map[string]SnapshotQuery{
			"block":              NewBlockQuery(ct),
			"accountBalance":     NewAccountBalanceQuery(),
			"nodeRegistration":   NewNodeRegistrationQuery(),
			"accountDataset":     NewAccountDatasetsQuery(),
			"participationScore": NewParticipationScoreQuery(),
			"publishedReceipt":   NewPublishedReceiptQuery(),
			"escrowTransaction":  NewEscrowTransactionQuery(),
			"pendingTransaction": NewPendingTransactionQuery(),
			"pendingSignature":   NewPendingSignatureQuery(),
			"multisignatureInfo": NewMultisignatureInfoQuery(),
		}
	default:
		snapshotQuery = map[string]SnapshotQuery{}
	}
	return snapshotQuery
}
