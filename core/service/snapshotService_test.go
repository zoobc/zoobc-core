package service

import (
	"database/sql"
	"errors"
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

type (
	mockFileDownloaderService struct {
		FileDownloaderService
		success bool
	}
)

func (mfdf *mockFileDownloaderService) DownloadFileByName(fileName string, fileHash []byte) error {
	if mfdf.success {
		return nil
	}
	return errors.New("DownloadFileByNameFail")
}

func TestSnapshotService_DownloadSnapshot(t *testing.T) {
	type fields struct {
		SpineBlockManifestService SpineBlockManifestServiceInterface
		SpineBlockDownloadService SpineBlockDownloadServiceInterface
		SnapshotBlockServices     map[int32]SnapshotBlockServiceInterface
		FileDownloaderService     FileDownloaderServiceInterface
		FileService               FileServiceInterface
		Logger                    *log.Logger
	}
	type args struct {
		spineBlockManifest *model.SpineBlockManifest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		errMsg  string
	}{
		{
			name: "DownloadSnapshot:fail-{zerolength}",
			args: args{
				spineBlockManifest: &model.SpineBlockManifest{
					FileChunkHashes: make([]byte, 0),
				},
			},
			wantErr: true,
			errMsg:  "ValidationErr: invalid file chunks hashes length",
		},
		{
			name: "DownloadSnapshot:fail-{DownloadFailed}",
			fields: fields{
				FileDownloaderService: &mockFileDownloaderService{
					success: false,
				},
				FileService: &mockFileService{
					successGetFileNameFromHash: true,
				},
				Logger: log.New(),
			},
			args: args{
				spineBlockManifest: &model.SpineBlockManifest{
					FileChunkHashes: make([]byte, 64),
				},
			},
			wantErr: true,
			errMsg: "AppErr: One or more snapshot chunks failed to download [vXu9Q01j1OWLRoqmIHW-KpyJBticdBS207Lg3OscPgyO" +
				" vXu9Q01j1OWLRoqmIHW-KpyJBticdBS207Lg3OscPgyO]",
		},
		{
			name: "DownloadSnapshot:success",
			fields: fields{
				FileDownloaderService: &mockFileDownloaderService{
					success: true,
				},
				FileService: &mockFileService{
					successGetFileNameFromHash: true,
				},
				Logger: log.New(),
			},
			args: args{
				spineBlockManifest: &model.SpineBlockManifest{
					FileChunkHashes: make([]byte, 64),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SnapshotService{
				SpineBlockManifestService: tt.fields.SpineBlockManifestService,
				SpineBlockDownloadService: tt.fields.SpineBlockDownloadService,
				SnapshotBlockServices:     tt.fields.SnapshotBlockServices,
				FileDownloaderService:     tt.fields.FileDownloaderService,
				FileService:               tt.fields.FileService,
				Logger:                    tt.fields.Logger,
			}
			if err := ss.DownloadSnapshot(tt.args.spineBlockManifest); err != nil {
				if !tt.wantErr {
					t.Errorf("SnapshotService.DownloadSnapshot() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.errMsg != err.Error() {
					t.Errorf("SnapshotService.DownloadSnapshot() error wrong test exit point: %v", err)
				}
			}
		})
	}
}
