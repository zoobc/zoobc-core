// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package query

import (
	"math"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
)

type (
	// DerivedQuery represent query that can be rolled back
	DerivedQuery interface {
		// Rollback return query string to rollback table to `height`
		Rollback(height uint32) (multiQueries [][]interface{})
	}
	SnapshotQuery interface {
		SelectDataForSnapshot(fromHeight, toHeight uint32) string
		TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string
		ImportSnapshot(interface{}) ([][]interface{}, error)
		RecalibrateVersionedTable() []string
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
			NewNodeAdmissionTimestampQuery(),
			NewMultiSignatureParticipantQuery(),
			NewBatchReceiptQuery(),
			NewMerkleTreeQuery(),
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
			"nodeAdmissionTimestamp":   NewNodeAdmissionTimestampQuery(),
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
			NewBatchReceiptQuery(),
			NewMerkleTreeQuery(),
		}
	default:
		pruneQuery = []PruneQuery{}
	}
	return pruneQuery
}

// CalculateBulkSize calculating max records might allowed in single sqlite transaction, since sqlite3 has maximum
// variables in single transactions called SQLITE_LIMIT_VARIABLE_NUMBER in sqlite3-binding.c which is 999
func CalculateBulkSize(totalFields, totalRecords int) (recordsPerPeriod, rounds, remaining int) {
	perPeriod := math.Floor(float64(constant.SQLiteLimitVariableNumber) / float64(totalFields))
	rounds = int(math.Floor(float64(totalRecords) / perPeriod))

	if perPeriod == 0 || rounds == 0 {
		return totalRecords, 1, 0
	}
	remaining = totalRecords % (rounds * int(perPeriod))
	return int(perPeriod), rounds, remaining
}
