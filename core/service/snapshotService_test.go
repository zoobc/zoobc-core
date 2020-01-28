package service

import (
	"database/sql"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockSnapshotServiceQueryExecutor struct {
		testName string
		query.Executor
	}
	mockSpinechain struct {
		chaintype.SpineChain
	}
	mockMainchain struct {
		chaintype.SpineChain
	}
)

var (
	ssSpinechain    = &chaintype.SpineChain{}
	ssMainchain     = &chaintype.MainChain{}
	ssMockMainBlock = &model.Block{
		Height: 720,
		Timestamp: constant.MainchainGenesisBlockTimestamp + ssMainchain.GetSmithingPeriod() + ssMainchain.
			GetChainSmithingDelayTime(),
	}
	ssMockSpineBlock = &model.Block{
		Height: 10,
		Timestamp: constant.SpinechainGenesisBlockTimestamp + ssSpinechain.GetSmithingPeriod() + ssSpinechain.
			GetChainSmithingDelayTime(),
	}
	// ssSnapshotInterval          = uint32(1440 * 60 * 30) // 30 days
	// ssSnapshotGenerationTimeout = int64(1440 * 60 * 3)   // 3 days
	ssMockFullHash = []byte{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
		3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3}
)

func (mqe *mockSnapshotServiceQueryExecutor) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	switch mqe.testName {
	case "GenerateSnapshot:success":
		switch qStr {
		case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, " +
			"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, " +
			"version FROM main_block ORDER BY height DESC LIMIT 1":
			mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(query.NewBlockQuery(ssMainchain).Fields).
				AddRow(
					ssMockMainBlock.ID,
					ssMockMainBlock.BlockHash,
					ssMockMainBlock.PreviousBlockHash,
					ssMockMainBlock.Height,
					ssMockMainBlock.Timestamp,
					ssMockMainBlock.BlockSeed,
					ssMockMainBlock.BlockSignature,
					ssMockMainBlock.CumulativeDifficulty,
					ssMockMainBlock.PayloadLength,
					ssMockMainBlock.PayloadHash,
					ssMockMainBlock.BlocksmithPublicKey,
					ssMockMainBlock.TotalAmount,
					ssMockMainBlock.TotalFee,
					ssMockMainBlock.TotalCoinBase,
					ssMockMainBlock.Version,
				))
		case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, " +
			"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, " +
			"version FROM spine_block ORDER BY height DESC LIMIT 1":
			mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(query.NewBlockQuery(ssMainchain).Fields).
				AddRow(
					ssMockSpineBlock.ID,
					ssMockSpineBlock.BlockHash,
					ssMockSpineBlock.PreviousBlockHash,
					ssMockSpineBlock.Height,
					ssMockSpineBlock.Timestamp,
					ssMockSpineBlock.BlockSeed,
					ssMockSpineBlock.BlockSignature,
					ssMockSpineBlock.CumulativeDifficulty,
					ssMockSpineBlock.PayloadLength,
					ssMockSpineBlock.PayloadHash,
					ssMockSpineBlock.BlocksmithPublicKey,
					ssMockSpineBlock.TotalAmount,
					ssMockSpineBlock.TotalFee,
					ssMockSpineBlock.TotalCoinBase,
					ssMockSpineBlock.Version,
				))
		default:
			return nil, fmt.Errorf("unmocked query for ExecuteSelectRow in test %s: %s", mqe.testName, qStr)
		}
	default:
		return nil, fmt.Errorf("test case not implemented %s: %s", mqe.testName, qStr)
	}

	row := db.QueryRow(qStr)
	return row, nil
}

func (*mockSnapshotServiceQueryExecutor) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}

func (*mockSnapshotServiceQueryExecutor) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func (*mockSpinechain) GetChainSmithingDelayTime() int64 {
	return 20
}

func (*mockSpinechain) GetSmithingPeriod() int64 {
	return 600
}

func (*mockMainchain) GetChainSmithingDelayTime() int64 {
	return 20
}

func (*mockMainchain) GetSmithingPeriod() int64 {
	return 15
}

