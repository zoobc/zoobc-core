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
			},
		},
		{
			name: "wantDerivedQuery:spinechain",
			args: args{chainType: spinechain},
			want: []DerivedQuery{
				NewBlockQuery(spinechain),
				NewSpinePublicKeyQuery(),
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
