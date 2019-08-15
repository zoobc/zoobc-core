package service

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/query"
)

func TestNewMempoolTransactionsService(t *testing.T) {
	type args struct {
		queryExecutor query.ExecutorInterface
	}
	tests := []struct {
		name string
		args args
		want *MempoolTransactionService
	}{
		{
			name: "NewMempoolTranscationService",
			args: args{
				queryExecutor: &query.Executor{},
			},
			want: &MempoolTransactionService{
				Query: &query.Executor{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMempoolTransactionsService(tt.args.queryExecutor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMempoolTransactionsService() = %v, want %v", got, tt.want)
			}
		})
	}
}
