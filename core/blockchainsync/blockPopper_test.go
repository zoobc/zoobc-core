package blockchainsync

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/service"
)

type (
	mockPopOffToBlockReturnCommonBlock struct {
		query.Executor
	}
	mockPopOffToBlockReturnBeginTxFunc struct {
		query.Executor
	}
	mockPopOffToBlockReturnWantFailOnCommit struct {
		query.Executor
	}
	mockPopOffToBlockReturnWantFailOnExecuteTransactions struct {
		query.Executor
	}
)

func (*mockPopOffToBlockReturnCommonBlock) BeginTx() error {
	return nil
}
func (*mockPopOffToBlockReturnCommonBlock) CommitTx() error {
	return nil
}
func (*mockPopOffToBlockReturnCommonBlock) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}
func (*mockPopOffToBlockReturnCommonBlock) ExecuteSelect(qSrt string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(query.NewMempoolQuery(chaintype.GetChainType(0)).Fields).AddRow(
			1,
			0,
			10,
			1000,
			[]byte{2, 0, 0, 0, 1, 112, 240, 249, 74, 0, 0, 0, 0, 44, 0, 0, 0, 66, 67, 90, 69, 71, 79, 98, 51, 87, 78, 120, 51,
				102, 68, 79, 86, 102, 57, 90, 83, 52, 69, 106, 118, 79, 73, 118, 95, 85, 101, 87, 52, 84, 86, 66, 81, 74, 95, 54,
				116, 72, 75, 108, 69, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 201, 0, 0, 0, 153, 58, 50, 200, 7, 61,
				108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213,
				135, 0, 0, 0, 0, 9, 0, 0, 0, 49, 50, 55, 46, 48, 46, 48, 46, 49, 0, 202, 154, 59, 0, 0, 0, 0, 86, 90, 118, 89, 100,
				56, 48, 112, 53, 83, 45, 114, 120, 83, 78, 81, 109, 77, 90, 119, 89, 88, 67, 55, 76, 121, 65, 122, 66, 109, 99, 102,
				99, 106, 52, 77, 85, 85, 65, 100, 117, 100, 87, 77, 198, 224, 91, 94, 235, 56, 96, 236, 211, 155, 119, 159, 171, 196,
				10, 175, 144, 215, 90, 167, 3, 27, 88, 212, 233, 202, 31, 112, 45, 147, 34, 18, 1, 0, 0, 0, 48, 128, 236, 38, 196, 0,
				66, 232, 114, 70, 30, 220, 206, 222, 141, 50, 152, 151, 150, 235, 72, 86, 150, 96, 70, 162, 253, 128, 108, 95, 26, 175,
				178, 108, 74, 76, 98, 68, 141, 131, 57, 209, 224, 251, 129, 224, 47, 156, 120, 9, 77, 251, 236, 230, 212, 109, 193, 67,
				250, 166, 49, 249, 198, 11, 0, 0, 0, 0, 162, 190, 223, 52, 221, 118, 195, 111, 129, 166, 99, 216, 213, 202, 203, 118, 28,
				231, 39, 137, 123, 228, 86, 52, 100, 8, 124, 254, 19, 181, 202, 139, 211, 184, 202, 54, 8, 166, 131, 96, 244, 101, 76,
				167, 176, 172, 85, 88, 93, 32, 173, 123, 229, 109, 128, 26, 192, 70, 155, 217, 107, 210, 254, 15},
			"BCZ",
			"ZCB",
		),
	)
	return db.Query("")
}
func (*mockPopOffToBlockReturnCommonBlock) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}
func (*mockPopOffToBlockReturnBeginTxFunc) BeginTx() error {
	return errors.New("i want this")
}
func (*mockPopOffToBlockReturnBeginTxFunc) CommitTx() error {
	return nil
}
func (*mockPopOffToBlockReturnWantFailOnCommit) BeginTx() error {
	return nil
}
func (*mockPopOffToBlockReturnWantFailOnCommit) CommitTx() error {
	return errors.New("i want this")
}
func (*mockPopOffToBlockReturnWantFailOnCommit) ExecuteSelect(qSrt string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(query.NewMempoolQuery(chaintype.GetChainType(0)).Fields).AddRow(
			1,
			0,
			10,
			1000,
			[]byte{1, 2, 3, 4, 5},
			"BCZ",
			"ZCB",
		),
	)
	return db.Query("")

}
func (*mockPopOffToBlockReturnWantFailOnExecuteTransactions) BeginTx() error {
	return nil
}
func (*mockPopOffToBlockReturnWantFailOnExecuteTransactions) CommitTx() error {
	return nil
}
func (*mockPopOffToBlockReturnWantFailOnExecuteTransactions) ExecuteTransactions(queries [][]interface{}) error {
	return errors.New("i want this")
}
func (*mockPopOffToBlockReturnWantFailOnExecuteTransactions) RollbackTx() error {
	return nil
}

