package strategy

import (
	"database/sql"
	"errors"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/storage"
	"math/big"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

var (
	bssNodePubKey1 = []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
	bssNodePubKey2 = []byte{1, 2, 3, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
	bssMockBlocksmiths = []*model.Blocksmith{
		{
			NodePublicKey: bssNodePubKey1,
			NodeID:        2,
			NodeOrder:     new(big.Int).SetInt64(1000),
			Score:         new(big.Int).SetInt64(1000),
		},
		{
			NodePublicKey: bssNodePubKey2,
			NodeID:        3,
			NodeOrder:     new(big.Int).SetInt64(2000),
			Score:         new(big.Int).SetInt64(2000),
		},
		{
			NodePublicKey: bssMockBlockData.BlocksmithPublicKey,
			NodeID:        4,
			NodeOrder:     new(big.Int).SetInt64(3000),
			Score:         new(big.Int).SetInt64(3000),
		},
	}
	bssMockBlockData = model.Block{
		ID:        constant.MainchainGenesisBlockID,
		BlockHash: make([]byte, 32),
		PreviousBlockHash: []byte{167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160,
			97, 214, 98, 245, 128, 255, 77, 228, 59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74},
		Height:    1,
		Timestamp: 1,
		BlockSeed: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		BlockSignature:       []byte{144, 246, 37, 144, 213, 135},
		CumulativeDifficulty: "1000",
		PayloadLength:        1,
		PayloadHash:          []byte{},
		BlocksmithPublicKey: []byte{1, 2, 3, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		TotalAmount:   1000,
		TotalFee:      0,
		TotalCoinBase: 1,
		Version:       0,
	}
)

type (
	mockQueryGetBlocksmithsMainSuccessNoBlocksmith struct {
		query.Executor
	}
	mockQueryGetBlocksmithsMainSuccessWithBlocksmith struct {
		query.Executor
	}

	mockQuerySortBlocksmithMainSuccessWithBlocksmiths struct {
		query.Executor
	}
	mockQueryGetBlocksmithsMainFail struct {
		query.Executor
	}
)

func (*mockQueryGetBlocksmithsMainFail) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	return nil, errors.New("mockError")
}

func (*mockQueryGetBlocksmithsMainSuccessNoBlocksmith) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mockNodeRegistrationQuery := query.NewNodeRegistrationQuery()
	mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows(mockNodeRegistrationQuery.Fields))
	rows, _ := db.Query("foo")
	return rows, nil
}

func (*mockQuerySortBlocksmithMainSuccessWithBlocksmiths) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows(
		[]string{"NodeID", "PublicKey", "Score", "maxHeight"},
	).AddRow(
		bssMockBlocksmiths[0].NodeID,
		bssMockBlocksmiths[0].NodePublicKey,
		bssMockBlocksmiths[0].Score.String(),
		uint32(1),
	).AddRow(
		bssMockBlocksmiths[1].NodeID,
		bssMockBlocksmiths[1].NodePublicKey,
		bssMockBlocksmiths[1].Score.String(),
		uint32(1),
	))
	rows, _ := db.Query("foo")
	return rows, nil
}

func (*mockQueryGetBlocksmithsMainSuccessWithBlocksmith) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("foo").WillReturnRows(sqlmock.NewRows(
		[]string{"NodeID", "PublicKey", "Score", "maxHeight"},
	).AddRow(
		bssMockBlocksmiths[0].NodeID,
		bssMockBlocksmiths[0].NodePublicKey,
		bssMockBlocksmiths[0].Score.String(),
		uint32(1),
	))
	rows, _ := db.Query("foo")
	return rows, nil
}

type (
	mockActiveNodeRegistryCacheSuccess struct {
		storage.NodeRegistryCacheStorage
	}
	mockActiveNodeRegistryCacheSuccessWithContent struct {
		storage.NodeRegistryCacheStorage
	}
)

func (*mockActiveNodeRegistryCacheSuccess) GetAllItems(item interface{}) error {
	castedItem := item.(*[]storage.NodeRegistry)
	*castedItem = make([]storage.NodeRegistry, 0)
	return nil
}

func (*mockActiveNodeRegistryCacheSuccessWithContent) GetAllItems(item interface{}) error {
	castedItem := item.(*[]storage.NodeRegistry)
	*castedItem = []storage.NodeRegistry{
		{
			Node: model.NodeRegistration{
				NodeID:        bssMockBlocksmiths[0].NodeID,
				NodePublicKey: bssMockBlocksmiths[0].NodePublicKey,
				Latest:        true,
				Height:        0,
			},
			ParticipationScore: bssMockBlocksmiths[0].Score.Int64(),
		},
	}
	return nil
}

func TestNewBlocksmithStrategy(t *testing.T) {
	type args struct {
		queryExecutor           query.ExecutorInterface
		nodeRegistrationQuery   query.NodeRegistrationQueryInterface
		skippedBlocksmithQuery  query.SkippedBlocksmithQueryInterface
		logger                  *log.Logger
		activeNodeRegistryCache storage.CacheStorageInterface
		chaintype               chaintype.ChainType
		rng                     *crypto.RandomNumberGenerator
	}
	tests := []struct {
		name string
		args args
		want *BlocksmithStrategyMain
	}{
		{
			name: "Success",
			args: args{
				logger: nil,
			},
			want: NewBlocksmithStrategyMain(
				nil, nil, nil, nil, nil, nil, nil, nil,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlocksmithStrategyMain(tt.args.queryExecutor, tt.args.nodeRegistrationQuery,
				tt.args.skippedBlocksmithQuery, tt.args.logger, nil, tt.args.activeNodeRegistryCache, tt.args.rng,
				tt.args.chaintype); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlocksmithStrategyMain() = %v, want %v", got, tt.want)
			}
		})
	}
}
