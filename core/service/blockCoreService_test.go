package service

import (
	"database/sql"
	"errors"
	"math/big"
	"reflect"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	mockSignature struct {
		crypto.Signature
	}
	mockQueryExecutorSuccess struct {
		query.Executor
	}
	mockQueryExecuteNotNil struct {
		query.Executor
	}
	mockQueryExecutorScanFail struct {
		query.Executor
	}
	mockQueryExecutorFail struct {
		query.Executor
	}
	mockTypeAction struct {
		transaction.SendMoney
	}
	mockTypeActionSuccess struct {
		mockTypeAction
	}
)

// mockTypeAction
func (*mockTypeAction) ApplyConfirmed() error {
	return nil
}
func (*mockTypeAction) Validate() error {
	return nil
}
func (*mockTypeAction) GetAmount() int64 {
	return 10
}
func (*mockTypeActionSuccess) GetTransactionType(tx *model.Transaction) transaction.TypeAction {
	return &mockTypeAction{}
}

// mockSignature
func (*mockSignature) SignBlock(payload []byte, nodeSeed string) []byte {
	return []byte{}
}
func (*mockSignature) VerifySignature(
	payload, signature []byte,
	accountType uint32,
	accountAddress string,
) bool {
	return true
}

// mockQueryExecutorScanFail
func (*mockQueryExecutorScanFail) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{
		"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
		"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase"}))
	rows, _ := db.Query(qe)
	return rows, nil
}

// mockQueryExecutorNotNil
func (*mockQueryExecuteNotNil) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"ID"}))
	return db.Query("")
}

// mockQueryExecutorFail
func (*mockQueryExecutorFail) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}
func (*mockQueryExecutorFail) ExecuteStatement(qe string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("MockedError")
}
func (*mockQueryExecutorFail) BeginTx() error { return nil }

func (*mockQueryExecutorFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("mockError:deleteMempoolFail")
}
func (*mockQueryExecutorFail) CommitTx() error { return errors.New("mockError:commitFail") }

// mockQueryExecutorSuccess
func (*mockQueryExecutorSuccess) BeginTx() error { return nil }

func (*mockQueryExecutorSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}
func (*mockQueryExecutorSuccess) CommitTx() error { return nil }