// FIXME: uncomment and fix the test once this method is completed
// func TestSnapshotService_GenerateSnapshot(t *testing.T) {
// 	type fields struct {
// 		QueryExecutor             query.ExecutorInterface
// 		SpineBlockManifestQuery            query.SpineBlockManifestQueryInterface
// 		SpineBlockQuery           query.BlockQueryInterface
// 		MainBlockQuery            query.BlockQueryInterface
// 		FileChunkQuery        query.FileChunkQueryInterface
// 		Logger                    *log.Logger
// 		Spinechain                chaintype.ChainType
// 		Mainchain                 chaintype.ChainType
// 		MainchainSnapshotInterval          int64
// 		SnapshotGenerationTimeout int64
// 	}
// 	type args struct {
// 		mainHeight uint32
// 		ct         chaintype.ChainType
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    *model.SpineBlockManifest
// 		wantErr bool
// 	}{
// 		{
// 			name: "GenerateSnapshot:success",
// 			fields: fields{
// 				QueryExecutor: &mockSnapshotServiceQueryExecutor{
// 					testName: "GenerateSnapshot:success",
// 				},
// 				SpineBlockQuery:           query.NewBlockQuery(ssSpinechain),
// 				MainBlockQuery:            query.NewBlockQuery(ssMainchain),
// 				SpineBlockManifestQuery:            query.NewSpineBlockManifestQuery(),
// 				FileChunkQuery:        query.NewFileChunkQuery(),
// 				Logger:                    log.New(),
// 				Spinechain:                &mockSpinechain{},
// 				Mainchain:                 &mockMainchain{},
// 				MainchainSnapshotInterval:          ssSnapshotInterval,
// 				SnapshotGenerationTimeout: ssSnapshotGenerationTimeout,
// 			},
// 			args: args{
// 				mainHeight: ssMockMainBlock.Height,
// 				ct:         &chaintype.MainChain{},
// 			},
// 			wantErr: false,
// 			want: &model.SpineBlockManifest{
// 				ID: int64(1919891213155270003),
// 				FullFileHash:     make([]byte, 64),
// 				SpineBlockManifestHeight:  ssMockMainBlock.Height,
// 				SpineBlockHeight: uint32(419),
// 				FileChunks:   make([]*model.FileChunk, 0),
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ss := &SnapshotService{
// 				QueryExecutor:             tt.fields.QueryExecutor,
// 				SpineBlockManifestQuery:            tt.fields.SpineBlockManifestQuery,
// 				SpineBlockQuery:           tt.fields.SpineBlockQuery,
// 				MainBlockQuery:            tt.fields.MainBlockQuery,
// 				FileChunkQuery:        tt.fields.FileChunkQuery,
// 				Logger:                    tt.fields.Logger,
// 				Spinechain:                tt.fields.Spinechain,
// 				Mainchain:                 tt.fields.Mainchain,
// 				MainchainSnapshotInterval:          tt.fields.MainchainSnapshotInterval,
// 				SnapshotGenerationTimeout: tt.fields.SnapshotGenerationTimeout,
// 			}
// 			got, err := ss.GenerateSnapshot(tt.args.mainHeight, tt.args.ct)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("SnapshotService.GenerateSnapshot() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("SnapshotService.GenerateSnapshot() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestSnapshotService_IsSnapshotHeight(t *testing.T) {
	type fields struct {
		QueryExecutor             query.ExecutorInterface
		SpineBlockQuery           query.BlockQueryInterface
		MainBlockQuery            query.BlockQueryInterface
		Logger                    *log.Logger
		Spinechain                chaintype.ChainType
		Mainchain                 chaintype.ChainType
		SnapshotInterval          uint32
		SnapshotGenerationTimeout int64
		SpineBlockManifestService SpineBlockManifestServiceInterface
	}
	type args struct {
		height           uint32
		snapshotInterval uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_1}:",
			args: args{
				height:           1,
				snapshotInterval: 10,
			},
			want: false,
		},
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_2}:",
			args: args{
				height:           constant.MinRollbackBlocks,
				snapshotInterval: 10,
			},
			want: true,
		},
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_3}:",
			args: args{
				height:           constant.MinRollbackBlocks + 9,
				snapshotInterval: 10,
			},
			want: false,
		},
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_4}:",
			args: args{
				height:           constant.MinRollbackBlocks + 10,
				snapshotInterval: 10,
			},
			want: true,
		},
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_5}:",
			args: args{
				height:           constant.MinRollbackBlocks + 20,
				snapshotInterval: 10,
			},
			want: true,
		},
		{
			name: "IsSnapshotHeight_{interval_higher_than_minRollback_1}:",
			args: args{
				height:           10,
				snapshotInterval: constant.MinRollbackBlocks + 10,
			},
			want: false,
		},
		{
			name: "IsSnapshotHeight_{interval_higher_than_minRollback_2}:",
			args: args{
				height:           constant.MinRollbackBlocks,
				snapshotInterval: constant.MinRollbackBlocks + 10,
			},
			want: false,
		},
		{
			name: "IsSnapshotHeight_{interval_higher_than_minRollback_3}:",
			args: args{
				height:           constant.MinRollbackBlocks + 10,
				snapshotInterval: constant.MinRollbackBlocks + 10,
			},
			want: true,
		},
		{
			name: "IsSnapshotHeight_{interval_higher_than_minRollback_4}:",
			args: args{
				height:           2 * (constant.MinRollbackBlocks + 10),
				snapshotInterval: constant.MinRollbackBlocks + 10,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SnapshotService{
				QueryExecutor:             tt.fields.QueryExecutor,
				SpineBlockQuery:           tt.fields.SpineBlockQuery,
				MainBlockQuery:            tt.fields.MainBlockQuery,
				Logger:                    tt.fields.Logger,
				Spinechain:                tt.fields.Spinechain,
				Mainchain:                 tt.fields.Mainchain,
				SnapshotInterval:          tt.fields.SnapshotInterval,
				SnapshotGenerationTimeout: tt.fields.SnapshotGenerationTimeout,
				SpineBlockManifestService: tt.fields.SpineBlockManifestService,
			}
			if got := s.IsSnapshotHeight(tt.args.height, tt.args.snapshotInterval); got != tt.want {
				t.Errorf("SnapshotService.IsSnapshotHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}
