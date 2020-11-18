package service

import (
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/query"
)

func TestNewLiquidTransactionService(t *testing.T) {
	type args struct {
		executor                      query.ExecutorInterface
		liquidPaymentTransactionQuery *query.LiquidPaymentTransactionQuery
	}

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()

	tests := []struct {
		name string
		args args
		want *LiquidTransactionService
	}{
		{
			name: "wantSuccess",
			args: args{
				executor: query.NewQueryExecutor(db),
			},
			want: &LiquidTransactionService{
				QueryExecutor: query.NewQueryExecutor(db),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLiquidTransactionService(tt.args.executor, tt.args.liquidPaymentTransactionQuery); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLiquidTransactionService() = %v, want %v", got, tt.want)
			}
		})
	}
}
