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
	ssSnapshotInterval          = int64(1440 * 60 * 30) // 30 days
	ssSnapshotGenerationTimeout = int64(1440 * 60 * 3)  // 3 days
	ssMockHash1                 = []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	ssMockHash2 = []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
		2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	ssMockFullHash = []byte{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
		3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3}
	ssMockFileChunk = &model.FileChunk{
		MegablockID: 1,
		ChunkHash:   ssMockHash1,
		ChainType:   ssMainchain.GetTypeInt(),
	}
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
		case "SELECT chunk_hash, megablock_id, chunk_index, previous_chunk_hash, spine_block_height, " +
			"chain_type FROM file_chunk WHERE chain_type = 0 ORDER BY spine_block_height, chunk_index DESC LIMIT 1":
			mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(query.NewFileChunkQuery().Fields).
				AddRow(
					ssMockFileChunk.ChunkHash,
					ssMockFileChunk.MegablockID,
					ssMockFileChunk.ChunkIndex,
					ssMockFileChunk.PreviousChunkHash,
					ssMockFileChunk.SpineBlockHeight,
					ssMockFileChunk.ChainType,
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

func TestBlockSpineSnapshotService_GetNextSnapshotHeight(t *testing.T) {
	type fields struct {
		QueryExecutor             query.ExecutorInterface
		FileChunkQuery            query.FileChunkQueryInterface
		Logger                    *log.Logger
		Spinechain                chaintype.ChainType
		Mainchain                 chaintype.ChainType
		SnapshotInterval          int64
		SnapshotGenerationTimeout int64
	}
	type args struct {
		mainHeight uint32
		ct         chaintype.ChainType
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   uint32
	}{
		{
			name: "GetNextSnapshotHeight:success-{height_less_than_minRollback_height}",
			fields: fields{
				Spinechain:                &mockSpinechain{},
				Mainchain:                 &mockMainchain{},
				SnapshotInterval:          ssSnapshotInterval,
				SnapshotGenerationTimeout: ssSnapshotGenerationTimeout,
			},
			args: args{
				mainHeight: 100,
				ct:         &chaintype.MainChain{},
			},
			want: uint32(74057),
		},
		{
			name: "GetNextSnapshotHeight:success-{height_lower_than_nextStep}",
			fields: fields{
				Spinechain:                &mockSpinechain{},
				Mainchain:                 &mockMainchain{},
				SnapshotInterval:          ssSnapshotInterval,
				SnapshotGenerationTimeout: ssSnapshotGenerationTimeout,
			},
			args: args{
				mainHeight: 1000,
				ct:         &chaintype.MainChain{},
			},
			want: uint32(74057),
		},
		{
			name: "GetNextSnapshotHeight:success-{height_same_as_nextStep}",
			fields: fields{
				Spinechain:                &mockSpinechain{},
				Mainchain:                 &mockMainchain{},
				SnapshotInterval:          ssSnapshotInterval,
				SnapshotGenerationTimeout: ssSnapshotGenerationTimeout,
			},
			args: args{
				mainHeight: 148114,
				ct:         &chaintype.MainChain{},
			},
			want: uint32(148114),
		},
		{
			name: "GetNextSnapshotHeight:success-{height_higher_than_nextStep}",
			fields: fields{
				Spinechain:                &mockSpinechain{},
				Mainchain:                 &mockMainchain{},
				SnapshotInterval:          ssSnapshotInterval,
				SnapshotGenerationTimeout: ssSnapshotGenerationTimeout,
			},
			args: args{
				mainHeight: 148115,
				ct:         &chaintype.MainChain{},
			},
			want: uint32(222171),
		},
		{
			name: "GetNextSnapshotHeight:success-{height_more_than_double_nextStep}",
			fields: fields{
				Spinechain:                &mockSpinechain{},
				Mainchain:                 &mockMainchain{},
				SnapshotInterval:          ssSnapshotInterval,
				SnapshotGenerationTimeout: ssSnapshotGenerationTimeout,
			},
			args: args{
				mainHeight: 296230,
				ct:         &chaintype.MainChain{},
			},
			want: uint32(370285),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mbl := &SnapshotService{
				QueryExecutor:             tt.fields.QueryExecutor,
				FileChunkQuery:            tt.fields.FileChunkQuery,
				Logger:                    tt.fields.Logger,
				Mainchain:                 tt.fields.Mainchain,
				Spinechain:                tt.fields.Spinechain,
				SnapshotInterval:          tt.fields.SnapshotInterval,
				SnapshotGenerationTimeout: tt.fields.SnapshotGenerationTimeout,
			}
			if got := mbl.GetNextSnapshotHeight(tt.args.mainHeight, tt.args.ct); got != tt.want {
				t.Errorf("SnapshotService.GetNextSnapshotHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

// FIXME: uncomment and fix the test once this method is completed
// func TestSnapshotService_GenerateSnapshot(t *testing.T) {
// 	type fields struct {
// 		QueryExecutor             query.ExecutorInterface
// 		MegablockQuery            query.MegablockQueryInterface
// 		SpineBlockQuery           query.BlockQueryInterface
// 		MainBlockQuery            query.BlockQueryInterface
// 		FileChunkQuery        query.FileChunkQueryInterface
// 		Logger                    *log.Logger
// 		Spinechain                chaintype.ChainType
// 		Mainchain                 chaintype.ChainType
// 		SnapshotInterval          int64
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
// 		want    *model.Megablock
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
// 				MegablockQuery:            query.NewMegablockQuery(),
// 				FileChunkQuery:        query.NewFileChunkQuery(),
// 				Logger:                    log.New(),
// 				Spinechain:                &mockSpinechain{},
// 				Mainchain:                 &mockMainchain{},
// 				SnapshotInterval:          ssSnapshotInterval,
// 				SnapshotGenerationTimeout: ssSnapshotGenerationTimeout,
// 			},
// 			args: args{
// 				mainHeight: ssMockMainBlock.Height,
// 				ct:         &chaintype.MainChain{},
// 			},
// 			wantErr: false,
// 			want: &model.Megablock{
// 				ID: int64(1919891213155270003),
// 				FullFileHash:     make([]byte, 64),
// 				MegablockHeight:  ssMockMainBlock.Height,
// 				SpineBlockHeight: uint32(419),
// 				FileChunks:   make([]*model.FileChunk, 0),
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ss := &SnapshotService{
// 				QueryExecutor:             tt.fields.QueryExecutor,
// 				MegablockQuery:            tt.fields.MegablockQuery,
// 				SpineBlockQuery:           tt.fields.SpineBlockQuery,
// 				MainBlockQuery:            tt.fields.MainBlockQuery,
// 				FileChunkQuery:        tt.fields.FileChunkQuery,
// 				Logger:                    tt.fields.Logger,
// 				Spinechain:                tt.fields.Spinechain,
// 				Mainchain:                 tt.fields.Mainchain,
// 				SnapshotInterval:          tt.fields.SnapshotInterval,
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