func TestService_PopOffToBlock(t *testing.T) {

	type fields struct {
		BlockService       service.BlockServiceInterface
		MempoolService     service.MempoolServiceInterface
		QueryExecutor      query.ExecutorInterface
		ChainType          chaintype.ChainType
		ActionTypeSwitcher transaction.TypeActionSwitcher
		KVDB               kvdb.KVExecutorInterface
	}
	type args struct {
		commonBlock *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Block
		wantErr bool
	}{
		{
			name: "want:TestService_PopOffToBlock error on getting LastBlock",
			fields: fields{
				BlockService:  &mockServiceBlockFailGetLastBlock{},
				ChainType:     &mockServiceChainType{},
				QueryExecutor: &mockServiceQueryExecutor{},
			},
			args: args{
				commonBlock: &model.Block{
					ID:     70,
					Height: 500,
				},
			},
			want:    []*model.Block{},
			wantErr: true,
		},
		{
			name: "want:TestService_PopOffToBlock error on getting BlockByHeight",
			fields: fields{
				BlockService:  &mockServiceBlockFailGetBlockByHeight{},
				ChainType:     &mockServiceChainType{},
				QueryExecutor: &mockServiceQueryExecutor{},
			},
			args: args{
				commonBlock: &model.Block{
					ID:     70,
					Height: 500,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "want:TestService_PopOffToBlock error on getting BlockByID",
			fields: fields{
				BlockService:  &mockServiceBlockFailGetBlockByID{},
				ChainType:     &mockServiceChainType{},
				QueryExecutor: &mockServiceQueryExecutor{},
			},
			args: args{
				commonBlock: &model.Block{
					ID:     70,
					Height: 500,
				},
			},
			want:    []*model.Block{},
			wantErr: true,
		},
		// SERVICE QUERY SERVICES FAIL
		{
			name: "want:TestService_PopOffToBlock error on BeginTx function",
			fields: fields{
				BlockService:  &mockServiceBlockSuccess{},
				ChainType:     &mockServiceChainType{},
				QueryExecutor: &mockPopOffToBlockReturnBeginTxFunc{},
				MempoolService: service.NewMempoolService(
					chaintype.GetChainType(0),
					kvdb.NewMockKVExecutorInterface(gomock.NewController(t)),
					&mockPopOffToBlockReturnCommonBlock{},
					query.NewMempoolQuery(chaintype.GetChainType(0)),
					nil,
					&transaction.TypeSwitcher{Executor: &mockPopOffToBlockReturnCommonBlock{}},
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
			},
			args: args{
				commonBlock: &model.Block{
					ID:     70,
					Height: 500,
				},
			},
			want:    []*model.Block{},
			wantErr: true,
		},
		{
			name: "want:TestService_PopOffToBlock error when committing transaction",
			fields: fields{
				BlockService:  &mockServiceBlockSuccess{},
				ChainType:     &mockServiceChainType{},
				QueryExecutor: &mockServiceQueryExecutorCommitTXFail{},
				MempoolService: service.NewMempoolService(
					chaintype.GetChainType(0),
					kvdb.NewMockKVExecutorInterface(gomock.NewController(t)),
					&mockPopOffToBlockReturnWantFailOnCommit{},
					query.NewMempoolQuery(chaintype.GetChainType(0)),
					nil,
					&transaction.TypeSwitcher{Executor: &mockPopOffToBlockReturnWantFailOnCommit{}},
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
			},
			args: args{
				commonBlock: &model.Block{
					ID:     70,
					Height: 500,
				},
			},
			want:    []*model.Block{},
			wantErr: true,
		},
		{
			name: "want:TestService_PopOffToBlock error when executing Transactions",
			fields: fields{
				BlockService:  &mockServiceBlockSuccess{},
				ChainType:     &mockServiceChainType{},
				QueryExecutor: &mockPopOffToBlockReturnWantFailOnExecuteTransactions{},
				MempoolService: service.NewMempoolService(
					chaintype.GetChainType(0),
					kvdb.NewMockKVExecutorInterface(gomock.NewController(t)),
					&mockPopOffToBlockReturnWantFailOnCommit{},
					query.NewMempoolQuery(chaintype.GetChainType(0)),
					nil,
					&transaction.TypeSwitcher{Executor: &mockPopOffToBlockReturnWantFailOnCommit{}},
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
				ActionTypeSwitcher: &transaction.TypeSwitcher{Executor: &mockPopOffToBlockReturnCommonBlock{}},
				KVDB:               kvdb.NewMockKVExecutorInterface(gomock.NewController(t)),
			},
			args: args{
				commonBlock: &model.Block{
					ID:     70,
					Height: 500,
				},
			},
			want:    []*model.Block{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BlockPopper{
				ChainType:          tt.fields.ChainType,
				BlockService:       tt.fields.BlockService,
				QueryExecutor:      tt.fields.QueryExecutor,
				MempoolService:     tt.fields.MempoolService,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
				KVDB:               tt.fields.KVDB,
			}
			got, err := bp.PopOffToBlock(tt.args.commonBlock)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.PopOffToBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.PopOffToBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}
