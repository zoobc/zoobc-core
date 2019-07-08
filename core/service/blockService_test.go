package service

import (
	"database/sql"
	"errors"
	"math/big"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/google/go-cmp/cmp"

	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

func TestNewBlockService(t *testing.T) {
	type args struct {
		chaintype     contract.ChainType
		queryExecutor query.ExecutorInterface
		blockQuery    query.BlockQueryInterface
	}
	test := struct {
		name string
		args args
		want *BlockService
	}{
		name: "NewBlockService:success",
		args: args{
			chaintype:     &chaintype.MainChain{},
			queryExecutor: nil,
			blockQuery:    nil,
		},
		want: &BlockService{
			Chaintype:     &chaintype.MainChain{},
			QueryExecutor: nil,
			BlockQuery:    nil,
		},
	}
	got := NewBlockService(test.args.chaintype, test.args.queryExecutor, test.args.blockQuery)

	if !cmp.Equal(got, test.want) {
		t.Errorf("NewBlockService() = %v, want %v", got, test.want)
	}

}

func TestBlockService_NewBlock(t *testing.T) {
	type fields struct {
		Chaintype     contract.ChainType
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
	}
	type args struct {
		version             uint32
		previousBlockHash   []byte
		blockSeed           []byte
		blocksmithID        []byte
		hash                string
		previousBlockHeight uint32
		timestamp           int64
		totalAmount         int64
		totalFee            int64
		totalCoinBase       int64
		transactions        []*model.Transaction
		payloadHash         []byte
		secretPhrase        string
	}
	test := struct {
		name   string
		fields fields
		args   args
		want   *model.Block
	}{
		name: "NewBlock:success",
		fields: fields{
			Chaintype:     &chaintype.MainChain{},
			QueryExecutor: nil,
			BlockQuery:    nil,
		},
		args: args{
			version:             1,
			previousBlockHash:   []byte{},
			blockSeed:           []byte{},
			blocksmithID:        []byte{},
			hash:                "hash",
			previousBlockHeight: 0,
			timestamp:           15875392,
			totalAmount:         0,
			totalFee:            0,
			totalCoinBase:       0,
			transactions:        []*model.Transaction{},
			payloadHash:         []byte{},
			secretPhrase:        "secretphrase",
		},
		want: &model.Block{
			Version:           1,
			PreviousBlockHash: []byte{},
			BlockSeed:         []byte{},
			BlocksmithID:      []byte{},
			Timestamp:         15875392,
			TotalAmount:       0,
			TotalFee:          0,
			TotalCoinBase:     0,
			Transactions:      []*model.Transaction{},
			PayloadHash:       []byte{},
		},
	}
	b := &BlockService{
		Chaintype:     test.fields.Chaintype,
		QueryExecutor: test.fields.QueryExecutor,
		BlockQuery:    test.fields.BlockQuery,
	}
	if got := b.NewBlock(test.args.version, test.args.previousBlockHash, test.args.blockSeed, test.args.blocksmithID, test.args.hash, test.args.previousBlockHeight, test.args.timestamp, test.args.totalAmount, test.args.totalFee, test.args.totalCoinBase, test.args.transactions, test.args.payloadHash, test.args.secretPhrase); !reflect.DeepEqual(got, test.want) {
		t.Errorf("BlockService.NewBlock() = %v, want %v", got, test.want)
	}

}

func TestBlockService_NewGenesisBlock(t *testing.T) {
	type fields struct {
		Chaintype     contract.ChainType
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
	}
	type args struct {
		version              uint32
		previousBlockHash    []byte
		blockSeed            []byte
		blocksmithID         []byte
		hash                 string
		previousBlockHeight  uint32
		timestamp            int64
		totalAmount          int64
		totalFee             int64
		totalCoinBase        int64
		transactions         []*model.Transaction
		payloadHash          []byte
		smithScale           int64
		cumulativeDifficulty *big.Int
		genesisSignature     []byte
	}
	test := struct {
		name   string
		fields fields
		args   args
		want   *model.Block
	}{
		name: "NewBlockGenesis:success",
		fields: fields{
			Chaintype:     &chaintype.MainChain{},
			QueryExecutor: nil,
			BlockQuery:    nil,
		},
		args: args{
			version:              1,
			previousBlockHash:    []byte{},
			blockSeed:            []byte{},
			blocksmithID:         []byte{},
			hash:                 "hash",
			previousBlockHeight:  0,
			timestamp:            15875392,
			totalAmount:          0,
			totalFee:             0,
			totalCoinBase:        0,
			transactions:         []*model.Transaction{},
			payloadHash:          []byte{},
			smithScale:           0,
			cumulativeDifficulty: big.NewInt(1),
			genesisSignature:     []byte{},
		},
		want: &model.Block{
			Version:              1,
			PreviousBlockHash:    []byte{},
			BlockSeed:            []byte{},
			BlocksmithID:         []byte{},
			Timestamp:            15875392,
			TotalAmount:          0,
			TotalFee:             0,
			TotalCoinBase:        0,
			Transactions:         []*model.Transaction{},
			PayloadHash:          []byte{},
			SmithScale:           0,
			CumulativeDifficulty: "1",
			BlockSignature:       []byte{},
		},
	}
	b := &BlockService{
		Chaintype:     test.fields.Chaintype,
		QueryExecutor: test.fields.QueryExecutor,
		BlockQuery:    test.fields.BlockQuery,
	}
	if got := b.NewGenesisBlock(test.args.version, test.args.previousBlockHash, test.args.blockSeed, test.args.blocksmithID, test.args.hash, test.args.previousBlockHeight, test.args.timestamp, test.args.totalAmount, test.args.totalFee, test.args.totalCoinBase, test.args.transactions, test.args.payloadHash, test.args.smithScale, test.args.cumulativeDifficulty, test.args.genesisSignature); !reflect.DeepEqual(got, test.want) {
		t.Errorf("BlockService.NewGenesisBlock() = %v, want %v", got, test.want)
	}
}

func TestBlockService_VerifySeed(t *testing.T) {
	type fields struct {
		Chaintype     contract.ChainType
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
	}
	type args struct {
		seed          *big.Int
		balance       *big.Int
		previousBlock model.Block
		timestamp     int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "VerifySeed:true-{prevTarget<seed<target && elapsed < 300}",
			fields: fields{
				Chaintype: &chaintype.MainChain{},
			},
			args: args{
				seed:    big.NewInt(1200),
				balance: big.NewInt(100),
				previousBlock: model.Block{
					Timestamp:  0,
					SmithScale: 10,
				},
				timestamp: 2,
			},
			want: true,
		},
		{
			name: "VerifySeed:true-{elapsedTime>300 && seed < target ",
			fields: fields{
				Chaintype: &chaintype.MainChain{},
			},
			args: args{
				seed:    big.NewInt(0),
				balance: big.NewInt(0),
				previousBlock: model.Block{
					Timestamp:  0,
					SmithScale: 0,
				},
				timestamp: 301,
			},
			want: false,
		},
		{
			name: "VerifySeed:true-{elapsedTime>300 && previousTarget > seed < target}",
			fields: fields{
				Chaintype: &chaintype.MainChain{},
			},
			args: args{
				seed:    big.NewInt(10),
				balance: big.NewInt(10),
				previousBlock: model.Block{
					Timestamp:  0,
					SmithScale: 10,
				},
				timestamp: 301,
			},
			want: true,
		},
		{
			name: "VerifySeed:false-{seed > target}",
			fields: fields{
				Chaintype: &chaintype.MainChain{},
			},
			args: args{
				seed:    big.NewInt(10000),
				balance: big.NewInt(10),
				previousBlock: model.Block{
					Timestamp:  0,
					SmithScale: 10,
				},
				timestamp: 0,
			},
			want: false,
		},
		{
			name: "VerifySeed:false-{seed < prevtarget}",
			fields: fields{
				Chaintype: &chaintype.MainChain{},
			},
			args: args{
				seed:    big.NewInt(0),
				balance: big.NewInt(10),
				previousBlock: model.Block{
					Timestamp:  0,
					SmithScale: 10,
				},
				timestamp: 0,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlockService{
				Chaintype:     tt.fields.Chaintype,
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
			}
			if got := b.VerifySeed(tt.args.seed, tt.args.balance, tt.args.previousBlock, tt.args.timestamp); got != tt.want {
				t.Errorf("BlockService.VerifySeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockQueryExecutorSuccess struct{}

func (*mockQueryExecutorSuccess) Execute(string) (sql.Result, error) {
	return nil, nil
}
func (*mockQueryExecutorSuccess) ExecuteSelect(query string) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if query == `SELECT (id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, payload_length, payload_hash, blocksmith_id, total_amount, total_fee, total_coinbase, version) FROM main_block ORDER BY height DESC LIMIT 1` {
		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT (id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, payload_length, payload_hash, blocksmith_id, total_amount, total_fee, total_coinbase, version)
		FROM main_block ORDER BY height DESC LIMIT 1`)).WillReturnRows(sqlmock.NewRows([]string{"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase", "Version"}).AddRow(1, []byte{}, 1, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1))
	} else if query == `SELECT (id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, payload_length, payload_hash, blocksmith_id, total_amount, total_fee, total_coinbase, version) FROM main_block WHERE height = 0 LIMIT 1` {
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT (id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, payload_length, payload_hash, blocksmith_id, total_amount, total_fee, total_coinbase, version) FROM main_block WHERE height = 0 LIMIT 1
		`)).WillReturnRows(sqlmock.NewRows([]string{"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase", "Version"}).
			AddRow(1, []byte{}, 0, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1))
	} else if query == "SELECT (id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, payload_length, payload_hash, blocksmith_id, total_amount, total_fee, total_coinbase, version) FROM main_block" {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT (id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, payload_length, payload_hash, blocksmith_id, total_amount, total_fee, total_coinbase, version) FROM main_block`)).WillReturnRows(sqlmock.NewRows([]string{"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase", "Version"}).
			AddRow(1, []byte{}, 0, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1))
	}

	rows, _ := db.Query(query)
	return rows, nil
}
func (*mockQueryExecutorSuccess) ExecuteTransactions(queries []string) ([]sql.Result, error) {
	return nil, nil
}
func (*mockQueryExecutorSuccess) ExecuteStatement(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

type MockBlockQuery struct{}

func (*MockBlockQuery) GetBlocks(height, size uint32) string {
	return "SELECT (id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, payload_length, payload_hash, blocksmith_id, total_amount, total_fee, total_coinbase, version) FROM main_block"
}
func (*MockBlockQuery) GetLastBlock() string {
	return "SELECT (id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, payload_length, payload_hash, blocksmith_id, total_amount, total_fee, total_coinbase, version) FROM main_block ORDER BY height DESC LIMIT 1"
}
func (*MockBlockQuery) GetGenesisBlock() string {
	return "SELECT (id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, payload_length, payload_hash, blocksmith_id, total_amount, total_fee, total_coinbase, version) FROM main_block WHERE height = 0 LIMIT 1"
}
func (*MockBlockQuery) GetBlockByID(int64) string                    { return "" }
func (*MockBlockQuery) GetBlockByHeight(uint32) string               { return "" }
func (*MockBlockQuery) InsertBlock() string                          { return "" }
func (*MockBlockQuery) ExtractModel(block model.Block) []interface{} { return []interface{}{1} }

type mockQueryExecutorFail struct {
	mockQueryExecutorSuccess
}

func (*mockQueryExecutorFail) Execute(string) (sql.Result, error) {
	return nil, errors.New("MockedError")
}
func (*mockQueryExecutorFail) ExecuteSelect(string) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}
func (*mockQueryExecutorFail) ExecuteTransactions(queries []string) ([]sql.Result, error) {
	return nil, errors.New("MockedError")
}
func (*mockQueryExecutorFail) ExecuteStatement(query string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("MockedError")
}

type mockQueryExecutorSqlFail struct {
	mockQueryExecutorSuccess
}

func (*mockQueryExecutorSqlFail) ExecuteSelect(query string) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
		"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase", "Version"}))
	rows, _ := db.Query(query)
	return rows, nil
}

func TestBlockService_PushBlock(t *testing.T) {
	type fields struct {
		Chaintype     contract.ChainType
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
	}
	type args struct {
		previousBlock model.Block
		block         model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "PushBlock:success",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorSuccess{},
				BlockQuery:    &MockBlockQuery{},
			},
			args: args{
				previousBlock: model.Block{
					ID:                   0,
					SmithScale:           10,
					Timestamp:            10000,
					CumulativeDifficulty: "10000",
				},
				block: model.Block{
					ID:        1,
					Timestamp: 12000,
				},
			},
			wantErr: false,
		},
		{
			name: "PushBlock:fail-{QueryExecutor:fail}",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorFail{},
				BlockQuery:    &MockBlockQuery{},
			},
			args: args{
				previousBlock: model.Block{
					ID:                   0,
					SmithScale:           10,
					Timestamp:            10000,
					CumulativeDifficulty: "10000",
				},
				block: model.Block{
					ID:        1,
					Timestamp: 12000,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:     tt.fields.Chaintype,
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
			}
			if err := bs.PushBlock(tt.args.previousBlock, tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("BlockService.PushBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockService_GetLastBlock(t *testing.T) {
	type fields struct {
		Chaintype     contract.ChainType
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
	}
	tests := []struct {
		name    string
		fields  fields
		want    model.Block
		wantErr bool
	}{
		{
			name: "GetLastBlock:success", // All is good
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorSuccess{},
				BlockQuery:    &MockBlockQuery{},
			},
			want: model.Block{
				ID:                   1,
				PreviousBlockHash:    []byte{},
				Height:               1,
				Timestamp:            10000,
				BlockSeed:            []byte{},
				BlockSignature:       []byte{},
				CumulativeDifficulty: "",
				SmithScale:           1,
				PayloadLength:        2,
				PayloadHash:          []byte{},
				BlocksmithID:         []byte{},
				TotalAmount:          0,
				TotalFee:             0,
				TotalCoinBase:        0,
				Version:              1,
			},
			wantErr: false,
		},
		{
			name: "GetLastBlock:fail", // ExecuteSelect return error != nil
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorFail{},
				BlockQuery:    &MockBlockQuery{},
			},
			want: model.Block{
				ID: -1,
			},
			wantErr: true,
		},
		{
			name: "GetLastBlock:fail-{sql.rows.Next = false}", // block not found | rows.Next() -> false
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorSqlFail{},
				BlockQuery:    &MockBlockQuery{},
			},
			want: model.Block{
				ID: -1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:     tt.fields.Chaintype,
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
			}
			got, err := bs.GetLastBlock()
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetLastBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetLastBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_GetGenesisBlock(t *testing.T) {
	type fields struct {
		Chaintype     contract.ChainType
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		Blocks        []model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		want    model.Block
		wantErr bool
	}{
		{
			name: "GetGenesisBlock:success",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorSuccess{},
				BlockQuery:    &MockBlockQuery{},
			},
			want: model.Block{
				ID:                   1,
				PreviousBlockHash:    []byte{},
				Height:               0,
				Timestamp:            10000,
				BlockSeed:            []byte{},
				BlockSignature:       []byte{},
				CumulativeDifficulty: "",
				SmithScale:           1,
				PayloadLength:        2,
				PayloadHash:          []byte{},
				BlocksmithID:         []byte{},
				TotalAmount:          0,
				TotalFee:             0,
				TotalCoinBase:        0,
				Version:              1,
			},
			wantErr: false,
		},
		{
			name: "GetGenesis:fail",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorFail{},
				BlockQuery:    &MockBlockQuery{},
			},
			want: model.Block{
				ID: -1,
			},
			wantErr: true,
		},
		{
			name: "GetGenesis:fail-{sql.rows.Next = false}", // genesis not found | rows.Next() -> false
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorSqlFail{},
				BlockQuery:    &MockBlockQuery{},
			},
			want: model.Block{
				ID: -1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:     tt.fields.Chaintype,
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
			}
			got, err := bs.GetGenesisBlock()
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetGenesisBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetGenesisBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_GetBlocks(t *testing.T) {
	type fields struct {
		Chaintype     contract.ChainType
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		Blocks        []model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		want    []model.Block
		wantErr bool
	}{
		{
			name: "GetBlocks:success",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorSuccess{},
				BlockQuery:    &MockBlockQuery{},
			},
			want: []model.Block{
				{
					ID:                   1,
					PreviousBlockHash:    []byte{},
					Height:               0,
					Timestamp:            10000,
					BlockSeed:            []byte{},
					BlockSignature:       []byte{},
					CumulativeDifficulty: "",
					SmithScale:           1,
					PayloadLength:        2,
					PayloadHash:          []byte{},
					BlocksmithID:         []byte{},
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Version:              1,
				},
			},
			wantErr: false,
		},
		{
			name: "GetBlocks:fail",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorFail{},
				BlockQuery:    &MockBlockQuery{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:     tt.fields.Chaintype,
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
			}
			got, err := bs.GetBlocks()
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetBlocks() = %v, want %v", got, tt.want)
			}
		})
	}
}
