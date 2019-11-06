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
	tests := []struct {
		name string
		args args
		want []DerivedQuery
	}{
		{
			name: "wantDerivedQuery",
			args: args{chainType: chaintype.GetChainType(0)},
			want: []DerivedQuery{
				NewBlockQuery(chaintype.GetChainType(0)),
				NewTransactionQuery(chaintype.GetChainType(0)),
				NewNodeRegistrationQuery(),
				NewAccountBalanceQuery(),
				NewAccountDatasetsQuery(),
				NewSkippedBlocksmithQuery(),
				NewParticipationScoreQuery(),
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
