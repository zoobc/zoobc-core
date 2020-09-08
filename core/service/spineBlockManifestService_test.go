package service

import (
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockSpineBlockManifestServiceQueryExecutor struct {
		testName string
		query.Executor
	}
)

var (
	ssMockSpineBlockManifest = &model.SpineBlockManifest{
		ID:                      1,
		FullFileHash:            ssMockFullHash,
		ManifestReferenceHeight: 720,
		FileChunkHashes:         []byte{},
		ChainType:               0,
		SpineBlockManifestType:  model.SpineBlockManifestType_Snapshot,
		ExpirationTimestamp:     1000,
	}
)

func (*mockSpineBlockManifestServiceQueryExecutor) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}

func (*mockSpineBlockManifestServiceQueryExecutor) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func (*mockSpineBlockManifestServiceQueryExecutor) BeginTx() error {
	return nil
}

func (*mockSpineBlockManifestServiceQueryExecutor) RollbackTx() error {
	return nil
}
func (*mockSpineBlockManifestServiceQueryExecutor) CommitTx() error {
	return nil
}

func TestBlockSpineSnapshotService_CreateSpineBlockManifest(t *testing.T) {
	type fields struct {
		QueryExecutor             query.ExecutorInterface
		SpineBlockManifestQuery   query.SpineBlockManifestQueryInterface
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
		mbType                  model.SpineBlockManifestType
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    *model.SpineBlockManifest
	}{
		{
			name: "CreateSpineBlockManifest:success",
			fields: fields{
				QueryExecutor: &mockSpineBlockManifestServiceQueryExecutor{
					testName: "CreateSpineBlockManifest:success",
				},
				SpineBlockManifestQuery: query.NewSpineBlockManifestQuery(),
				SpineBlockQuery:         query.NewBlockQuery(&chaintype.SpineChain{}),
				Logger:                  log.New(),
			},
			args: args{
				snapshotHash:           make([]byte, 32),
				mainHeight:             ssMockMainBlock.Height,
				megablockTimestamp:     ssMockMainBlock.Timestamp,
				sortedFileChunksHashes: make([][]byte, 0),
				lastFileChunkHash:      make([]byte, 32),
				ct:                     &chaintype.MainChain{},
				mbType:                 model.SpineBlockManifestType_Snapshot,
			},
			wantErr: false,
			want: &model.SpineBlockManifest{
				ID:                      int64(-6343438915916024408),
				FullFileHash:            make([]byte, 32),
				ManifestReferenceHeight: ssMockMainBlock.Height,
				ExpirationTimestamp:     int64(1596708015),
				FileChunkHashes:         make([]byte, 0),
				SpineBlockManifestType:  model.SpineBlockManifestType_Snapshot,
				ChainType:               int32(0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mbl := &SpineBlockManifestService{
				QueryExecutor:           tt.fields.QueryExecutor,
				SpineBlockManifestQuery: tt.fields.SpineBlockManifestQuery,
				SpineBlockQuery:         tt.fields.SpineBlockQuery,
				Logger:                  tt.fields.Logger,
			}
			got, err := mbl.CreateSpineBlockManifest(tt.args.snapshotHash, tt.args.mainHeight, tt.args.megablockTimestamp,
				tt.args.sortedFileChunksHashes, tt.args.ct, tt.args.mbType)
			if (err != nil) != tt.wantErr {
				t.Errorf("SnapshotService.CreateSpineBlockManifest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SnapshotService.CreateSpineBlockManifest() error = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestSnapshotService_GetSpineBlockManifestBytes(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		SpineBlockManifestQuery query.SpineBlockManifestQueryInterface
		SpineBlockQuery         query.BlockQueryInterface
		Logger                  *log.Logger
	}
	type args struct {
		spineBlockManifest *model.SpineBlockManifest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		{
			name:   "GetSpineBlockManifestBytes:success",
			fields: fields{},
			args: args{
				spineBlockManifest: ssMockSpineBlockManifest,
			},
			want: []byte{1, 0, 0, 0, 0, 0, 0, 0, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
				3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
				3, 3, 3, 3, 3, 3, 3, 208, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 232, 3, 0, 0, 0, 0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SpineBlockManifestService{
				QueryExecutor:           tt.fields.QueryExecutor,
				SpineBlockManifestQuery: tt.fields.SpineBlockManifestQuery,
				SpineBlockQuery:         tt.fields.SpineBlockQuery,
				Logger:                  tt.fields.Logger,
			}
			got := ss.GetSpineBlockManifestBytes(tt.args.spineBlockManifest)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SnapshotService.GetSpineBlockManifestBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
