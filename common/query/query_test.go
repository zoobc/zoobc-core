package query

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
)

func TestGetDerivedQuery(t *testing.T) {
	type args struct {
		chainType chaintype.ChainType
	}
	mainchain := &chaintype.MainChain{}
	spinechain := &chaintype.SpineChain{}
	tests := []struct {
		name string
		args args
		want []DerivedQuery
	}{
		{
			name: "wantDerivedQuery:mainchain",
			args: args{chainType: mainchain},
			want: []DerivedQuery{
				NewBlockQuery(mainchain),
				NewTransactionQuery(mainchain),
				NewNodeRegistrationQuery(),
				NewAccountBalanceQuery(),
				NewAccountDatasetsQuery(),
				NewMempoolQuery(mainchain),
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
				NewAtomicTransactionQuery(),
				NewMerkleTreeQuery(),
			},
		},
		{
			name: "wantDerivedQuery:spinechain",
			args: args{chainType: spinechain},
			want: []DerivedQuery{
				NewBlockQuery(spinechain),
				NewSpinePublicKeyQuery(),
				NewSpineBlockManifestQuery(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDerivedQuery(tt.args.chainType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDerivedQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_CalculateBulkSize(t *testing.T) {
	type args struct {
		totalFields  int
		totalRecords int
	}
	tests := []struct {
		name                 string
		args                 args
		wantRecordsPerPeriod int
		wantRounds           int
		wantRemaining        int
	}{
		{
			name: "WantSuccess",
			args: args{
				totalRecords: 1421,
				totalFields:  12,
			},
			wantRecordsPerPeriod: 83,
			wantRounds:           17,
			wantRemaining:        10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRecordsPerPeriod, gotRounds, gotRemaining := CalculateBulkSize(tt.args.totalFields, tt.args.totalRecords)
			if gotRecordsPerPeriod != tt.wantRecordsPerPeriod {
				t.Errorf("calculateBulkSize() gotRecordsPerPeriod = %v, want %v", gotRecordsPerPeriod, tt.wantRecordsPerPeriod)
			}
			if gotRounds != tt.wantRounds {
				t.Errorf("calculateBulkSize() gotRounds = %v, want %v", gotRounds, tt.wantRounds)
			}
			if gotRemaining != tt.wantRemaining {
				t.Errorf("calculateBulkSize() gotRemaining = %v, want %v", gotRemaining, tt.wantRemaining)
			}
		})
	}
}
