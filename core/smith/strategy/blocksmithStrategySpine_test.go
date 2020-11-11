package strategy

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockQueryGetBlocksmithsSpineSuccessNoBlocksmith struct {
		query.Executor
	}
	mockQueryGetBlocksmithsSpineSuccessWithBlocksmith struct {
		query.Executor
	}

	mockQuerySortBlocksmithSpineSuccessWithBlocksmiths struct {
		query.Executor
	}
	mockQueryGetBlocksmithsSpineFail struct {
		query.Executor
	}
)

func (*mockQueryGetBlocksmithsSpineFail) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	return nil, errors.New("mockError")
}

func (*mockQueryGetBlocksmithsSpineFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(query.NewBlockQuery(&chaintype.SpineChain{}).Fields).AddRow(
			mockBlock.GetID(),
			mockBlock.GetBlockHash(),
			mockBlock.GetPreviousBlockHash(),
			mockBlock.GetHeight(),
			mockBlock.GetTimestamp(),
			mockBlock.GetBlockSeed(),
			mockBlock.GetBlockSignature(),
			mockBlock.GetCumulativeDifficulty(),
			mockBlock.GetPayloadLength(),
			mockBlock.GetPayloadHash(),
			mockBlock.GetBlocksmithPublicKey(),
			mockBlock.GetTotalAmount(),
			mockBlock.GetTotalFee(),
			mockBlock.GetTotalCoinBase(),
			mockBlock.GetVersion(),
			mockBlock.GetMerkleRoot(),
			mockBlock.GetMerkleTree(),
			mockBlock.GetReferenceBlockHeight(),
		))
	return db.QueryRow(qStr), nil
}

func (*mockQueryGetBlocksmithsSpineSuccessWithBlocksmith) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	var (
		rows *sql.Rows
	)
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}

	defer db.Close()
	switch qStr {
	case "SELECT node_public_key, node_id, public_key_action, main_block_height, latest, height " +
		"FROM spine_public_key WHERE height >= 0 AND height <= 1 AND " +
		"public_key_action=0 AND latest=1 ORDER BY height":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
			[]string{
				"node_public_key",
				"node_id",
				"public_key_action",
				"main_block_height",
				"latest",
				"height",
			},
		).AddRow(
			bssMockBlocksmiths[0].NodePublicKey,
			bssMockBlocksmiths[0].NodeID,
			uint32(model.SpinePublicKeyAction_AddKey),
			1,
			true,
			uint32(1),
		))
		rows, _ = db.Query(qStr)
	default:
		return nil, fmt.Errorf("mockQueryGetBlocksmithsSpineSuccessWithBlocksmith - unmocked query: %s", qStr)
	}
	return rows, nil
}

func (*mockQuerySortBlocksmithSpineSuccessWithBlocksmiths) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	var (
		rows *sql.Rows
	)
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qStr {
	case "SELECT node_public_key, node_id, public_key_action, latest, height FROM spine_public_key " +
		"WHERE height <= 1 AND public_key_action=0 AND latest=1 ORDER BY height":
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows(
			[]string{
				"node_public_key",
				"node_id",
				"public_key_action",
				"latest",
				"height",
			},
		).AddRow(
			bssMockBlocksmiths[0].NodePublicKey,
			uint32(model.SpinePublicKeyAction_AddKey),
			true,
			uint32(1),
		).AddRow(
			bssMockBlocksmiths[1].NodePublicKey,
			uint32(model.SpinePublicKeyAction_AddKey),
			true,
			uint32(1),
		))
		rows, _ = db.Query("A")
	default:
		return nil, errors.New("MockErr")
	}
	return rows, nil
}

func (*mockQuerySortBlocksmithSpineSuccessWithBlocksmiths) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(query.NewBlockQuery(&chaintype.SpineChain{}).Fields).AddRow(
			mockBlock.GetID(),
			mockBlock.GetBlockHash(),
			mockBlock.GetPreviousBlockHash(),
			mockBlock.GetHeight(),
			mockBlock.GetTimestamp(),
			mockBlock.GetBlockSeed(),
			mockBlock.GetBlockSignature(),
			mockBlock.GetCumulativeDifficulty(),
			mockBlock.GetPayloadLength(),
			mockBlock.GetPayloadHash(),
			mockBlock.GetBlocksmithPublicKey(),
			mockBlock.GetTotalAmount(),
			mockBlock.GetTotalFee(),
			mockBlock.GetTotalCoinBase(),
			mockBlock.GetVersion(),
			mockBlock.GetMerkleRoot(),
			mockBlock.GetMerkleTree(),
			mockBlock.GetReferenceBlockHeight(),
		))
	return db.QueryRow(qStr), nil
}

