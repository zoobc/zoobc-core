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

func (*mockMegablockServiceQueryExecutor) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}

func (*mockMegablockServiceQueryExecutor) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func TestBlockSpineSnapshotService_CreateMegablock(t *testing.T) {
	type fields struct {
		QueryExecutor             query.ExecutorInterface
		MegablockQuery            query.MegablockQueryInterface
		SpineBlockQuery           query.BlockQueryInterface
		MainBlockQuery            query.BlockQueryInterface
		FileChunkQuery            query.FileChunkQueryInterface
		Logger                    *log.Logger
		Spinechain                chaintype.ChainType
		Mainchain                 chaintype.ChainType
		SnapshotInterval          int64
		SnapshotGenerationTimeout int64
	}
	type args struct {
		snapshotHash            []byte
		mainHeight, spineHeight uint32
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
				MegablockQuery: query.NewMegablockQuery(),
				FileChunkQuery: query.NewFileChunkQuery(),
				Logger:         log.New(),
			},
			args: args{
				snapshotHash:           make([]byte, 64),
				mainHeight:             ssMockMainBlock.Height,
				spineHeight:            ssMockSpineBlock.Height,
				sortedFileChunksHashes: make([][]byte, 0),
				lastFileChunkHash:      make([]byte, 64),
				ct:                     &chaintype.MainChain{},
				mbType:                 model.MegablockType_Snapshot,
			},
			wantErr: false,
			want: &model.Megablock{
				ID:                     0,
				FullFileHash:           make([]byte, 64),
				MegablockPayloadLength: 0,
				MegablockPayloadHash: []byte{166, 159, 115, 204,
					162, 58, 154, 197, 200, 181, 103, 220, 24, 90, 117, 110, 151, 201, 130, 22, 79,
					226, 88, 89, 224, 209, 220, 193, 71, 92, 128, 166, 21, 178, 18, 58, 241, 245,
					249, 76, 17, 227, 233, 64, 44, 58, 197, 88, 245, 0, 25, 157, 149, 182, 211, 227,
					1, 117, 133, 134, 40, 29, 205, 38},
				MegablockHeight:  ssMockMainBlock.Height,
				SpineBlockHeight: ssMockSpineBlock.Height,
				FileChunks:       make([]*model.FileChunk, 0),
			},
		},
	}
	for _, tt := range tests {
		fmt.Println(t.Name())
		t.Run(tt.name, func(t *testing.T) {
			mbl := &MegablockService{
				QueryExecutor:  tt.fields.QueryExecutor,
				MegablockQuery: tt.fields.MegablockQuery,
				FileChunkQuery: tt.fields.FileChunkQuery,
				Logger:         tt.fields.Logger,
			}
			got, err := mbl.CreateMegablock(tt.args.snapshotHash, tt.args.mainHeight, tt.args.spineHeight,
				tt.args.sortedFileChunksHashes,
				tt.args.lastFileChunkHash, tt.args.ct, tt.args.mbType)
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
		QueryExecutor  query.ExecutorInterface
		MegablockQuery query.MegablockQueryInterface
		FileChunkQuery query.FileChunkQueryInterface
		Logger         *log.Logger
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
			want: []byte{0, 0, 0, 0, 0, 0, 0, 0, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
				3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
				3, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10,
				0, 0, 0, 208, 2, 0, 0, 0, 0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &MegablockService{
				QueryExecutor:  tt.fields.QueryExecutor,
				MegablockQuery: tt.fields.MegablockQuery,
				FileChunkQuery: tt.fields.FileChunkQuery,
				Logger:         tt.fields.Logger,
			}
			got := ss.GetMegablockBytes(tt.args.megablock)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SnapshotService.GetMegablockBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
