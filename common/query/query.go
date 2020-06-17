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
	// PruneQuery represent query to delete the prunable data from manage table
	PruneQuery interface {
		PruneData(blockHeight, limit uint32) (qStr string, args []interface{})
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
			NewFeeScaleQuery(),
			NewFeeVoteCommitmentVoteQuery(),
			NewFeeVoteRevealVoteQuery(),
		}
		derivedQuery = append(derivedQuery, mainchainDerivedQuery...)
	case *chaintype.SpineChain:
		spinechainDerivedQuery := []DerivedQuery{
			NewSpinePublicKeyQuery(),
			NewSpineBlockManifestQuery(),
		}
		derivedQuery = append(derivedQuery, spinechainDerivedQuery...)
	}
	return derivedQuery
}

// GetSnapshotQuery func to get all query repos that have a SelectDataForSnapshot method (that have data to be included in snapshots)
func GetSnapshotQuery(ct chaintype.ChainType) (snapshotQuery map[string]SnapshotQuery) {
	switch ct.(type) {
	case *chaintype.MainChain:
		snapshotQuery = map[string]SnapshotQuery{
			"block":                    NewBlockQuery(ct),
			"accountBalance":           NewAccountBalanceQuery(),
			"nodeRegistration":         NewNodeRegistrationQuery(),
			"accountDataset":           NewAccountDatasetsQuery(),
			"participationScore":       NewParticipationScoreQuery(),
			"publishedReceipt":         NewPublishedReceiptQuery(),
			"escrowTransaction":        NewEscrowTransactionQuery(),
			"pendingTransaction":       NewPendingTransactionQuery(),
			"pendingSignature":         NewPendingSignatureQuery(),
			"multisignatureInfo":       NewMultisignatureInfoQuery(),
			"skippedBlocksmith":        NewSkippedBlocksmithQuery(),
			"feeScale":                 NewFeeScaleQuery(),
			"feeVoteCommit":            NewFeeVoteCommitmentVoteQuery(),
			"feeVoteReveal":            NewFeeVoteRevealVoteQuery(),
			"liquidPaymentTransaction": NewLiquidPaymentTransactionQuery(),
		}
	default:
		snapshotQuery = map[string]SnapshotQuery{}
	}
	return snapshotQuery
}

// GetBlocksmithSafeQuery func to get all query repos that must save their full history in snapshots,
// for a minRollbackHeight number of blocks, to not break blocksmith process logic
func GetBlocksmithSafeQuery(ct chaintype.ChainType) (snapshotQuery map[string]bool) {
	switch ct.(type) {
	case *chaintype.MainChain:
		snapshotQuery = map[string]bool{
			"block":            true,
			"nodeRegistration": true,
			"publishedReceipt": true,
		}
	default:
		snapshotQuery = map[string]bool{}
	}
	return snapshotQuery
}

// GetPruneQuery func to get all query that have PruneData method. Query to delete prunable data
func GetPruneQuery(ct chaintype.ChainType) (pruneQuery []PruneQuery) {
	switch ct.(type) {
	case *chaintype.MainChain:
		pruneQuery = []PruneQuery{
			NewNodeReceiptQuery(),
			NewMerkleTreeQuery(),
		}
	default:
		pruneQuery = []PruneQuery{}
	}
	return pruneQuery
}