func (*mockQueryGetBlocksmithsSpineSuccessNoBlocksmith) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows(
		[]string{
			"node_public_key",
			"node_id",
			"public_key_action",
			"latest",
			"height",
		},
	))
	rows, _ := db.Query("A")
	return rows, nil
}

func (*mockQueryGetBlocksmithsSpineSuccessNoBlocksmith) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(query.NewBlockQuery(&chaintype.SpineChain{}).Fields).AddRow(
			mockBlock.GetID(),
			mockBlock.GetBlockHash(),
			mockBlock.GetPreviousBlockHash(),
			mockBlock.GetHeight(),
			mockBlock.GetTimestamp(),
			mockBlock.GetBlockSeed(),
			mockBlock.GetBlockSignature(),
			mockBlock.GetCumulativeDifficulty(),
			mockBlock.GetPayloadLength(),
			mockBlock.GetPayloadHash(),
			mockBlock.GetBlocksmithPublicKey(),
			mockBlock.GetTotalAmount(),
			mockBlock.GetTotalFee(),
			mockBlock.GetTotalCoinBase(),
			mockBlock.GetVersion(),
			mockBlock.GetMerkleRoot(),
			mockBlock.GetMerkleTree(),
			mockBlock.GetReferenceBlockHeight(),
		))
	return db.QueryRow(qStr), nil
}

func TestBlocksmithStrategySpine_SortBlocksmiths(t *testing.T) {
	type fields struct {
		QueryExecutor        query.ExecutorInterface
		SpinePublicKeyQuery  query.SpinePublicKeyQueryInterface
		Logger               *log.Logger
		SortedBlocksmiths    []*model.Blocksmith
		LastSortedBlockID    int64
		SortedBlocksmithsMap map[string]*int64
		SpineBlockQuery      query.BlockQueryInterface
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Success",
			fields: fields{
				QueryExecutor:        &mockQuerySortBlocksmithSpineSuccessWithBlocksmiths{},
				SpinePublicKeyQuery:  query.NewSpinePublicKeyQuery(),
				Logger:               log.New(),
				SortedBlocksmiths:    nil,
				SortedBlocksmithsMap: make(map[string]*int64),
				LastSortedBlockID:    1,
				SpineBlockQuery:      query.NewBlockQuery(&chaintype.SpineChain{}),
			},
			args: args{
				block: mockBlock,
			},
		},
		{
			name: "GetBlocksmiths:fail-{noSpinePublicKeyFound}",
			fields: fields{
				QueryExecutor:        &mockQueryGetBlocksmithsSpineFail{},
				SpinePublicKeyQuery:  query.NewSpinePublicKeyQuery(),
				Logger:               log.New(),
				SortedBlocksmiths:    nil,
				SortedBlocksmithsMap: make(map[string]*int64),
				LastSortedBlockID:    1,
				SpineBlockQuery:      query.NewBlockQuery(&chaintype.SpineChain{}),
			},
			args: args{mockBlock},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategySpine{
				QueryExecutor:        tt.fields.QueryExecutor,
				SpinePublicKeyQuery:  tt.fields.SpinePublicKeyQuery,
				Logger:               tt.fields.Logger,
				SortedBlocksmiths:    tt.fields.SortedBlocksmiths,
				LastSortedBlockID:    tt.fields.LastSortedBlockID,
				SortedBlocksmithsMap: tt.fields.SortedBlocksmithsMap,
				SpineBlockQuery:      tt.fields.SpineBlockQuery,
			}
			bss.SortedBlocksmiths = make([]*model.Blocksmith, 0)
			bss.SortBlocksmiths(tt.args.block, true)
			if len(bss.SortedBlocksmiths) > 0 && !bytes.Equal(bss.SortedBlocksmiths[0].NodePublicKey,
				bssMockBlocksmiths[1].NodePublicKey) && !bytes.Equal(bss.SortedBlocksmiths[1].NodePublicKey,
				bssMockBlocksmiths[0].NodePublicKey) {
				t.Errorf("sorting fail")
			}
		})
	}
}

