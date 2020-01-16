package service

import (
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"reflect"
	"regexp"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockQueryExecutor struct {
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
	ssSnapshotInterval          = int64(1440 * 60 * 30) // 30 days
	ssSnapshotGenerationTimeout = int64(1440 * 60 * 3)  // 3 days
)

func (mqe *mockQueryExecutor) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	switch mqe.testName {
	case "CreateMegablock:success":
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
		default:
			return nil, fmt.Errorf("unmocked query for ExecuteSelectRow in test %s: %s", mqe.testName, qStr)
		}
	default:
		return nil, fmt.Errorf("test case not implemented %s: %s", mqe.testName, qStr)
	}

	row := db.QueryRow(qStr)
	return row, nil
}

func (*mockQueryExecutor) ExecuteTransaction(query string, args ...interface{}) error {
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
		MegablockQuery            query.MegablockQueryInterface
		Logger                    *log.Logger
		Spinechain                chaintype.ChainType
		Mainchain                 chaintype.ChainType
		SnapshotInterval          int64
		SnapshotGenerationTimeout int64
	}
	type args struct {
		mainHeight uint32
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
			},
			want: uint32(370285),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mbl := &BlockSpineSnapshotService{
				QueryExecutor:             tt.fields.QueryExecutor,
				MegablockQuery:            tt.fields.MegablockQuery,
				Logger:                    tt.fields.Logger,
				Mainchain:                 tt.fields.Mainchain,
				Spinechain:                tt.fields.Spinechain,
				SnapshotInterval:          tt.fields.SnapshotInterval,
				SnapshotGenerationTimeout: tt.fields.SnapshotGenerationTimeout,
			}
			if got := mbl.GetNextSnapshotHeight(tt.args.mainHeight); got != tt.want {
				t.Errorf("BlockSpineSnapshotService.GetNextSnapshotHeight() = %v, want %v", got, tt.want)
			}
		})
	}
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
		snapshotHash []byte
		mainHeight   uint32
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
				QueryExecutor: &mockQueryExecutor{
					testName: "CreateMegablock:success",
				},
				SpineBlockQuery:           query.NewBlockQuery(ssSpinechain),
				MainBlockQuery:            query.NewBlockQuery(ssMainchain),
				MegablockQuery:            query.NewMegablockQuery(),
				Logger:                    log.New(),
				Spinechain:                &mockSpinechain{},
				Mainchain:                 &mockMainchain{},
				SnapshotInterval:          ssSnapshotInterval,
				SnapshotGenerationTimeout: ssSnapshotGenerationTimeout,
			},
			args: args{
				snapshotHash: make([]byte, 64),
				mainHeight:   ssMockMainBlock.Height,
			},
			wantErr: false,
			want: &model.Megablock{
				FullSnapshotHash: make([]byte, 64),
				MainBlockHeight:  ssMockMainBlock.Height,
				SpineBlockHeight: uint32(419),
			},
		},
	}
	for _, tt := range tests {
		fmt.Println(t.Name())
		t.Run(tt.name, func(t *testing.T) {
			mbl := &BlockSpineSnapshotService{
				QueryExecutor:             tt.fields.QueryExecutor,
				MegablockQuery:            tt.fields.MegablockQuery,
				SpineBlockQuery:           tt.fields.SpineBlockQuery,
				MainBlockQuery:            tt.fields.MainBlockQuery,
				Logger:                    tt.fields.Logger,
				Spinechain:                tt.fields.Spinechain,
				Mainchain:                 tt.fields.Mainchain,
				SnapshotInterval:          tt.fields.SnapshotInterval,
				SnapshotGenerationTimeout: tt.fields.SnapshotGenerationTimeout,
			}
			got, err := mbl.CreateMegablock(tt.args.snapshotHash, tt.args.mainHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineSnapshotService.CreateMegablock() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockSpineSnapshotService.CreateMegablock() error = %v, want %v", got, tt.want)
			}
		})
	}
}