func (*mockQueryExecutorSuccess) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, " +
		"payload_length, payload_hash, blocksmith_id, total_amount, total_fee, total_coinbase, version FROM main_block ORDER BY " +
		"height DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase",
			"Version"},
		).AddRow(1, []byte{}, 1, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1))
	case "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, " +
		"payload_length, payload_hash, blocksmith_id, total_amount, total_fee, total_coinbase, version FROM main_block " +
		"WHERE height = 0 LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase", "Version"}).
			AddRow(1, []byte{}, 0, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1))
	case "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, " +
		"payload_length, payload_hash, blocksmith_id, total_amount, total_fee, total_coinbase, version FROM main_block WHERE height >= 0 " +
		"LIMIT 100":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase", "Version"}).
			AddRow(1, []byte{}, 0, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1))
	}

	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockQueryExecutorSuccess) ExecuteStatement(qe string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func TestNewBlockService(t *testing.T) {
	type args struct {
		ct                  contract.ChainType
		queryExecutor       query.ExecutorInterface
		blockQuery          query.BlockQueryInterface
		mempoolQuery        query.MempoolQueryInterface
		transactionQuery    query.TransactionQueryInterface
		signature           crypto.SignatureInterface
		mempoolService      MempoolServiceInterface
		txTypeSwitcher      transaction.TypeActionSwitcher
		accountBalanceQuery query.AccountBalanceQueryInterface
		obsr                *observer.Observer
	}
	tests := []struct {
		name string
		args args
		want *BlockService
	}{
		{
			name: "wantSuccess",
			args: args{
				ct:   &chaintype.MainChain{},
				obsr: observer.NewObserver(),
			},
			want: &BlockService{
				Chaintype: &chaintype.MainChain{},
				Observer:  observer.NewObserver(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockService(
				tt.args.ct, tt.args.queryExecutor,
				tt.args.blockQuery,
				tt.args.mempoolQuery,
				tt.args.transactionQuery,
				tt.args.signature,
				tt.args.mempoolService,
				tt.args.txTypeSwitcher,
				tt.args.accountBalanceQuery,
				tt.args.obsr,
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_NewBlock(t *testing.T) {
	type fields struct {
		Chaintype          contract.ChainType
		QueryExecutor      query.ExecutorInterface
		BlockQuery         query.BlockQueryInterface
		MempoolQuery       query.MempoolQueryInterface
		TransactionQuery   query.TransactionQueryInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
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
		payloadLength       uint32
		secretPhrase        string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *model.Block
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Chaintype: &chaintype.MainChain{},
				Signature: &mockSignature{},
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
				payloadLength:       0,
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
				PayloadLength:     0,
				BlockSignature:    []byte{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:          tt.fields.Chaintype,
				QueryExecutor:      tt.fields.QueryExecutor,
				BlockQuery:         tt.fields.BlockQuery,
				MempoolQuery:       tt.fields.MempoolQuery,
				TransactionQuery:   tt.fields.TransactionQuery,
				Signature:          tt.fields.Signature,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
			}
			if got := bs.NewBlock(
				tt.args.version,
				tt.args.previousBlockHash,
				tt.args.blockSeed,
				tt.args.blocksmithID,
				tt.args.hash,
				tt.args.previousBlockHeight,
				tt.args.timestamp,
				tt.args.totalAmount,
				tt.args.totalFee,
				tt.args.totalCoinBase,
				tt.args.transactions,
				tt.args.payloadHash,
				tt.args.payloadLength,
				tt.args.secretPhrase,
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.NewBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_NewGenesisBlock(t *testing.T) {
	type fields struct {
		Chaintype          contract.ChainType
		QueryExecutor      query.ExecutorInterface
		BlockQuery         query.BlockQueryInterface
		MempoolQuery       query.MempoolQueryInterface
		TransactionQuery   query.TransactionQueryInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
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
		payloadLength        uint32
		smithScale           int64
		cumulativeDifficulty *big.Int
		genesisSignature     []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *model.Block
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Chaintype: &chaintype.MainChain{},
				Signature: &mockSignature{},
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
				payloadLength:        8,
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
				PayloadLength:        8,
				SmithScale:           0,
				CumulativeDifficulty: "1",
				BlockSignature:       []byte{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:          tt.fields.Chaintype,
				QueryExecutor:      tt.fields.QueryExecutor,
				BlockQuery:         tt.fields.BlockQuery,
				MempoolQuery:       tt.fields.MempoolQuery,
				TransactionQuery:   tt.fields.TransactionQuery,
				Signature:          tt.fields.Signature,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
			}
			if got := bs.NewGenesisBlock(
				tt.args.version,
				tt.args.previousBlockHash,
				tt.args.blockSeed,
				tt.args.blocksmithID,
				tt.args.hash,
				tt.args.previousBlockHeight,
				tt.args.timestamp,
				tt.args.totalAmount,
				tt.args.totalFee,
				tt.args.totalCoinBase,
				tt.args.transactions,
				tt.args.payloadHash,
				tt.args.payloadLength,
				tt.args.smithScale,
				tt.args.cumulativeDifficulty,
				tt.args.genesisSignature,
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.NewGenesisBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_VerifySeed(t *testing.T) {
	type fields struct {
		Chaintype          contract.ChainType
		QueryExecutor      query.ExecutorInterface
		BlockQuery         query.BlockQueryInterface
		MempoolQuery       query.MempoolQueryInterface
		TransactionQuery   query.TransactionQueryInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
	}
	type args struct {
		seed          *big.Int
		balance       *big.Int
		previousBlock *model.Block
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
				previousBlock: &model.Block{
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
				previousBlock: &model.Block{
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
				previousBlock: &model.Block{
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
				previousBlock: &model.Block{
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
				previousBlock: &model.Block{
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
				Chaintype:          tt.fields.Chaintype,
				QueryExecutor:      tt.fields.QueryExecutor,
				BlockQuery:         tt.fields.BlockQuery,
				MempoolQuery:       tt.fields.MempoolQuery,
				TransactionQuery:   tt.fields.TransactionQuery,
				Signature:          tt.fields.Signature,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
			}
			if got := b.VerifySeed(tt.args.seed, tt.args.balance, tt.args.previousBlock, tt.args.timestamp); got != tt.want {
				t.Errorf("BlockService.VerifySeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_PushBlock(t *testing.T) {
	type fields struct {
		Chaintype          contract.ChainType
		QueryExecutor      query.ExecutorInterface
		BlockQuery         query.BlockQueryInterface
		MempoolQuery       query.MempoolQueryInterface
		TransactionQuery   query.TransactionQueryInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		Observer           *observer.Observer
	}
	type args struct {
		previousBlock *model.Block
		block         *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "PushBlock:Transactions<0",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
				Observer:      observer.NewObserver(),
			},
			args: args{
				previousBlock: &model.Block{
					ID:                   0,
					SmithScale:           10,
					Timestamp:            10000,
					CumulativeDifficulty: "10000",
					Version:              1,
					PreviousBlockHash:    []byte{},
					BlockSeed:            []byte{},
					BlocksmithID:         []byte{},
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					BlockSignature:       []byte{},
				},
				block: &model.Block{
					ID:                1,
					Timestamp:         12000,
					Version:           1,
					PreviousBlockHash: []byte{},
					BlockSeed:         []byte{},
					BlocksmithID:      []byte{},
					TotalAmount:       0,
					TotalFee:          0,
					TotalCoinBase:     0,
					Transactions:      []*model.Transaction{},
					PayloadHash:       []byte{},
					BlockSignature:    []byte{},
				},
			},
			wantErr: false,
		},
		{
			name: "PushBlock:ApplyConfirmedFailed",
			fields: fields{
				Chaintype:          &chaintype.MainChain{},
				QueryExecutor:      &mockQueryExecutorSuccess{},
				BlockQuery:         query.NewBlockQuery(&chaintype.MainChain{}),
				TransactionQuery:   query.NewTransactionQuery(&chaintype.MainChain{}),
				MempoolQuery:       query.NewMempoolQuery(&chaintype.MainChain{}),
				ActionTypeSwitcher: &transaction.TypeSwitcher{},
				Observer:           observer.NewObserver(),
			},
			args: args{
				previousBlock: &model.Block{
					ID:                   0,
					SmithScale:           10,
					Timestamp:            10000,
					CumulativeDifficulty: "10000",
					Version:              1,
					PreviousBlockHash:    []byte{},
					BlockSeed:            []byte{},
					BlocksmithID:         []byte{},
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					BlockSignature:       []byte{},
				},
				block: &model.Block{
					ID:                1,
					Timestamp:         12000,
					Version:           1,
					PreviousBlockHash: []byte{},
					BlockSeed:         []byte{},
					BlocksmithID:      []byte{},
					TotalAmount:       0,
					TotalFee:          0,
					TotalCoinBase:     0,
					Transactions: []*model.Transaction{
						{
							Height: 0,
						},
					},
					PayloadHash:    []byte{},
					BlockSignature: []byte{},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:          tt.fields.Chaintype,
				QueryExecutor:      tt.fields.QueryExecutor,
				BlockQuery:         tt.fields.BlockQuery,
				MempoolQuery:       tt.fields.MempoolQuery,
				TransactionQuery:   tt.fields.TransactionQuery,
				Signature:          tt.fields.Signature,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
				Observer:           tt.fields.Observer,
			}
			if err := bs.PushBlock(tt.args.previousBlock, tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("BlockService.PushBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockService_GetLastBlock(t *testing.T) {
	type fields struct {
		Chaintype          contract.ChainType
		QueryExecutor      query.ExecutorInterface
		BlockQuery         query.BlockQueryInterface
		MempoolQuery       query.MempoolQueryInterface
		TransactionQuery   query.TransactionQueryInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
	}
	tests := []struct {
		name    string
		fields  fields
		want    *model.Block
		wantErr bool
	}{
		{
			name: "GetLastBlock:Success", // All is good
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			want: &model.Block{
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
			name: "GetLastBlock:SelectFail",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorFail{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			want: &model.Block{
				ID: -1,
			},
			wantErr: true,
		},
		{
			name: "GetLastBlock:SelectGotNil",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecuteNotNil{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			want: &model.Block{
				ID: -1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:          tt.fields.Chaintype,
				QueryExecutor:      tt.fields.QueryExecutor,
				BlockQuery:         tt.fields.BlockQuery,
				MempoolQuery:       tt.fields.MempoolQuery,
				TransactionQuery:   tt.fields.TransactionQuery,
				Signature:          tt.fields.Signature,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
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
		Chaintype          contract.ChainType
		QueryExecutor      query.ExecutorInterface
		BlockQuery         query.BlockQueryInterface
		MempoolQuery       query.MempoolQueryInterface
		TransactionQuery   query.TransactionQueryInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
	}
	tests := []struct {
		name    string
		fields  fields
		want    *model.Block
		wantErr bool
	}{
		{
			name: "GetGenesisBlock:success",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			want: &model.Block{
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
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			want: &model.Block{
				ID: -1,
			},
			wantErr: true,
		},
		{
			name: "GetGenesis:fail-{sql.rows.Next = false}", // genesis not found | rows.Next() -> false
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorScanFail{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			want: &model.Block{
				ID: -1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:          tt.fields.Chaintype,
				QueryExecutor:      tt.fields.QueryExecutor,
				BlockQuery:         tt.fields.BlockQuery,
				MempoolQuery:       tt.fields.MempoolQuery,
				TransactionQuery:   tt.fields.TransactionQuery,
				Signature:          tt.fields.Signature,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
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
		Chaintype          contract.ChainType
		QueryExecutor      query.ExecutorInterface
		BlockQuery         query.BlockQueryInterface
		MempoolQuery       query.MempoolQueryInterface
		TransactionQuery   query.TransactionQueryInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*model.Block
		wantErr bool
	}{
		{
			name: "GetBlocks:success",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			want: []*model.Block{
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
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:          tt.fields.Chaintype,
				QueryExecutor:      tt.fields.QueryExecutor,
				BlockQuery:         tt.fields.BlockQuery,
				MempoolQuery:       tt.fields.MempoolQuery,
				TransactionQuery:   tt.fields.TransactionQuery,
				Signature:          tt.fields.Signature,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
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
func TestBlockService_RemoveMempoolTransactions(t *testing.T) {
	type fields struct {
		Chaintype     contract.ChainType
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		MempoolQuery  query.MempoolQueryInterface
		Signature     crypto.SignatureInterface
	}
	type args struct {
		transactions []*model.Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "RemoveMempoolTransaction:Success",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				MempoolQuery:  query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor: &mockQueryExecutorSuccess{},
			},
			args: args{
				transactions: []*model.Transaction{
					buildTransaction(1562893303, "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE", "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"),
				},
			},
			wantErr: false,
		},
		{
			name: "RemoveMempoolTransaction:Fail",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				MempoolQuery:  query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor: &mockQueryExecutorFail{},
			},
			args: args{
				transactions: []*model.Transaction{
					buildTransaction(1562893303, "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE", "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"),
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
				MempoolQuery:  tt.fields.MempoolQuery,
				Signature:     tt.fields.Signature,
			}
			if err := bs.RemoveMempoolTransactions(tt.args.transactions); (err != nil) != tt.wantErr {
				t.Errorf("BlockService.RemoveMempoolTransactions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockMempoolServiceSelectFail struct {
		MempoolService
	}
	mockMempoolServiceSelectWrongTransactionBytes struct {
		MempoolService
	}
	mockMempoolServiceSelectSuccess struct {
		MempoolService
	}
	mockQueryExecutorMempoolSuccess struct {
		query.Executor
	}
)

// mockQueryExecutorMempoolSuccess
func (*mockQueryExecutorMempoolSuccess) ExecuteSelect(qStr string, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
		"id",
		"fee_per_byte",
		"arrival_timestamp",
		"transaction_bytes",
	}).AddRow(
		1,
		1,
		123456,
		getTestSignedMempoolTransaction(1, 1562893305).TransactionBytes),
	)
	return db.Query("")
}

// mockMempoolServiceSelectSuccess
func (*mockMempoolServiceSelectSuccess) SelectTransactionFromMempool(
	blockTimestamp int64,
) ([]*model.MempoolTransaction, error) {
	return []*model.MempoolTransaction{
		{
			FeePerByte:       1,
			TransactionBytes: getTestSignedMempoolTransaction(1, 1562893305).TransactionBytes,
		},
	}, nil
}

// mockMempoolServiceSelectFail
func (*mockMempoolServiceSelectFail) SelectTransactionsFromMempool(
	blockTimestamp int64,
) ([]*model.MempoolTransaction, error) {
	return nil, errors.New("want error on select")
}

// mockMempoolServiceSelectSuccess
func (*mockMempoolServiceSelectWrongTransactionBytes) SelectTransactionsFromMempool(
	blockTimestamp int64,
) ([]*model.MempoolTransaction, error) {
	return []*model.MempoolTransaction{
		{
			FeePerByte: 1,
		},
	}, nil
}

func TestBlockService_GenerateBlock(t *testing.T) {
	type fields struct {
		Chaintype          contract.ChainType
		QueryExecutor      query.ExecutorInterface
		BlockQuery         query.BlockQueryInterface
		MempoolQuery       query.MempoolQueryInterface
		TransactionQuery   query.TransactionQueryInterface
		Signature          crypto.SignatureInterface
		MempoolService     MempoolServiceInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
	}
	type args struct {
		previousBlock       *model.Block
		secretPhrase        string
		timestamp           int64
		blockSmithAccountID []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Block
		wantErr bool
	}{
		{
			name: "wantFail:MempoolServiceSelectTransaction",
			fields: fields{
				Chaintype:      &chaintype.MainChain{},
				Signature:      &mockSignature{},
				MempoolQuery:   query.NewMempoolQuery(&chaintype.MainChain{}),
				MempoolService: &mockMempoolServiceSelectFail{},
			},
			args: args{
				previousBlock: &model.Block{
					Version:           1,
					PreviousBlockHash: []byte{},
					BlockSeed:         []byte{},
					BlocksmithID:      []byte{},
					Timestamp:         12344587645,
					TotalAmount:       0,
					TotalFee:          0,
					TotalCoinBase:     0,
					Transactions:      []*model.Transaction{},
					PayloadHash:       []byte{},
					PayloadLength:     0,
					BlockSignature:    []byte{},
				},
				secretPhrase:        "pharsepress",
				timestamp:           12344587645,
				blockSmithAccountID: []byte{1, 2, 4},
			},
			wantErr: true,
		},
		{
			name: "wantFail:ParseTransactionToByte",
			fields: fields{
				Chaintype:      &chaintype.MainChain{},
				Signature:      &mockSignature{},
				MempoolQuery:   query.NewMempoolQuery(&chaintype.MainChain{}),
				MempoolService: &mockMempoolServiceSelectWrongTransactionBytes{},
			},
			args: args{
				previousBlock: &model.Block{
					Version:           1,
					PreviousBlockHash: []byte{},
					BlockSeed:         []byte{},
					BlocksmithID:      []byte{},
					Timestamp:         12344587645,
					TotalAmount:       0,
					TotalFee:          0,
					TotalCoinBase:     0,
					Transactions:      []*model.Transaction{},
					PayloadHash:       []byte{},
					PayloadLength:     0,
					BlockSignature:    []byte{},
				},
				secretPhrase:        "pharsepress",
				timestamp:           12344587645,
				blockSmithAccountID: []byte{1, 2, 4},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess:ParseTransactionToByte",
			fields: fields{
				Chaintype:    &chaintype.MainChain{},
				Signature:    &mockSignature{},
				BlockQuery:   query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery: query.NewMempoolQuery(&chaintype.MainChain{}),
				MempoolService: &mockMempoolServiceSelectSuccess{
					MempoolService{
						QueryExecutor:      &mockQueryExecutorMempoolSuccess{},
						ActionTypeSwitcher: &mockTypeActionSuccess{},
					},
				},
				ActionTypeSwitcher: &mockTypeActionSuccess{},
			},
			args: args{
				previousBlock: &model.Block{
					Version:           1,
					PreviousBlockHash: []byte{},
					BlockSeed:         []byte{},
					BlocksmithID:      []byte{},
					Timestamp:         12344587645,
					TotalAmount:       0,
					TotalFee:          0,
					TotalCoinBase:     0,
					Transactions:      []*model.Transaction{},
					PayloadHash:       []byte{},
					PayloadLength:     0,
					BlockSignature:    []byte{},
				},
				secretPhrase:        "",
				timestamp:           12345678,
				blockSmithAccountID: []byte{1, 2, 4},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:          tt.fields.Chaintype,
				QueryExecutor:      tt.fields.QueryExecutor,
				BlockQuery:         tt.fields.BlockQuery,
				MempoolQuery:       tt.fields.MempoolQuery,
				TransactionQuery:   tt.fields.TransactionQuery,
				Signature:          tt.fields.Signature,
				MempoolService:     tt.fields.MempoolService,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
			}
			_, err := bs.GenerateBlock(
				tt.args.previousBlock,
				tt.args.secretPhrase,
				tt.args.timestamp,
				tt.args.blockSmithAccountID,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GenerateBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func TestBlockService_AddGenesis(t *testing.T) {
	type fields struct {
		Chaintype          contract.ChainType
		QueryExecutor      query.ExecutorInterface
		BlockQuery         query.BlockQueryInterface
		MempoolQuery       query.MempoolQueryInterface
		TransactionQuery   query.TransactionQueryInterface
		Signature          crypto.SignatureInterface
		MempoolService     MempoolServiceInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		Observer           *observer.Observer
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Chaintype:          &chaintype.MainChain{},
				Signature:          &mockSignature{},
				MempoolQuery:       query.NewMempoolQuery(&chaintype.MainChain{}),
				MempoolService:     &mockMempoolServiceSelectFail{},
				ActionTypeSwitcher: &mockTypeActionSuccess{},
				QueryExecutor:      &mockQueryExecutorSuccess{},
				BlockQuery:         query.NewBlockQuery(&chaintype.MainChain{}),
				TransactionQuery:   query.NewTransactionQuery(&chaintype.MainChain{}),
				Observer:           observer.NewObserver(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:          tt.fields.Chaintype,
				QueryExecutor:      tt.fields.QueryExecutor,
				BlockQuery:         tt.fields.BlockQuery,
				MempoolQuery:       tt.fields.MempoolQuery,
				TransactionQuery:   tt.fields.TransactionQuery,
				Signature:          tt.fields.Signature,
				MempoolService:     tt.fields.MempoolService,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
				Observer:           tt.fields.Observer,
			}
			if err := bs.AddGenesis(); (err != nil) != tt.wantErr {
				t.Errorf("BlockService.AddGenesis() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockQueryExecutorCheckGenesisTrue struct {
		query.Executor
	}
	mockQueryExecutorCheckGenesisFalse struct {
		query.Executor
	}
)

func (*mockQueryExecutorCheckGenesisFalse) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
		"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
		"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase",
		"Version",
	}))
	return db.Query("")
}
func (*mockQueryExecutorCheckGenesisTrue) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
		"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
		"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase",
		"Version",
	}).AddRow(4545420970999433273, []byte{}, 1, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1))
	return db.Query("")
}

func TestBlockService_CheckGenesis(t *testing.T) {
	type fields struct {
		Chaintype          contract.ChainType
		QueryExecutor      query.ExecutorInterface
		BlockQuery         query.BlockQueryInterface
		MempoolQuery       query.MempoolQueryInterface
		TransactionQuery   query.TransactionQueryInterface
		Signature          crypto.SignatureInterface
		MempoolService     MempoolServiceInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "wantTrue",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorCheckGenesisTrue{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			want: true,
		},
		{
			name: "wantFalse",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorCheckGenesisFalse{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:          tt.fields.Chaintype,
				QueryExecutor:      tt.fields.QueryExecutor,
				BlockQuery:         tt.fields.BlockQuery,
				MempoolQuery:       tt.fields.MempoolQuery,
				TransactionQuery:   tt.fields.TransactionQuery,
				Signature:          tt.fields.Signature,
				MempoolService:     tt.fields.MempoolService,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
			}
			if got := bs.CheckGenesis(); got != tt.want {
				t.Errorf("BlockService.CheckGenesis() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_CheckSignatureBlock(t *testing.T) {
	type fields struct {
		Chaintype           contract.ChainType
		QueryExecutor       query.ExecutorInterface
		BlockQuery          query.BlockQueryInterface
		MempoolQuery        query.MempoolQueryInterface
		TransactionQuery    query.TransactionQueryInterface
		Signature           crypto.SignatureInterface
		MempoolService      MempoolServiceInterface
		ActionTypeSwitcher  transaction.TypeActionSwitcher
		AccountBalanceQuery query.AccountBalanceQueryInterface
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "wantFalse::BlockSignatureNil",
			fields: fields{
				Chaintype:        &chaintype.MainChain{},
				QueryExecutor:    &mockQueryExecutorSuccess{},
				BlockQuery:       query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:     query.NewMempoolQuery(&chaintype.MainChain{}),
				TransactionQuery: query.NewTransactionQuery(&chaintype.MainChain{}),
				Signature:        &mockSignature{},
			},
			args: args{
				block: &model.Block{},
			},
			want: false,
		},
		{
			name: "wantFalse::GetAddressFiled",
			fields: fields{
				Chaintype:        &chaintype.MainChain{},
				QueryExecutor:    &mockQueryExecutorSuccess{},
				BlockQuery:       query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:     query.NewMempoolQuery(&chaintype.MainChain{}),
				TransactionQuery: query.NewTransactionQuery(&chaintype.MainChain{}),
				Signature:        &mockSignature{},
			},
			args: args{
				block: &model.Block{
					Version:           1,
					PreviousBlockHash: []byte{},
					BlockSeed:         []byte{},
					BlocksmithID:      nil,
					Timestamp:         15875392,
					TotalAmount:       0,
					TotalFee:          0,
					TotalCoinBase:     0,
					Transactions:      []*model.Transaction{},
					PayloadHash:       []byte{},
					PayloadLength:     0,
					BlockSignature:    []byte{},
				},
			},
			want: false,
		},
		{
			name: "wantFalse::GetAddressFiled",
			fields: fields{
				Chaintype:        &chaintype.MainChain{},
				QueryExecutor:    &mockQueryExecutorSuccess{},
				BlockQuery:       query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:     query.NewMempoolQuery(&chaintype.MainChain{}),
				TransactionQuery: query.NewTransactionQuery(&chaintype.MainChain{}),
				Signature:        &mockSignature{},
			},
			args: args{
				block: &model.Block{
					Version:           1,
					PreviousBlockHash: []byte{},
					BlockSeed:         []byte{},
					BlocksmithID:      nil,
					Timestamp:         15875392,
					TotalAmount:       0,
					TotalFee:          0,
					TotalCoinBase:     0,
					Transactions:      []*model.Transaction{},
					PayloadHash:       []byte{},
					PayloadLength:     0,
					BlockSignature:    []byte{},
				},
			},
			want: false,
		},
		{
			name: "wantTrue::VerifySignature",
			fields: fields{
				Chaintype:        &chaintype.MainChain{},
				QueryExecutor:    &mockQueryExecutorSuccess{},
				BlockQuery:       query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:     query.NewMempoolQuery(&chaintype.MainChain{}),
				TransactionQuery: query.NewTransactionQuery(&chaintype.MainChain{}),
				Signature:        &mockSignature{},
			},
			args: args{
				block: &model.Block{
					Version:           1,
					PreviousBlockHash: []byte{},
					BlockSeed:         []byte{},
					BlocksmithID:      make([]byte, 32),
					Timestamp:         15875392,
					TotalAmount:       0,
					TotalFee:          0,
					TotalCoinBase:     0,
					Transactions:      []*model.Transaction{},
					PayloadHash:       []byte{},
					PayloadLength:     0,
					BlockSignature:    []byte{},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:           tt.fields.Chaintype,
				QueryExecutor:       tt.fields.QueryExecutor,
				BlockQuery:          tt.fields.BlockQuery,
				MempoolQuery:        tt.fields.MempoolQuery,
				TransactionQuery:    tt.fields.TransactionQuery,
				Signature:           tt.fields.Signature,
				MempoolService:      tt.fields.MempoolService,
				ActionTypeSwitcher:  tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
			}
			if got := bs.CheckSignatureBlock(tt.args.block); got != tt.want {
				t.Errorf("BlockService.CheckSignatureBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_ReceivedBlockListener(t *testing.T) {
	type (
		fields struct {
			Chaintype           contract.ChainType
			QueryExecutor       query.ExecutorInterface
			BlockQuery          query.BlockQueryInterface
			MempoolQuery        query.MempoolQueryInterface
			TransactionQuery    query.TransactionQueryInterface
			Signature           crypto.SignatureInterface
			MempoolService      MempoolServiceInterface
			ActionTypeSwitcher  transaction.TypeActionSwitcher
			AccountBalanceQuery query.AccountBalanceQueryInterface
			Observer            *observer.Observer
		}
		args struct {
			block *model.Block
		}
	)
	tests := []struct {
		name   string
		fields fields
		args   args
		want   observer.Listener
	}{
		{
			name: "wantLastBlockAndPrevBlockNotEqual",
			fields: fields{
				Signature:     &mockSignature{},
				QueryExecutor: &mockQueryExecutorSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			args: args{
				&model.Block{
					ID:                   0,
					Height:               0,
					Version:              1,
					CumulativeDifficulty: "",
					SmithScale:           0,
					PreviousBlockHash:    []byte{},
					BlockSeed:            []byte{},
					BlocksmithID:         make([]byte, 32),
					Timestamp:            12345678,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					PayloadLength:        0,
					BlockSignature:       []byte{},
				},
			},
			want: observer.Listener{
				OnNotify: func(data interface{}, args interface{}) {

				},
			},
		},
		{
			name: "wantGetLasBlockFail",
			fields: fields{
				Signature:     &mockSignature{},
				QueryExecutor: &mockQueryExecuteNotNil{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			args: args{
				&model.Block{
					ID:                   0,
					Height:               0,
					Version:              1,
					CumulativeDifficulty: "",
					SmithScale:           0,
					PreviousBlockHash:    []byte{},
					BlockSeed:            []byte{},
					BlocksmithID:         make([]byte, 32),
					Timestamp:            12345678,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					PayloadLength:        0,
					BlockSignature:       []byte{},
				},
			},
			want: observer.Listener{
				OnNotify: func(data interface{}, args interface{}) {

				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:           tt.fields.Chaintype,
				QueryExecutor:       tt.fields.QueryExecutor,
				BlockQuery:          tt.fields.BlockQuery,
				MempoolQuery:        tt.fields.MempoolQuery,
				TransactionQuery:    tt.fields.TransactionQuery,
				Signature:           tt.fields.Signature,
				MempoolService:      tt.fields.MempoolService,
				ActionTypeSwitcher:  tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				Observer:            tt.fields.Observer,
			}

			got := bs.ReceivedBlockListener()
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("BlockService.ReceivedBlockListener() = %v, want %v", got, tt.want)
			}
			testOnNotify(got.OnNotify, tt.args.block)
		})
	}
}

func testOnNotify(fn observer.OnNotify, block *model.Block) {
	fn(block, nil)
}