func TestBlocksmithStrategySpine_GetSortedBlocksmithsMap(t *testing.T) {
	var mockBlocksmithMap = make(map[string]*int64)
	for index, mockBlocksmith := range bssMockBlocksmiths {
		mockIndex := int64(index)
		mockBlocksmithMap[string(mockBlocksmith.NodePublicKey)] = &mockIndex
	}
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		SpinePublicKeyQuery   query.SpinePublicKeyQueryInterface
		Logger                *log.Logger
		SortedBlocksmiths     []*model.Blocksmith
		LastSortedBlockID     int64
		SortedBlocksmithsLock sync.RWMutex
		SortedBlocksmithsMap  map[string]*int64
		SpineBlockQuery       query.BlockQueryInterface
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]*int64
	}{
		{
			name: "success:noSorting",
			args: args{
				block: &model.Block{ID: 0},
			},
			fields: fields{
				QueryExecutor:        &mockQueryGetBlocksmithsSpineSuccessNoBlocksmith{},
				SpinePublicKeyQuery:  query.NewSpinePublicKeyQuery(),
				Logger:               log.New(),
				SortedBlocksmiths:    bssMockBlocksmiths,
				SortedBlocksmithsMap: mockBlocksmithMap,
				SpineBlockQuery:      query.NewBlockQuery(&chaintype.SpineChain{}),
			},
			want: mockBlocksmithMap,
		},
		{
			name: "success",
			args: args{
				block: &model.Block{ID: 1},
			},
			fields: fields{
				QueryExecutor:        &mockQueryGetBlocksmithsSpineSuccessNoBlocksmith{},
				SpinePublicKeyQuery:  query.NewSpinePublicKeyQuery(),
				Logger:               log.New(),
				SortedBlocksmiths:    bssMockBlocksmiths,
				SortedBlocksmithsMap: mockBlocksmithMap,
				SpineBlockQuery:      query.NewBlockQuery(&chaintype.SpineChain{}),
			},
			want: mockBlocksmithMap,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategySpine{
				QueryExecutor:         tt.fields.QueryExecutor,
				SpinePublicKeyQuery:   tt.fields.SpinePublicKeyQuery,
				Logger:                tt.fields.Logger,
				SortedBlocksmiths:     tt.fields.SortedBlocksmiths,
				LastSortedBlockID:     tt.fields.LastSortedBlockID,
				SortedBlocksmithsLock: tt.fields.SortedBlocksmithsLock,
				SortedBlocksmithsMap:  tt.fields.SortedBlocksmithsMap,
				SpineBlockQuery:       tt.fields.SpineBlockQuery,
			}
			if got := bss.GetSortedBlocksmithsMap(tt.args.block); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlocksmithStrategySpine.GetSortedBlocksmithsMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlocksmithStrategySpine_GetSortedBlocksmiths(t *testing.T) {
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		SpinePublicKeyQuery   query.SpinePublicKeyQueryInterface
		Logger                *log.Logger
		SortedBlocksmiths     []*model.Blocksmith
		LastSortedBlockID     int64
		SortedBlocksmithsLock sync.RWMutex
		SortedBlocksmithsMap  map[string]*int64
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*model.Blocksmith
	}{
		{
			name: "success : last sorted block id = incoming block id",
			args: args{
				block: &model.Block{ID: 0},
			},
			fields: fields{
				QueryExecutor:         nil,
				SpinePublicKeyQuery:   nil,
				Logger:                log.New(),
				SortedBlocksmiths:     bssMockBlocksmiths,
				SortedBlocksmithsMap:  nil,
				SortedBlocksmithsLock: sync.RWMutex{},
			},
			want: bssMockBlocksmiths,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategySpine{
				QueryExecutor:         tt.fields.QueryExecutor,
				SpinePublicKeyQuery:   tt.fields.SpinePublicKeyQuery,
				Logger:                tt.fields.Logger,
				SortedBlocksmiths:     tt.fields.SortedBlocksmiths,
				LastSortedBlockID:     tt.fields.LastSortedBlockID,
				SortedBlocksmithsLock: tt.fields.SortedBlocksmithsLock,
				SortedBlocksmithsMap:  tt.fields.SortedBlocksmithsMap,
			}
			if got := bss.GetSortedBlocksmiths(tt.args.block); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlocksmithStrategySpine.GetSortedBlocksmiths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlocksmithStrategySpine_CalculateScore(t *testing.T) {
	type fields struct {
		QueryExecutor        query.ExecutorInterface
		SpinePublicKeyQuery  query.SpinePublicKeyQueryInterface
		Logger               *log.Logger
		SortedBlocksmiths    []*model.Blocksmith
		LastSortedBlockID    int64
		SortedBlocksmithsMap map[string]*int64
	}
	type args struct {
		lastBlock       *model.Block
		blocksmithIndex int64
		generator       *model.Blocksmith
		score           int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Blocksmith
		wantErr bool
	}{
		{
			name: "CalculateSmith:success",
			args: args{
				lastBlock:       mockBlock,
				blocksmithIndex: 0,
				generator: &model.Blocksmith{
					NodePublicKey: bssNodePubKey1,
					NodeID:        1,
				},
				score: constant.DefaultParticipationScore,
			},
			want: &model.Blocksmith{
				NodePublicKey: bssNodePubKey1,
				NodeID:        1,
				Score:         big.NewInt(1000000000),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategySpine{
				QueryExecutor:        tt.fields.QueryExecutor,
				SpinePublicKeyQuery:  tt.fields.SpinePublicKeyQuery,
				Logger:               tt.fields.Logger,
				SortedBlocksmiths:    tt.fields.SortedBlocksmiths,
				LastSortedBlockID:    tt.fields.LastSortedBlockID,
				SortedBlocksmithsMap: tt.fields.SortedBlocksmithsMap,
			}
			err := bss.CalculateScore(tt.args.generator, tt.args.score)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlocksmithStrategySpine.CalculateSmith() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := tt.args.generator
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlocksmithStrategySpine.CalculateSmith() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBlocksmithStrategySpine(t *testing.T) {
	type args struct {
		queryExecutor       query.ExecutorInterface
		spinePublicKeyQuery query.SpinePublicKeyQueryInterface
		logger              *log.Logger
		spineBlockQuery     query.BlockQueryInterface
	}
	tests := []struct {
		name string
		args args
		want *BlocksmithStrategySpine
	}{
		{
			name: "Success",
			args: args{
				logger: nil,
			},
			want: NewBlocksmithStrategySpine(nil, nil, nil, nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlocksmithStrategySpine(tt.args.queryExecutor, tt.args.spinePublicKeyQuery,
				tt.args.logger, tt.args.spineBlockQuery); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlocksmithStrategySpine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlocksmithStrategySpine_IsBlockTimestampValid(t *testing.T) {
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		SpinePublicKeyQuery   query.SpinePublicKeyQueryInterface
		Logger                *log.Logger
		SortedBlocksmiths     []*model.Blocksmith
		LastSortedBlockID     int64
		SortedBlocksmithsLock sync.RWMutex
		SortedBlocksmithsMap  map[string]*int64
		SpineBlockQuery       query.BlockQueryInterface
	}
	type args struct {
		blocksmithIndex     int64
		numberOfBlocksmiths int64
		previousBlock       *model.Block
		currentBlock        *model.Block
	}
	spine := &chaintype.SpineChain{}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "BlocksmithIndex = 0 - true",
			fields: fields{},
			args: args{
				blocksmithIndex:     0,
				numberOfBlocksmiths: 0,
				previousBlock: &model.Block{
					Timestamp: 0,
				},
				currentBlock: &model.Block{
					Timestamp: spine.GetSmithingPeriod(),
				},
			},
			wantErr: false,
		},
		{
			name:   "BlocksmithIndex > 0 - true",
			fields: fields{},
			args: args{
				blocksmithIndex:     1,
				numberOfBlocksmiths: 0,
				previousBlock: &model.Block{
					Timestamp: 0,
				},
				currentBlock: &model.Block{
					Timestamp: spine.GetSmithingPeriod() + (1 * spine.GetBlocksmithTimeGap()),
				},
			},
			wantErr: false,
		},
		{
			name:   "BlocksmithIndex > 0 - false",
			fields: fields{},
			args: args{
				blocksmithIndex:     1,
				numberOfBlocksmiths: 0,
				previousBlock: &model.Block{
					Timestamp: 0,
				},
				currentBlock: &model.Block{
					Timestamp: spine.GetSmithingPeriod(),
				},
			},
			wantErr: true,
		},
		{
			name:   "BlocksmithIndex = 0 - false",
			fields: fields{},
			args: args{
				blocksmithIndex:     1,
				numberOfBlocksmiths: 0,
				previousBlock: &model.Block{
					Timestamp: 0,
				},
				currentBlock: &model.Block{
					Timestamp: spine.GetSmithingPeriod() - 1,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategySpine{
				QueryExecutor:         tt.fields.QueryExecutor,
				SpinePublicKeyQuery:   tt.fields.SpinePublicKeyQuery,
				Logger:                tt.fields.Logger,
				SortedBlocksmiths:     tt.fields.SortedBlocksmiths,
				LastSortedBlockID:     tt.fields.LastSortedBlockID,
				SortedBlocksmithsLock: tt.fields.SortedBlocksmithsLock,
				SortedBlocksmithsMap:  tt.fields.SortedBlocksmithsMap,
				SpineBlockQuery:       tt.fields.SpineBlockQuery,
			}
			if err := bss.IsBlockTimestampValid(tt.args.blocksmithIndex, tt.args.numberOfBlocksmiths, tt.args.previousBlock,
				tt.args.currentBlock); (err != nil) != tt.wantErr {
				t.Errorf("IsBlockTimestampValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlocksmithStrategySpine_CanPersistBlock(t *testing.T) {
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		SpinePublicKeyQuery   query.SpinePublicKeyQueryInterface
		Logger                *log.Logger
		SortedBlocksmiths     []*model.Blocksmith
		LastSortedBlockID     int64
		SortedBlocksmithsLock sync.RWMutex
		SortedBlocksmithsMap  map[string]*int64
		SpineBlockQuery       query.BlockQueryInterface
	}
	type args struct {
		blocksmithIndex     int64
		numberOfBlocksmiths int64
		previousBlock       *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Default - Spinechain Does not need persist block function",
			fields:  fields{},
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bl := &BlocksmithStrategySpine{
				QueryExecutor:         tt.fields.QueryExecutor,
				SpinePublicKeyQuery:   tt.fields.SpinePublicKeyQuery,
				Logger:                tt.fields.Logger,
				SortedBlocksmiths:     tt.fields.SortedBlocksmiths,
				LastSortedBlockID:     tt.fields.LastSortedBlockID,
				SortedBlocksmithsLock: tt.fields.SortedBlocksmithsLock,
				SortedBlocksmithsMap:  tt.fields.SortedBlocksmithsMap,
				SpineBlockQuery:       tt.fields.SpineBlockQuery,
			}
			if err := bl.CanPersistBlock(tt.args.blocksmithIndex, tt.args.numberOfBlocksmiths, tt.args.previousBlock); (err != nil) != tt.wantErr {
				t.Errorf("CanPersistBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlocksmithStrategySpine_IsValidSmithTime(t *testing.T) {
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		SpinePublicKeyQuery   query.SpinePublicKeyQueryInterface
		Logger                *log.Logger
		SortedBlocksmiths     []*model.Blocksmith
		LastSortedBlockID     int64
		SortedBlocksmithsLock sync.RWMutex
		SortedBlocksmithsMap  map[string]*int64
		SpineBlockQuery       query.BlockQueryInterface
	}
	type args struct {
		blocksmithIndex     int64
		numberOfBlocksmiths int64
		previousBlock       *model.Block
	}
	spine := &chaintype.SpineChain{}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "TimeSinceLastBlock < SmithPeriod",
			fields: fields{},
			args: args{
				blocksmithIndex:     0,
				numberOfBlocksmiths: 3,
				previousBlock: &model.Block{
					ID:        1,
					Timestamp: time.Now().Unix(),
				},
			},
			wantErr: true,
		},
		{
			name:   "SmithingPending",
			fields: fields{},
			args: args{
				blocksmithIndex:     0,
				numberOfBlocksmiths: 100,
				previousBlock: &model.Block{
					ID: 1,
					Timestamp: time.Now().Unix() - spine.GetSmithingPeriod() - spine.GetBlocksmithBlockCreationTime() -
						spine.GetBlocksmithNetworkTolerance() - 1,
				},
			},
			wantErr: true,
		},
		{
			name:   "allowedBegin-one round",
			fields: fields{},
			args: args{
				blocksmithIndex:     1,
				numberOfBlocksmiths: 6,
				previousBlock: &model.Block{
					ID: 1,
					Timestamp: time.Now().Unix() - spine.GetSmithingPeriod() -
						spine.GetBlocksmithTimeGap() - 1,
				},
			},
			wantErr: false,
		},
		{
			name:   "allowedBegin-multiple round",
			fields: fields{},
			args: args{
				blocksmithIndex:     1,
				numberOfBlocksmiths: 6,
				previousBlock: &model.Block{
					ID: 1,
					Timestamp: time.Now().Unix() - spine.GetSmithingPeriod() -
						(11 * spine.GetBlocksmithTimeGap()) - 1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategySpine{
				QueryExecutor:         tt.fields.QueryExecutor,
				SpinePublicKeyQuery:   tt.fields.SpinePublicKeyQuery,
				Logger:                tt.fields.Logger,
				SortedBlocksmiths:     tt.fields.SortedBlocksmiths,
				LastSortedBlockID:     tt.fields.LastSortedBlockID,
				SortedBlocksmithsLock: tt.fields.SortedBlocksmithsLock,
				SortedBlocksmithsMap:  tt.fields.SortedBlocksmithsMap,
				SpineBlockQuery:       tt.fields.SpineBlockQuery,
			}
			if err := bss.IsValidSmithTime(tt.args.blocksmithIndex, tt.args.numberOfBlocksmiths, tt.args.previousBlock); (err != nil) != tt.wantErr {
				t.Errorf("IsValidSmithTime() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
