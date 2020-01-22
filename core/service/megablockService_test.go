package service

import (
	"fmt"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockMegablockServiceQueryExecutor struct {
		testName string
		query.Executor
	}
)

var (
	ssMockMegablock = &model.Megablock{
		ID:                  1,
		FullFileHash:        ssMockFullHash,
		MegablockHeight:     720,
		FileChunkHashes:     []byte{},
		ChainType:           0,
		MegablockType:       model.MegablockType_Snapshot,
		ExpirationTimestamp: 1000,
	}
)

func (*mockMegablockServiceQueryExecutor) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}

func (*mockMegablockServiceQueryExecutor) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func (*mockMegablockServiceQueryExecutor) BeginTx() error {
	return nil
}

func (*mockMegablockServiceQueryExecutor) RollbackTx() error {
	return nil
}
func (*mockMegablockServiceQueryExecutor) CommitTx() error {
	return nil
}

func TestBlockSpineSnapshotService_CreateMegablock(t *testing.T) {
	type fields struct {
		QueryExecutor             query.ExecutorInterface
		MegablockQuery            query.MegablockQueryInterface
		SpineBlockQuery           query.BlockQueryInterface
		MainBlockQuery            query.BlockQueryInterface
		Logger                    *log.Logger
		Spinechain                chaintype.ChainType
		Mainchain                 chaintype.ChainType
		SnapshotInterval          int64
		SnapshotGenerationTimeout int64
	}
	type args struct {
		snapshotHash            []byte
		mainHeight, spineHeight uint32
		megablockTimestamp      int64
		sortedFileChunksHashes  [][]byte
		lastFileChunkHash       []byte
		ct                      chaintype.ChainType
		mbType                  model.MegablockType
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    *model.Megablock
	}{
		{
			name: "CreateMegablock:success",
			fields: fields{
				QueryExecutor: &mockMegablockServiceQueryExecutor{
					testName: "CreateMegablock:success",
				},
				MegablockQuery:  query.NewMegablockQuery(),
				SpineBlockQuery: query.NewBlockQuery(&chaintype.SpineChain{}),
				Logger:          log.New(),
			},
			args: args{
				snapshotHash:           make([]byte, 32),
				mainHeight:             ssMockMainBlock.Height,
				megablockTimestamp:     ssMockMainBlock.Timestamp,
				sortedFileChunksHashes: make([][]byte, 0),
				lastFileChunkHash:      make([]byte, 32),
				ct:                     &chaintype.MainChain{},
				mbType:                 model.MegablockType_Snapshot,
			},
			wantErr: false,
			want: &model.Megablock{
				ID:                  int64(5585293634049981880),
				FullFileHash:        make([]byte, 32),
				MegablockHeight:     ssMockMainBlock.Height,
				ExpirationTimestamp: int64(1562117306),
				FileChunkHashes:     make([]byte, 0),
				MegablockType:       model.MegablockType_Snapshot,
				ChainType:           0,
			},
		},
	}
	for _, tt := range tests {
		fmt.Println(t.Name())
		t.Run(tt.name, func(t *testing.T) {
			mbl := &MegablockService{
				QueryExecutor:   tt.fields.QueryExecutor,
				MegablockQuery:  tt.fields.MegablockQuery,
				SpineBlockQuery: tt.fields.SpineBlockQuery,
				Logger:          tt.fields.Logger,
			}
			got, err := mbl.CreateMegablock(tt.args.snapshotHash, tt.args.mainHeight, tt.args.megablockTimestamp,
				tt.args.sortedFileChunksHashes, tt.args.ct, tt.args.mbType)
			if (err != nil) != tt.wantErr {
				t.Errorf("SnapshotService.CreateMegablock() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SnapshotService.CreateMegablock() error = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSnapshotService_GetMegablockBytes(t *testing.T) {
	type fields struct {
		QueryExecutor   query.ExecutorInterface
		MegablockQuery  query.MegablockQueryInterface
		SpineBlockQuery query.BlockQueryInterface
		Logger          *log.Logger
	}
	type args struct {
		megablock *model.Megablock
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		{
			name:   "GetMegablockBytes:success",
			fields: fields{},
			args: args{
				megablock: ssMockMegablock,
			},
			want: []byte{1, 0, 0, 0, 0, 0, 0, 0, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
				3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
				3, 3, 208, 2, 0, 0, 0, 0, 0, 0, 232, 3, 0, 0, 0, 0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &MegablockService{
				QueryExecutor:   tt.fields.QueryExecutor,
				MegablockQuery:  tt.fields.MegablockQuery,
				SpineBlockQuery: tt.fields.SpineBlockQuery,
				Logger:          tt.fields.Logger,
			}
			got := ss.GetMegablockBytes(tt.args.megablock)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SnapshotService.GetMegablockBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
