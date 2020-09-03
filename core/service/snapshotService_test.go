package service

import (
	"database/sql"
	"fmt"
	"reflect"
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

	mockSnapshotMainBlockService struct {
		SnapshotMainBlockService
	}
)

func (*mockSnapshotMainBlockService) NewSnapshotFile(block *model.Block) (*model.SnapshotFileInfo, error) {
	return new(model.SnapshotFileInfo), nil
}

var (
	ssSpinechain    = &chaintype.SpineChain{}
	ssMainchain     = &chaintype.MainChain{}
	ssMockMainBlock = &model.Block{
		Height:    720,
		Timestamp: constant.MainchainGenesisBlockTimestamp + ssMainchain.GetSmithingPeriod(),
	}
	ssMockSpineBlock = &model.Block{
		Height:    10,
		Timestamp: constant.SpinechainGenesisBlockTimestamp + ssSpinechain.GetSmithingPeriod(),
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
		case "SELECT MAX(height), id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
			"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, " +
			"version FROM main_block":
			mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(query.NewBlockQuery(ssMainchain).Fields).
				AddRow(
					ssMockMainBlock.Height,
					ssMockMainBlock.ID,
					ssMockMainBlock.BlockHash,
					ssMockMainBlock.PreviousBlockHash,
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
		case "SELECT MAX(height), id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
			"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, " +
			"version FROM spine_block":
			mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(query.NewBlockQuery(ssMainchain).Fields).
				AddRow(
					ssMockSpineBlock.Height,
					ssMockSpineBlock.ID,
					ssMockSpineBlock.BlockHash,
					ssMockSpineBlock.PreviousBlockHash,
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

func TestSnapshotService_GenerateSnapshot(t *testing.T) {
	type fields struct {
		SpineBlockManifestService SpineBlockManifestServiceInterface
		BlockchainStatusService   BlockchainStatusServiceInterface
		SnapshotBlockServices     map[int32]SnapshotBlockServiceInterface
		Logger                    *log.Logger
	}
	type args struct {
		block                    *model.Block
		ct                       chaintype.ChainType
		snapshotChunkBytesLength int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.SnapshotFileInfo
		wantErr bool
	}{
		{
			name: "GenerateSnapshot",
			args: args{
				ct:    ssMainchain,
				block: ssMockMainBlock,
			},
			fields: fields{
				SnapshotBlockServices: map[int32]SnapshotBlockServiceInterface{
					0: &mockSnapshotMainBlockService{},
				},
				Logger: log.New(),
			},
			want: new(model.SnapshotFileInfo),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SnapshotService{
				SpineBlockManifestService: tt.fields.SpineBlockManifestService,
				BlockchainStatusService:   tt.fields.BlockchainStatusService,
				SnapshotBlockServices:     tt.fields.SnapshotBlockServices,
				Logger:                    tt.fields.Logger,
			}
			got, err := ss.GenerateSnapshot(tt.args.block, tt.args.ct, tt.args.snapshotChunkBytesLength)
			if (err != nil) != tt.wantErr {
				t.Errorf("SnapshotService.GenerateSnapshot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SnapshotService.GenerateSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}
