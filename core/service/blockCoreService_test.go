package service

import (
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	util2 "github.com/zoobc/zoobc-core/core/util"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	mockSignature struct {
		crypto.Signature
	}
	mockSignatureFail struct {
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
	mockQueryExecutorNotFound struct {
		query.Executor
	}
	mockTypeAction struct {
		transaction.SendMoney
	}
	mockTypeActionSuccess struct {
		mockTypeAction
	}

	mockKVExecutorSuccess struct {
		kvdb.KVExecutor
	}

	mockKVExecutorSuccessKeyNotFound struct {
		mockKVExecutorSuccess
	}

	mockKVExecutorFailOtherError struct {
		mockKVExecutorSuccess
	}
)

var mockBlock = &model.Block{
	ID:                   0,
	PreviousBlockHash:    nil,
	Height:               0,
	Timestamp:            0,
	BlockSeed:            nil,
	BlockSignature:       nil,
	CumulativeDifficulty: "",
	SmithScale:           0,
	BlocksmithPublicKey:  []byte{},
	TotalAmount:          0,
	TotalFee:             0,
	TotalCoinBase:        0,
	Version:              0,
	PayloadLength:        0,
	PayloadHash:          nil,
	Transactions:         nil,
}

func (*mockKVExecutorSuccess) Get(key string) ([]byte, error) {
	return nil, nil
}

func (*mockKVExecutorSuccess) Insert(key string, value []byte, expiry int) error {
	return nil
}

func (*mockKVExecutorSuccessKeyNotFound) Get(key string) ([]byte, error) {
	return nil, badger.ErrKeyNotFound
}

func (*mockKVExecutorFailOtherError) Get(key string) ([]byte, error) {
	return nil, badger.ErrInvalidKey
}

func (*mockKVExecutorFailOtherError) Insert(key string, value []byte, expiry int) error {
	return badger.ErrInvalidKey
}

var (
	bcsAddress1    = "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"
	bcsAddress2    = "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J"
	bcsAddress3    = "nK_ouxdDDwuJiogiDAi_zs1LqeN7f5ZsXbFtXGqGc0Pd"
	bcsNodePubKey1 = []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
	bcsNodePubKey2 = []byte{1, 2, 3, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
	bcsNodePubKey3 = []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
	bcsBlock1 = &model.Block{
		ID:                   0,
		PreviousBlockHash:    []byte{},
		Height:               1,
		Timestamp:            1562806389280,
		BlockSeed:            []byte{},
		BlockSignature:       []byte{},
		CumulativeDifficulty: string(100000000),
		SmithScale:           1,
		PayloadLength:        0,
		PayloadHash:          []byte{},
		BlocksmithPublicKey:  bcsNodePubKey1,
		TotalAmount:          100000000,
		TotalFee:             10000000,
		TotalCoinBase:        1,
		Version:              0,
	}
	mockTransaction = &model.Transaction{
		ID:                      1,
		BlockID:                 1,
		Height:                  0,
		SenderAccountAddress:    "BCZ",
		RecipientAccountAddress: "ZCB",
		TransactionType:         1,
		Fee:                     10,
		Timestamp:               1000,
		TransactionHash:         []byte{},
		TransactionBodyLength:   8,
		TransactionBodyBytes:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Signature:               []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Version:                 1,
		TransactionIndex:        1,
	}
)

// mockTypeAction
func (*mockTypeAction) ApplyConfirmed() error {
	return nil
}
func (*mockTypeAction) Validate(bool) error {
	return nil
}
func (*mockTypeAction) GetAmount() int64 {
	return 10
}
func (*mockTypeActionSuccess) GetTransactionType(tx *model.Transaction) (transaction.TypeAction, error) {
	return &mockTypeAction{}, nil
}

// mockSignature
func (*mockSignature) SignByNode(payload []byte, nodeSeed string) []byte {
	return []byte{}
}

func (*mockSignature) VerifyNodeSignature(
	payload, signature, nodePublicKey []byte,
) bool {
	return true
}

func (*mockSignatureFail) VerifyNodeSignature(
	payload, signature, nodePublicKey []byte,
) bool {
	return false
}

// mockQueryExecutorScanFail
func (*mockQueryExecutorScanFail) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{
		"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
		"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase"}))
	rows, _ := db.Query(qe)
	return rows, nil
}

// mockQueryExecutorNotFound
func (*mockQueryExecutorNotFound) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, queued, latest, height " +
		"FROM node_registry WHERE node_public_key = ? AND height <= ? ORDER BY height DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
			"Version"},
		))
	default:
		return nil, errors.New("mockQueryExecutorNotFound - InvalidQuery")
	}
	rows, _ := db.Query(qe)
	return rows, nil
}

// mockQueryExecutorNotNil
func (*mockQueryExecuteNotNil) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"ID"}))
	return db.Query("")
}

// mockQueryExecutorFail
func (*mockQueryExecutorFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}
func (*mockQueryExecutorFail) ExecuteStatement(qe string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("MockedError")
}
func (*mockQueryExecutorFail) BeginTx() error { return nil }

func (*mockQueryExecutorFail) RollbackTx() error { return nil }

func (*mockQueryExecutorFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("mockError:deleteMempoolFail")
}
func (*mockQueryExecutorFail) CommitTx() error { return errors.New("mockError:commitFail") }

// mockQueryExecutorSuccess
func (*mockQueryExecutorSuccess) BeginTx() error { return nil }

func (*mockQueryExecutorSuccess) RollbackTx() error { return nil }

func (*mockQueryExecutorSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}
func (*mockQueryExecutorSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}
func (*mockQueryExecutorSuccess) CommitTx() error { return nil }

func (*mockQueryExecutorSuccess) ExecuteSelectRow(qStr string, args ...interface{}) *sql.Row {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{
		"ID", "Tree", "Timestamp",
	}))
	row := db.QueryRow(qStr)
	return row
}

func (*mockQueryExecutorSuccess) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"queued, latest, height FROM node_registry WHERE id = ? AND latest=1":
		for idx, arg := range args {
			if idx == 0 {
				nodeID := fmt.Sprintf("%d", arg)
				switch nodeID {
				case "1":
					mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "node_address", "locked_balance", "queued", "latest", "height",
					}).AddRow(1, bcsNodePubKey1, bcsAddress1, 10, "10.10.10.1", 100000000, false, true, 100))
				case "2":
					mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "node_address", "locked_balance", "queued", "latest", "height",
					}).AddRow(2, bcsNodePubKey2, bcsAddress2, 20, "10.10.10.2", 100000000, false, true, 200))
				case "3":
					mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "node_address", "locked_balance", "queued", "latest", "height",
					}).AddRow(3, bcsNodePubKey3, bcsAddress3, 30, "10.10.10.3", 100000000, false, true, 300))
				}
			}
		}
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, queued, latest, height " +
		"FROM node_registry WHERE node_public_key = ? AND height <= ? ORDER BY height DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
			"account_address", "registration_height", "node_address", "locked_balance", "queued", "latest", "height",
		}).AddRow(1, bcsNodePubKey1, bcsAddress1, 10, "10.10.10.10", 100000000, true, true, 100))
	case "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version FROM main_block WHERE height = 0":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
			"Version"},
		).AddRow(1, []byte{}, 1, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, bcsNodePubKey1, 0, 0, 0, 1))
	case "SELECT A.node_id, A.score, A.latest, A.height FROM participation_score as A INNER JOIN node_registry as B " +
		"ON A.node_id = B.id WHERE B.node_public_key=? AND B.latest=1 AND B.queued=0 AND A.latest=1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"node_id",
			"score",
			"latest",
			"height",
		},
		).AddRow(-1, 100000, true, 0))
	case "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version FROM main_block WHERE id = 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
			"Version"},
		).AddRow(1, []byte{}, 1, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, bcsNodePubKey1, 0, 0, 0, 1))
	case "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version FROM main_block " +
		"WHERE HEIGHT >= 0 ORDER BY HEIGHT LIMIT 2":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
			"Version"},
		).AddRow(1, []byte{}, 1, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, bcsNodePubKey1, 0, 0, 0, 1).AddRow(
			2, []byte{}, 1, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, bcsNodePubKey1, 0, 0, 0, 1))
	case "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version FROM main_block ORDER BY " +
		"height DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
			"Version"},
		).AddRow(1, []byte{}, 1, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, bcsNodePubKey1, 0, 0, 0, 1))
	case "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version FROM main_block " +
		"WHERE height = 0 LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase", "Version"}).
			AddRow(1, []byte{}, 0, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, bcsNodePubKey1, 0, 0, 0, 1))
	case "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, smith_scale, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version FROM main_block " +
		"WHERE height >= 0 ORDER BY height ASC LIMIT 100":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase", "Version"}).
			AddRow(1, []byte{}, 0, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, bcsNodePubKey1, 0, 0, 0, 1))
	case "SELECT id, block_id, block_height, sender_account_address, recipient_account_address, transaction_type, fee, timestamp, " +
		"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version, " +
		"transaction_index FROM \"transaction\" WHERE block_id = ? ORDER BY transaction_index ASC":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "BlockID", "BlockHeight", "SenderAccountAddress", "RecipientAccountAddress", "TransactionType",
			"Fee", "Timestamp", "TransactionHash", "TransactionBodyLength", "TransactionBodyBytes", "Signature",
			"Version", "TransactionIndex"},
		).AddRow(
			mockTransaction.ID,
			mockTransaction.BlockID,
			mockTransaction.Height,
			mockTransaction.SenderAccountAddress,
			mockTransaction.RecipientAccountAddress,
			mockTransaction.TransactionType,
			mockTransaction.Fee,
			mockTransaction.Timestamp,
			mockTransaction.TransactionHash,
			mockTransaction.TransactionBodyLength,
			mockTransaction.TransactionBodyBytes,
			mockTransaction.Signature,
			mockTransaction.Version,
			mockTransaction.TransactionIndex))
	case "SELECT id, fee_per_byte, arrival_timestamp, transaction_bytes, sender_account_address, recipient_account_address " +
		"FROM mempool WHERE id = :id":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "FeePerByte", "ArrivalTimestamp", "TransactionBytes", "SenderAccountAddress", "RecipientAccountAddress",
		}))
	}
	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockQueryExecutorSuccess) ExecuteStatement(qe string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func TestNewBlockService(t *testing.T) {
	type args struct {
		ct                      chaintype.ChainType
		kvExecutor              kvdb.KVExecutorInterface
		queryExecutor           query.ExecutorInterface
		blockQuery              query.BlockQueryInterface
		mempoolQuery            query.MempoolQueryInterface
		transactionQuery        query.TransactionQueryInterface
		merkleTreeQuery         query.MerkleTreeQueryInterface
		publishedReceiptQuery   query.PublishedReceiptQueryInterface
		signature               crypto.SignatureInterface
		mempoolService          MempoolServiceInterface
		receiptService          ReceiptServiceInterface
		txTypeSwitcher          transaction.TypeActionSwitcher
		accountBalanceQuery     query.AccountBalanceQueryInterface
		participationScoreQuery query.ParticipationScoreQueryInterface
		nodeRegistrationQuery   query.NodeRegistrationQueryInterface
		obsr                    *observer.Observer
		sortedBlocksmiths       *[]model.Blocksmith
		logger                  *log.Logger
	}
	tests := []struct {
		name string
		args args
		want *BlockService
	}{
		{
			name: "wantSuccess",
			args: args{
				ct:                &chaintype.MainChain{},
				obsr:              observer.NewObserver(),
				sortedBlocksmiths: &([]model.Blocksmith{}),
			},
			want: &BlockService{
				Chaintype:         &chaintype.MainChain{},
				Observer:          observer.NewObserver(),
				SortedBlocksmiths: &([]model.Blocksmith{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockService(
				tt.args.ct,
				tt.args.kvExecutor,
				tt.args.queryExecutor,
				tt.args.blockQuery,
				tt.args.mempoolQuery,
				tt.args.transactionQuery,
				tt.args.merkleTreeQuery,
				tt.args.publishedReceiptQuery,
				tt.args.signature,
				tt.args.mempoolService,
				tt.args.receiptService,
				tt.args.txTypeSwitcher,
				tt.args.accountBalanceQuery,
				tt.args.participationScoreQuery,
				tt.args.nodeRegistrationQuery,
				tt.args.obsr,
				tt.args.sortedBlocksmiths,
				tt.args.logger,
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_NewBlock(t *testing.T) {
	type fields struct {
		Chaintype          chaintype.ChainType
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
		blockSmithPublicKey []byte
		hash                string
		previousBlockHeight uint32
		timestamp           int64
		totalAmount         int64
		totalFee            int64
		totalCoinBase       int64
		transactions        []*model.Transaction
		publishedReceipts   []*model.PublishedReceipt
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
				blockSmithPublicKey: bcsNodePubKey1,
				hash:                "hash",
				previousBlockHeight: 0,
				timestamp:           15875392,
				totalAmount:         0,
				totalFee:            0,
				totalCoinBase:       0,
				transactions:        []*model.Transaction{},
				publishedReceipts:   []*model.PublishedReceipt{},
				payloadHash:         []byte{},
				payloadLength:       0,
				secretPhrase:        "secretphrase",
			},
			want: &model.Block{
				Version:             1,
				PreviousBlockHash:   []byte{},
				BlockSeed:           []byte{},
				BlocksmithPublicKey: bcsNodePubKey1,
				Timestamp:           15875392,
				TotalAmount:         0,
				TotalFee:            0,
				TotalCoinBase:       0,
				Transactions:        []*model.Transaction{},
				PublishedReceipts:   []*model.PublishedReceipt{},
				PayloadHash:         []byte{},
				PayloadLength:       0,
				BlockSignature:      []byte{},
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
				tt.args.blockSmithPublicKey,
				tt.args.hash,
				tt.args.previousBlockHeight,
				tt.args.timestamp,
				tt.args.totalAmount,
				tt.args.totalFee,
				tt.args.totalCoinBase,
				tt.args.transactions,
				tt.args.publishedReceipts,
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
		Chaintype          chaintype.ChainType
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
		blockSmithPublicKey  []byte
		hash                 string
		previousBlockHeight  uint32
		timestamp            int64
		totalAmount          int64
		totalFee             int64
		totalCoinBase        int64
		transactions         []*model.Transaction
		publishedReceipts    []*model.PublishedReceipt
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
				blockSmithPublicKey:  bcsNodePubKey1,
				hash:                 "hash",
				previousBlockHeight:  0,
				timestamp:            15875392,
				totalAmount:          0,
				totalFee:             0,
				totalCoinBase:        0,
				transactions:         []*model.Transaction{},
				publishedReceipts:    []*model.PublishedReceipt{},
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
				BlocksmithPublicKey:  bcsNodePubKey1,
				Timestamp:            15875392,
				TotalAmount:          0,
				TotalFee:             0,
				TotalCoinBase:        0,
				Transactions:         []*model.Transaction{},
				PublishedReceipts:    []*model.PublishedReceipt{},
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
				tt.args.blockSmithPublicKey,
				tt.args.hash,
				tt.args.previousBlockHeight,
				tt.args.timestamp,
				tt.args.totalAmount,
				tt.args.totalFee,
				tt.args.totalCoinBase,
				tt.args.transactions,
				tt.args.publishedReceipts,
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
		Chaintype          chaintype.ChainType
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
				timestamp: 3601,
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
	var mockBlocksmiths = &[]model.Blocksmith{
		{
			NodeID:    2,
			NodeOrder: new(big.Int).SetInt64(1000),
		},
	}

	type fields struct {
		Chaintype             chaintype.ChainType
		QueryExecutor         query.ExecutorInterface
		BlockQuery            query.BlockQueryInterface
		MempoolQuery          query.MempoolQueryInterface
		TransactionQuery      query.TransactionQueryInterface
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		Signature             crypto.SignatureInterface
		ActionTypeSwitcher    transaction.TypeActionSwitcher
		Observer              *observer.Observer
		SortedBlocksmiths     *[]model.Blocksmith
	}
	type args struct {
		previousBlock *model.Block
		block         *model.Block
		broadcast     bool
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
				Chaintype:             &chaintype.MainChain{},
				QueryExecutor:         &mockQueryExecutorSuccess{},
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Observer:              observer.NewObserver(),
				SortedBlocksmiths:     mockBlocksmiths,
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
					BlocksmithPublicKey:  bcsNodePubKey1,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					BlockSignature:       []byte{},
				},
				block: &model.Block{
					ID:                  1,
					Timestamp:           12000,
					Version:             1,
					PreviousBlockHash:   []byte{},
					BlockSeed:           []byte{},
					BlocksmithPublicKey: bcsNodePubKey1,
					TotalAmount:         0,
					TotalFee:            0,
					TotalCoinBase:       0,
					Transactions:        []*model.Transaction{},
					PayloadHash:         []byte{},
					BlockSignature:      []byte{},
				},
				broadcast: false,
			},
			wantErr: false,
		},
		{
			name: "PushBlock:Transactions<0 : broadcast true",
			fields: fields{
				Chaintype:             &chaintype.MainChain{},
				QueryExecutor:         &mockQueryExecutorSuccess{},
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Observer:              observer.NewObserver(),
				SortedBlocksmiths:     mockBlocksmiths,
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
					BlocksmithPublicKey:  bcsNodePubKey1,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					BlockSignature:       []byte{},
				},
				block: &model.Block{
					ID:                  1,
					Timestamp:           12000,
					Version:             1,
					PreviousBlockHash:   []byte{},
					BlockSeed:           []byte{},
					BlocksmithPublicKey: bcsNodePubKey1,
					TotalAmount:         0,
					TotalFee:            0,
					TotalCoinBase:       0,
					Transactions:        []*model.Transaction{},
					PayloadHash:         []byte{},
					BlockSignature:      []byte{},
				},
				broadcast: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:             tt.fields.Chaintype,
				QueryExecutor:         tt.fields.QueryExecutor,
				BlockQuery:            tt.fields.BlockQuery,
				MempoolQuery:          tt.fields.MempoolQuery,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				TransactionQuery:      tt.fields.TransactionQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				Signature:             tt.fields.Signature,
				ActionTypeSwitcher:    tt.fields.ActionTypeSwitcher,
				Observer:              tt.fields.Observer,
				SortedBlocksmiths:     tt.fields.SortedBlocksmiths,
			}
			if err := bs.PushBlock(tt.args.previousBlock, tt.args.block, false,
				tt.args.broadcast); (err != nil) != tt.wantErr {
				t.Errorf("BlockService.PushBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockService_GetLastBlock(t *testing.T) {
	type fields struct {
		Chaintype          chaintype.ChainType
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
				Chaintype:        &chaintype.MainChain{},
				QueryExecutor:    &mockQueryExecutorSuccess{},
				TransactionQuery: query.NewTransactionQuery(&chaintype.MainChain{}),
				BlockQuery:       query.NewBlockQuery(&chaintype.MainChain{}),
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
				BlocksmithPublicKey:  bcsNodePubKey1,
				TotalAmount:          0,
				Transactions: []*model.Transaction{
					mockTransaction,
				},
				TotalFee:      0,
				TotalCoinBase: 0,
				Version:       1,
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
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetLastBlock:SelectGotNil",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecuteNotNil{},
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
		Chaintype          chaintype.ChainType
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
				BlocksmithPublicKey:  bcsNodePubKey1,
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
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetGenesis:fail-{sql.rows.Next = false}", // genesis not found | rows.Next() -> false
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorScanFail{},
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
		Chaintype          chaintype.ChainType
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
					BlocksmithPublicKey:  bcsNodePubKey1,
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
		Chaintype     chaintype.ChainType
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		MempoolQuery  query.MempoolQueryInterface
		Signature     crypto.SignatureInterface
		Logger        *log.Logger
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
				Logger:        log.New(),
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
				Logger:        log.New(),
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
				Logger:        tt.fields.Logger,
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
	mockReceiptServiceReturnEmpty struct {
		ReceiptService
	}
)

func (*mockReceiptServiceReturnEmpty) SelectReceipts(
	blockTimestamp int64,
	numberOfReceipt int,
	lastBlockHeight uint32,
) ([]*model.PublishedReceipt, error) {
	return []*model.PublishedReceipt{}, nil
}

// mockQueryExecutorMempoolSuccess
func (*mockQueryExecutorMempoolSuccess) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
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

// mockMempoolServiceSelectSuccess
func (*mockMempoolServiceSelectSuccess) SelectTransactionsFromMempool(
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
		Chaintype          chaintype.ChainType
		QueryExecutor      query.ExecutorInterface
		BlockQuery         query.BlockQueryInterface
		MempoolQuery       query.MempoolQueryInterface
		TransactionQuery   query.TransactionQueryInterface
		Signature          crypto.SignatureInterface
		MempoolService     MempoolServiceInterface
		ReceiptService     ReceiptServiceInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
	}
	type args struct {
		previousBlock            *model.Block
		secretPhrase             string
		timestamp                int64
		blockSmithAccountAddress string
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
					Version:             1,
					PreviousBlockHash:   []byte{},
					BlockSeed:           []byte{},
					BlocksmithPublicKey: bcsNodePubKey1,
					Timestamp:           12344587645,
					TotalAmount:         0,
					TotalFee:            0,
					TotalCoinBase:       0,
					Transactions:        []*model.Transaction{},
					PayloadHash:         []byte{},
					PayloadLength:       0,
					BlockSignature:      []byte{},
				},
				secretPhrase:             "phasepress",
				timestamp:                12344587645,
				blockSmithAccountAddress: "BCZ",
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
					Version:             1,
					PreviousBlockHash:   []byte{},
					BlockSeed:           []byte{},
					BlocksmithPublicKey: bcsNodePubKey1,
					Timestamp:           12344587645,
					TotalAmount:         0,
					TotalFee:            0,
					TotalCoinBase:       0,
					Transactions:        []*model.Transaction{},
					PayloadHash:         []byte{},
					PayloadLength:       0,
					BlockSignature:      []byte{},
				},
				secretPhrase:             "pharsepress",
				timestamp:                12344587645,
				blockSmithAccountAddress: "BCZ",
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
				ReceiptService:     &mockReceiptServiceReturnEmpty{},
				ActionTypeSwitcher: &mockTypeActionSuccess{},
			},
			args: args{
				previousBlock: &model.Block{
					Version:             1,
					PreviousBlockHash:   []byte{},
					BlockSeed:           []byte{},
					BlocksmithPublicKey: bcsNodePubKey1,
					Timestamp:           12344587645,
					TotalAmount:         0,
					TotalFee:            0,
					TotalCoinBase:       0,
					Transactions:        []*model.Transaction{},
					PayloadHash:         []byte{},
					PayloadLength:       0,
					BlockSignature:      []byte{},
				},
				secretPhrase: "",
				timestamp:    12345678,
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
				ReceiptService:     tt.fields.ReceiptService,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
			}
			_, err := bs.GenerateBlock(
				tt.args.previousBlock,
				tt.args.secretPhrase,
				tt.args.timestamp,
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
		Chaintype           chaintype.ChainType
		QueryExecutor       query.ExecutorInterface
		BlockQuery          query.BlockQueryInterface
		MempoolQuery        query.MempoolQueryInterface
		TransactionQuery    query.TransactionQueryInterface
		AccountBalanceQuery query.AccountBalanceQueryInterface
		Signature           crypto.SignatureInterface
		MempoolService      MempoolServiceInterface
		ActionTypeSwitcher  transaction.TypeActionSwitcher
		Observer            *observer.Observer
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Chaintype:           &chaintype.MainChain{},
				Signature:           &mockSignature{},
				MempoolQuery:        query.NewMempoolQuery(&chaintype.MainChain{}),
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				MempoolService:      &mockMempoolServiceSelectFail{},
				ActionTypeSwitcher:  &mockTypeActionSuccess{},
				QueryExecutor:       &mockQueryExecutorSuccess{},
				BlockQuery:          query.NewBlockQuery(&chaintype.MainChain{}),
				TransactionQuery:    query.NewTransactionQuery(&chaintype.MainChain{}),
				Observer:            observer.NewObserver(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:           tt.fields.Chaintype,
				QueryExecutor:       tt.fields.QueryExecutor,
				BlockQuery:          tt.fields.BlockQuery,
				MempoolQuery:        tt.fields.MempoolQuery,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				TransactionQuery:    tt.fields.TransactionQuery,
				Signature:           tt.fields.Signature,
				MempoolService:      tt.fields.MempoolService,
				ActionTypeSwitcher:  tt.fields.ActionTypeSwitcher,
				Observer:            tt.fields.Observer,
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

func (*mockQueryExecutorCheckGenesisFalse) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
		"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
		"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
		"Version",
	}))
	return db.Query("")
}
func (*mockQueryExecutorCheckGenesisTrue) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
		"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
		"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
		"Version",
	}).AddRow((&chaintype.MainChain{}).GetGenesisBlockID(), []byte{}, 1, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1))
	return db.Query("")
}

func TestBlockService_CheckGenesis(t *testing.T) {
	type fields struct {
		Chaintype          chaintype.ChainType
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

func TestBlockService_GetBlockByHeight(t *testing.T) {
	type fields struct {
		Chaintype           chaintype.ChainType
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
	type args struct {
		height uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Block
		wantErr bool
	}{
		{
			name: "GetBlockByHeight:Success", // All is good
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
				BlocksmithPublicKey:  bcsNodePubKey1,
				TotalAmount:          0,
				TotalFee:             0,
				TotalCoinBase:        0,
				Version:              1,
			},
			wantErr: false,
		},
		{
			name: "GetBlockByHeight:FailNoEntryFound", // All is good
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
			got, err := bs.GetBlockByHeight(tt.args.height)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetBlockByHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetBlockByHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_GetBlockByID(t *testing.T) {
	type fields struct {
		Chaintype           chaintype.ChainType
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
	type args struct {
		ID int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Block
		wantErr bool
	}{
		{
			name: "GetBlockByID:Success", // All is good
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			args: args{
				ID: int64(1),
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
				BlocksmithPublicKey:  bcsNodePubKey1,
				TotalAmount:          0,
				TotalFee:             0,
				TotalCoinBase:        0,
				Version:              1,
			},
			wantErr: false,
		},
		{
			name: "GetBlockByID:FailNoEntryFound", // All is good
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
			got, err := bs.GetBlockByID(tt.args.ID)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetBlockByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetBlockByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_GetBlocksFromHeight(t *testing.T) {
	type fields struct {
		Chaintype           chaintype.ChainType
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
	type args struct {
		startHeight, limit uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Block
		wantErr bool
	}{
		{
			name: "GetBlocksFromHeight:Success", // All is good
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			args: args{
				startHeight: 0,
				limit:       2,
			},
			want: []*model.Block{
				{
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
					BlocksmithPublicKey:  bcsNodePubKey1,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Version:              1,
				},
				{
					ID:                   2,
					PreviousBlockHash:    []byte{},
					Height:               1,
					Timestamp:            10000,
					BlockSeed:            []byte{},
					BlockSignature:       []byte{},
					CumulativeDifficulty: "",
					SmithScale:           1,
					PayloadLength:        2,
					PayloadHash:          []byte{},
					BlocksmithPublicKey:  bcsNodePubKey1,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Version:              1,
				},
			},
			wantErr: false,
		},
		{
			name: "GetBlocksFromHeight:FailNoEntryFound", // All is good
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
			got, err := bs.GetBlocksFromHeight(tt.args.startHeight, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetBlocksFromHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetBlocksFromHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_ReceiveBlock(t *testing.T) {
	type fields struct {
		Chaintype           chaintype.ChainType
		KVExecutor          kvdb.KVExecutorInterface
		QueryExecutor       query.ExecutorInterface
		BlockQuery          query.BlockQueryInterface
		MempoolQuery        query.MempoolQueryInterface
		TransactionQuery    query.TransactionQueryInterface
		MerkleTreeQuery     query.MerkleTreeQueryInterface
		Signature           crypto.SignatureInterface
		MempoolService      MempoolServiceInterface
		ActionTypeSwitcher  transaction.TypeActionSwitcher
		AccountBalanceQuery query.AccountBalanceQueryInterface
		Observer            *observer.Observer
		SortedBlocksmiths   *[]model.Blocksmith
	}
	type args struct {
		senderPublicKey  []byte
		lastBlock        *model.Block
		block            *model.Block
		nodeSecretPhrase string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.BatchReceipt
		wantErr bool
	}{
		{
			name: "ReceiveBlock:fail - {incoming block.previousBlockHash == nil}",
			args: args{
				senderPublicKey: nil,
				lastBlock:       nil,
				block: &model.Block{
					PreviousBlockHash: nil,
				},
				nodeSecretPhrase: "",
			},
			fields: fields{
				Chaintype:           nil,
				QueryExecutor:       nil,
				BlockQuery:          nil,
				MempoolQuery:        nil,
				TransactionQuery:    nil,
				Signature:           nil,
				MempoolService:      nil,
				ActionTypeSwitcher:  nil,
				AccountBalanceQuery: nil,
				Observer:            nil,
				SortedBlocksmiths:   nil,
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "ReceiveBlock:fail - {incoming block signature == nil}",
			args: args{
				senderPublicKey: nil,
				lastBlock:       nil,
				block: &model.Block{
					PreviousBlockHash: nil,
					BlockSignature:    nil,
				},
				nodeSecretPhrase: "",
			},
			fields: fields{
				Chaintype:           nil,
				QueryExecutor:       nil,
				BlockQuery:          nil,
				MempoolQuery:        nil,
				TransactionQuery:    nil,
				Signature:           nil,
				MempoolService:      nil,
				ActionTypeSwitcher:  nil,
				AccountBalanceQuery: nil,
				Observer:            nil,
				SortedBlocksmiths:   nil,
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "ReceiveBlock:fail - {signature validation fail}",
			args: args{
				senderPublicKey:  bcsNodePubKey1,
				block:            bcsBlock1,
				lastBlock:        nil,
				nodeSecretPhrase: "",
			},
			fields: fields{
				Chaintype:           nil,
				QueryExecutor:       nil,
				BlockQuery:          nil,
				MempoolQuery:        nil,
				TransactionQuery:    nil,
				Signature:           &mockSignatureFail{},
				MempoolService:      nil,
				ActionTypeSwitcher:  nil,
				AccountBalanceQuery: nil,
				Observer:            nil,
				SortedBlocksmiths:   nil,
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "ReceiveBlock:fail - {get last block byte error : no signature}",
			args: args{
				senderPublicKey: nil,
				lastBlock: &model.Block{
					BlockSignature: nil,
				},
				block: &model.Block{
					PreviousBlockHash: []byte{},
					BlockSignature:    nil,
				},
				nodeSecretPhrase: "",
			},
			fields: fields{
				Chaintype:           nil,
				QueryExecutor:       nil,
				BlockQuery:          nil,
				MempoolQuery:        nil,
				TransactionQuery:    nil,
				Signature:           &mockSignature{},
				MempoolService:      nil,
				ActionTypeSwitcher:  nil,
				AccountBalanceQuery: nil,
				Observer:            nil,
				SortedBlocksmiths:   nil,
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "ReceiveBlock:fail - {last block hash != previousBlockHash}",
			args: args{
				senderPublicKey: nil,
				lastBlock: &model.Block{
					BlockSignature: []byte{},
				},
				block: &model.Block{
					PreviousBlockHash: []byte{},
					BlockSignature:    nil,
				},
				nodeSecretPhrase: "",
			},
			fields: fields{
				Chaintype:           nil,
				KVExecutor:          &mockKVExecutorSuccess{},
				QueryExecutor:       nil,
				BlockQuery:          nil,
				MempoolQuery:        nil,
				TransactionQuery:    nil,
				Signature:           &mockSignature{},
				MempoolService:      nil,
				ActionTypeSwitcher:  nil,
				AccountBalanceQuery: nil,
				Observer:            nil,
				SortedBlocksmiths:   nil,
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "ReceiveBlock:fail - {last block hash != previousBlockHash - kvExecutor KeyNotFound - generate batch receipt success}",
			args: args{
				senderPublicKey: []byte{1, 3, 4, 5, 6},
				lastBlock: &model.Block{
					BlockSignature: []byte{},
				},
				block: &model.Block{
					PreviousBlockHash: []byte{133, 198, 93, 19, 200, 113, 155, 159, 136, 63, 230, 29, 21, 173, 160, 40,
						169, 25, 61, 85, 203, 79, 43, 182, 5, 236, 141, 124, 46, 193, 223, 255, 0},
					BlockSignature:      nil,
					SmithScale:          1,
					BlocksmithPublicKey: []byte{1, 3, 4, 5, 6},
				},
				nodeSecretPhrase: "",
			},
			fields: fields{
				Chaintype:           nil,
				KVExecutor:          &mockKVExecutorSuccessKeyNotFound{},
				QueryExecutor:       &mockQueryExecutorSuccess{},
				BlockQuery:          nil,
				MempoolQuery:        nil,
				MerkleTreeQuery:     query.NewMerkleTreeQuery(),
				TransactionQuery:    nil,
				Signature:           &mockSignature{},
				MempoolService:      nil,
				ActionTypeSwitcher:  nil,
				AccountBalanceQuery: nil,
				Observer:            nil,
				SortedBlocksmiths:   nil,
			},
			wantErr: false,
			want: &model.BatchReceipt{
				SenderPublicKey: []byte{1, 3, 4, 5, 6},
				RecipientPublicKey: []byte{
					88, 220, 21, 76, 132, 107, 209, 213, 213, 206, 112, 50, 201, 183, 134, 250, 90, 163, 91, 63, 176,
					223, 177, 77, 197, 161, 178, 55, 31, 225, 233, 115,
				},
				DatumType: constant.ReceiptDatumTypeBlock,
				DatumHash: []byte{
					167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160, 97, 214, 98, 245, 128, 255, 77, 228,
					59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74,
				},
				ReferenceBlockHeight: 0,
				ReferenceBlockHash: []byte{133, 198, 93, 19, 200, 113, 155, 159, 136, 63, 230, 29, 21, 173, 160, 40,
					169, 25, 61, 85, 203, 79, 43, 182, 5, 236, 141, 124, 46, 193, 223, 255},
				RMRLinked:          nil,
				RecipientSignature: []byte{},
			},
		},
		{
			name: "ReceiveBlock:fail - {last block hash != previousBlockHash - kvExecutor other error - generate batch receipt success}",
			args: args{
				senderPublicKey: []byte{1, 3, 4, 5, 6},
				lastBlock: &model.Block{
					BlockSignature: []byte{},
				},
				block: &model.Block{
					PreviousBlockHash: []byte{133, 198, 93, 19, 200, 113, 155, 159, 136, 63, 230, 29, 21, 173, 160, 40,
						169, 25, 61, 85, 203, 79, 43, 182, 5, 236, 141, 124, 46, 193, 223, 255, 0},
					BlockSignature:      nil,
					SmithScale:          1,
					BlocksmithPublicKey: []byte{1, 3, 4, 5, 6},
				},
				nodeSecretPhrase: "",
			},
			fields: fields{
				Chaintype:           nil,
				KVExecutor:          &mockKVExecutorFailOtherError{},
				QueryExecutor:       &mockQueryExecutorSuccess{},
				BlockQuery:          nil,
				MempoolQuery:        nil,
				TransactionQuery:    nil,
				Signature:           &mockSignature{},
				MempoolService:      nil,
				ActionTypeSwitcher:  nil,
				AccountBalanceQuery: nil,
				Observer:            nil,
				SortedBlocksmiths:   nil,
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "ReceiveBlock:pushBlockFail",
			args: args{
				senderPublicKey: []byte{1, 3, 4, 5, 6},
				lastBlock: &model.Block{
					BlockSignature:       []byte{},
					CumulativeDifficulty: "123",
					SmithScale:           123,
				},
				block: &model.Block{
					BlocksmithPublicKey: []byte{1, 3, 4, 5, 6},
					PreviousBlockHash: []byte{
						133, 198, 93, 19, 200, 113, 155, 159, 136, 63, 230, 29, 21, 173, 160, 40,
						169, 25, 61, 85, 203, 79, 43, 182, 5, 236, 141, 124, 46, 193, 223, 255,
					},
					BlockSignature: nil,
					SmithScale:     1,
				},
				nodeSecretPhrase: "",
			},
			fields: fields{
				Chaintype:           &chaintype.MainChain{},
				QueryExecutor:       &mockQueryExecutorFail{},
				BlockQuery:          query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:        nil,
				TransactionQuery:    nil,
				Signature:           &mockSignature{},
				MempoolService:      nil,
				ActionTypeSwitcher:  nil,
				AccountBalanceQuery: nil,
				Observer:            observer.NewObserver(),
				SortedBlocksmiths: &[]model.Blocksmith{
					{
						NodePublicKey: []byte{1, 3, 4, 5, 6},
					},
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "ReceiveBlock:success",
			args: args{
				senderPublicKey: []byte{1, 3, 4, 5, 6},
				lastBlock: &model.Block{
					BlockSignature:       []byte{},
					CumulativeDifficulty: "123",
					SmithScale:           123,
				},
				block: &model.Block{
					PreviousBlockHash: []byte{133, 198, 93, 19, 200, 113, 155, 159, 136, 63, 230, 29, 21, 173, 160, 40,
						169, 25, 61, 85, 203, 79, 43, 182, 5, 236, 141, 124, 46, 193, 223, 255},
					BlockSignature:      nil,
					SmithScale:          1,
					BlocksmithPublicKey: []byte{1, 3, 4, 5, 6},
				},
				nodeSecretPhrase: "",
			},
			fields: fields{
				Chaintype:           &chaintype.MainChain{},
				KVExecutor:          &mockKVExecutorSuccess{},
				QueryExecutor:       &mockQueryExecutorSuccess{},
				BlockQuery:          query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:        nil,
				TransactionQuery:    nil,
				MerkleTreeQuery:     query.NewMerkleTreeQuery(),
				Signature:           &mockSignature{},
				MempoolService:      nil,
				ActionTypeSwitcher:  nil,
				AccountBalanceQuery: nil,
				Observer:            observer.NewObserver(),
				SortedBlocksmiths: &[]model.Blocksmith{
					{
						NodePublicKey: []byte{1, 3, 4, 5, 6},
					},
				},
			},
			wantErr: false,
			want: &model.BatchReceipt{
				SenderPublicKey: []byte{1, 3, 4, 5, 6},
				RecipientPublicKey: []byte{
					88, 220, 21, 76, 132, 107, 209, 213, 213, 206, 112, 50, 201, 183, 134, 250, 90, 163, 91, 63, 176,
					223, 177, 77, 197, 161, 178, 55, 31, 225, 233, 115,
				},
				DatumType: constant.ReceiptDatumTypeBlock,
				DatumHash: []byte{
					167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160, 97, 214, 98, 245, 128, 255, 77, 228,
					59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74,
				},
				ReferenceBlockHeight: 0,
				ReferenceBlockHash: []byte{133, 198, 93, 19, 200, 113, 155, 159, 136, 63, 230, 29, 21, 173, 160, 40,
					169, 25, 61, 85, 203, 79, 43, 182, 5, 236, 141, 124, 46, 193, 223, 255},
				RMRLinked:          nil,
				RecipientSignature: []byte{},
			},
		},
	}
	// test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:           tt.fields.Chaintype,
				KVExecutor:          tt.fields.KVExecutor,
				QueryExecutor:       tt.fields.QueryExecutor,
				BlockQuery:          tt.fields.BlockQuery,
				MempoolQuery:        tt.fields.MempoolQuery,
				TransactionQuery:    tt.fields.TransactionQuery,
				MerkleTreeQuery:     tt.fields.MerkleTreeQuery,
				Signature:           tt.fields.Signature,
				MempoolService:      tt.fields.MempoolService,
				ActionTypeSwitcher:  tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				Observer:            tt.fields.Observer,
				SortedBlocksmiths:   tt.fields.SortedBlocksmiths,
			}
			got, err := bs.ReceiveBlock(
				tt.args.senderPublicKey, tt.args.lastBlock, tt.args.block, tt.args.nodeSecretPhrase)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReceiveBlock() error = \n%v, wantErr \n%v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReceiveBlock() got = \n%v want \n%v", got, tt.want)
			}
		})
	}
}

func TestBlockService_GetBlockExtendedInfo(t *testing.T) {
	block := &model.Block{
		ID:                   999,
		PreviousBlockHash:    []byte{1, 1, 1, 1, 1, 1, 1, 1},
		Height:               1,
		Timestamp:            1562806389280,
		BlockSeed:            []byte{},
		BlockSignature:       []byte{},
		CumulativeDifficulty: string(100000000),
		SmithScale:           1,
		PayloadLength:        0,
		PayloadHash:          []byte{},
		BlocksmithPublicKey:  bcsNodePubKey1,
		TotalAmount:          100000000,
		TotalFee:             10000000,
		TotalCoinBase:        1,
		Version:              0,
	}
	genesisBlock := &model.Block{
		ID:                   999,
		PreviousBlockHash:    []byte{1, 1, 1, 1, 1, 1, 1, 1},
		Height:               0,
		Timestamp:            1562806389280,
		BlockSeed:            []byte{},
		BlockSignature:       []byte{},
		CumulativeDifficulty: string(100000000),
		SmithScale:           1,
		PayloadLength:        0,
		PayloadHash:          []byte{},
		BlocksmithPublicKey:  bcsNodePubKey1,
		TotalAmount:          100000000,
		TotalFee:             10000000,
		TotalCoinBase:        1,
		Version:              0,
	}
	type fields struct {
		Chaintype               chaintype.ChainType
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		Observer                *observer.Observer
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.BlockExtendedInfo
		wantErr bool
	}{
		{
			name: "GetBlockExtendedInfo:fail - {VersionedNodeRegistrationNotFound}",
			args: args{
				block: block,
			},
			fields: fields{
				QueryExecutor:         &mockQueryExecutorNotFound{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "GetBlockExtendedInfo:success-{genesisBlock}",
			args: args{
				block: genesisBlock,
			},
			fields: fields{
				QueryExecutor:         &mockQueryExecutorSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			wantErr: false,
			want: &model.BlockExtendedInfo{
				Block: &model.Block{
					ID:                   999,
					PreviousBlockHash:    []byte{1, 1, 1, 1, 1, 1, 1, 1},
					Height:               0,
					Timestamp:            1562806389280,
					BlockSeed:            []byte{},
					BlockSignature:       []byte{},
					CumulativeDifficulty: string(100000000),
					SmithScale:           1,
					PayloadLength:        0,
					PayloadHash:          []byte{},
					BlocksmithPublicKey:  bcsNodePubKey1,
					TotalAmount:          100000000,
					TotalFee:             10000000,
					TotalCoinBase:        1,
					Version:              0,
				},
				BlocksmithAccountAddress: constant.MainchainGenesisAccountAddress,
				TotalReceipts:            99,
				ReceiptValue:             99,
				PopChange:                -20,
			},
		},
		{
			name: "GetBlockExtendedInfo:success",
			args: args{
				block: block,
			},
			fields: fields{
				QueryExecutor:         &mockQueryExecutorSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			wantErr: false,
			want: &model.BlockExtendedInfo{
				Block: &model.Block{
					ID:                   999,
					PreviousBlockHash:    []byte{1, 1, 1, 1, 1, 1, 1, 1},
					Height:               1,
					Timestamp:            1562806389280,
					BlockSeed:            []byte{},
					BlockSignature:       []byte{},
					CumulativeDifficulty: string(100000000),
					SmithScale:           1,
					PayloadLength:        0,
					PayloadHash:          []byte{},
					BlocksmithPublicKey:  bcsNodePubKey1,
					TotalAmount:          100000000,
					TotalFee:             10000000,
					TotalCoinBase:        1,
					Version:              0,
				},
				BlocksmithAccountAddress: bcsAddress1,
				TotalReceipts:            99,
				ReceiptValue:             99,
				PopChange:                -20,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:               tt.fields.Chaintype,
				QueryExecutor:           tt.fields.QueryExecutor,
				BlockQuery:              tt.fields.BlockQuery,
				MempoolQuery:            tt.fields.MempoolQuery,
				TransactionQuery:        tt.fields.TransactionQuery,
				Signature:               tt.fields.Signature,
				MempoolService:          tt.fields.MempoolService,
				ActionTypeSwitcher:      tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				Observer:                tt.fields.Observer,
			}
			got, err := bs.GetBlockExtendedInfo(tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetBlockExtendedInfo() error = \n%v, wantErr \n%v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetBlockExtendedInfo() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestBlockService_RewardBlocksmithAccountAddresses(t *testing.T) {
	type fields struct {
		Chaintype               chaintype.ChainType
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		Observer                *observer.Observer
		SortedBlocksmiths       *[]model.Blocksmith
	}
	type args struct {
		blocksmithAccountAddresses []string
		totalReward                int64
		height                     uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "RewardBlocksmithAccountAddress:success",
			args: args{
				blocksmithAccountAddresses: []string{bcsAddress1},
				totalReward:                10000,
				height:                     1,
			},
			fields: fields{
				QueryExecutor:       &mockQueryExecutorSuccess{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:               tt.fields.Chaintype,
				QueryExecutor:           tt.fields.QueryExecutor,
				BlockQuery:              tt.fields.BlockQuery,
				MempoolQuery:            tt.fields.MempoolQuery,
				TransactionQuery:        tt.fields.TransactionQuery,
				Signature:               tt.fields.Signature,
				MempoolService:          tt.fields.MempoolService,
				ActionTypeSwitcher:      tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				Observer:                tt.fields.Observer,
				SortedBlocksmiths:       tt.fields.SortedBlocksmiths,
			}
			if err := bs.RewardBlocksmithAccountAddresses(tt.args.blocksmithAccountAddresses, tt.args.totalReward,
				tt.args.height); (err != nil) != tt.wantErr {
				t.Errorf("BlockService.RewardBlocksmithAccountAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockService_CoinbaseLotteryWinners(t *testing.T) {
	var mockBlocksmiths = &[]model.Blocksmith{
		{
			NodeID:    1,
			NodeOrder: new(big.Int).SetInt64(8000),
		},
		{
			NodeID:    2,
			NodeOrder: new(big.Int).SetInt64(1000),
		},
		{
			NodeID:    3,
			NodeOrder: new(big.Int).SetInt64(5000),
		},
	}

	type fields struct {
		Chaintype               chaintype.ChainType
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		Observer                *observer.Observer
		SortedBlocksmiths       *[]model.Blocksmith
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "CoinbaseLotteryWinners:success",
			fields: fields{
				QueryExecutor:         &mockQueryExecutorSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				SortedBlocksmiths:     mockBlocksmiths,
			},
			wantErr: false,
			want: []string{
				bcsAddress2,
				bcsAddress3,
				bcsAddress1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:               tt.fields.Chaintype,
				QueryExecutor:           tt.fields.QueryExecutor,
				BlockQuery:              tt.fields.BlockQuery,
				MempoolQuery:            tt.fields.MempoolQuery,
				TransactionQuery:        tt.fields.TransactionQuery,
				Signature:               tt.fields.Signature,
				MempoolService:          tt.fields.MempoolService,
				ActionTypeSwitcher:      tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				Observer:                tt.fields.Observer,
				SortedBlocksmiths:       tt.fields.SortedBlocksmiths,
			}
			got, err := bs.CoinbaseLotteryWinners()
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.CoinbaseLotteryWinners() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.CoinbaseLotteryWinners() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_GetBlocksmiths(t *testing.T) {
	type fields struct {
		Chaintype               chaintype.ChainType
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		Observer                *observer.Observer
		SortedBlocksmiths       *[]model.Blocksmith
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Blocksmith
		wantErr bool
	}{
		{
			name: "GetBlocksmiths:success",
			fields: fields{
				QueryExecutor:         &nrsMockQueryExecutorSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			args: args{
				block: nrsBlock2,
			},
			want: []*model.Blocksmith{
				{
					NodeID:        1,
					NodePublicKey: nrsNodePubKey1,
					SmithOrder:    util2.CalculateSmithOrder(new(big.Int).SetInt64(8000), new(big.Int).SetBytes(nrsBlock2.BlockSeed), 1),
					NodeOrder:     util2.CalculateNodeOrder(new(big.Int).SetInt64(8000), new(big.Int).SetBytes(nrsBlock2.BlockSeed), 1),
					BlockSeed:     new(big.Int).SetBytes(nrsBlock2.BlockSeed),
					Score:         new(big.Int).SetInt64(8000),
				},
			},
			wantErr: false,
		},
		{
			name: "GetBlocksmiths:fail-{ExecuteSelect}",
			fields: fields{
				QueryExecutor:         &nrsMockQueryExecutorFailActiveNodeRegistrations{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			args: args{
				block: nrsBlock2,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:               tt.fields.Chaintype,
				QueryExecutor:           tt.fields.QueryExecutor,
				BlockQuery:              tt.fields.BlockQuery,
				MempoolQuery:            tt.fields.MempoolQuery,
				TransactionQuery:        tt.fields.TransactionQuery,
				Signature:               tt.fields.Signature,
				MempoolService:          tt.fields.MempoolService,
				ActionTypeSwitcher:      tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				Observer:                tt.fields.Observer,
				SortedBlocksmiths:       tt.fields.SortedBlocksmiths,
			}
			got, err := bs.GetBlocksmiths(tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetBlocksmiths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetBlocksmiths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_generateBlockReceipt(t *testing.T) {
	mockSecretPhrase := ""
	mockSenderPublicKey, mockCurrentBlockHash := make([]byte, 32), make([]byte, 32)
	mockReceiptKey, _ := util.GetReceiptKey(mockCurrentBlockHash, mockSenderPublicKey)

	mockNodePublicKey := util.GetPublicKeyFromSeed(mockSecretPhrase)
	mockSuccessBatchReceipt, _ := util.GenerateBatchReceipt(
		mockBlock,
		mockSenderPublicKey,
		mockNodePublicKey,
		mockCurrentBlockHash,
		nil,
		constant.ReceiptDatumTypeBlock,
	)
	mockSuccessBatchReceipt.RecipientSignature = (&crypto.Signature{}).SignByNode(
		util.GetUnsignedBatchReceiptBytes(mockSuccessBatchReceipt),
		mockSecretPhrase,
	)
	type fields struct {
		Chaintype               chaintype.ChainType
		KVExecutor              kvdb.KVExecutorInterface
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		MerkleTreeQuery         query.MerkleTreeQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		Observer                *observer.Observer
		SortedBlocksmiths       *[]model.Blocksmith
	}
	type args struct {
		currentBlockHash []byte
		previousBlock    *model.Block
		senderPublicKey  []byte
		receiptKey       []byte
		nodeSecretPhrase string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.BatchReceipt
		wantErr bool
	}{
		{
			name: "generateBlockReceipt:kvDBInsertFail",
			fields: fields{
				Chaintype:               nil,
				KVExecutor:              &mockKVExecutorFailOtherError{},
				QueryExecutor:           &mockQueryExecutorSuccess{},
				BlockQuery:              nil,
				MempoolQuery:            nil,
				TransactionQuery:        nil,
				MerkleTreeQuery:         query.NewMerkleTreeQuery(),
				Signature:               &crypto.Signature{},
				MempoolService:          nil,
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				ParticipationScoreQuery: nil,
				NodeRegistrationQuery:   nil,
				Observer:                nil,
				SortedBlocksmiths:       nil,
			},
			args: args{
				currentBlockHash: mockCurrentBlockHash,
				previousBlock:    mockBlock,
				senderPublicKey:  mockSenderPublicKey,
				receiptKey:       mockReceiptKey,
				nodeSecretPhrase: mockSecretPhrase,
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "generateBlockReceipt:success",
			fields: fields{
				Chaintype:               nil,
				KVExecutor:              &mockKVExecutorSuccess{},
				QueryExecutor:           &mockQueryExecutorSuccess{},
				BlockQuery:              nil,
				MempoolQuery:            nil,
				TransactionQuery:        nil,
				MerkleTreeQuery:         query.NewMerkleTreeQuery(),
				Signature:               &crypto.Signature{},
				MempoolService:          nil,
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				ParticipationScoreQuery: nil,
				NodeRegistrationQuery:   nil,
				Observer:                nil,
				SortedBlocksmiths:       nil,
			},
			args: args{
				currentBlockHash: mockCurrentBlockHash,
				previousBlock:    mockBlock,
				senderPublicKey:  mockSenderPublicKey,
				receiptKey:       mockReceiptKey,
				nodeSecretPhrase: mockSecretPhrase,
			},
			wantErr: false,
			want:    mockSuccessBatchReceipt,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:               tt.fields.Chaintype,
				KVExecutor:              tt.fields.KVExecutor,
				QueryExecutor:           tt.fields.QueryExecutor,
				BlockQuery:              tt.fields.BlockQuery,
				MempoolQuery:            tt.fields.MempoolQuery,
				TransactionQuery:        tt.fields.TransactionQuery,
				MerkleTreeQuery:         tt.fields.MerkleTreeQuery,
				Signature:               tt.fields.Signature,
				MempoolService:          tt.fields.MempoolService,
				ActionTypeSwitcher:      tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				Observer:                tt.fields.Observer,
				SortedBlocksmiths:       tt.fields.SortedBlocksmiths,
			}
			got, err := bs.generateBlockReceipt(tt.args.currentBlockHash, tt.args.previousBlock,
				tt.args.senderPublicKey, tt.args.receiptKey, tt.args.nodeSecretPhrase)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateBlockReceipt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateBlockReceipt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_GenerateGenesisBlock(t *testing.T) {
	type fields struct {
		Chaintype               chaintype.ChainType
		KVExecutor              kvdb.KVExecutorInterface
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		MerkleTreeQuery         query.MerkleTreeQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		Observer                *observer.Observer
		SortedBlocksmiths       *[]model.Blocksmith
		Logger                  *log.Logger
	}
	type args struct {
		genesisEntries []constant.MainchainGenesisConfigEntry
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "GenerateGenesisBlock:success",
			fields: fields{
				Chaintype:               &chaintype.MainChain{},
				KVExecutor:              nil,
				QueryExecutor:           nil,
				BlockQuery:              nil,
				MempoolQuery:            nil,
				TransactionQuery:        nil,
				MerkleTreeQuery:         nil,
				Signature:               nil,
				MempoolService:          nil,
				ActionTypeSwitcher:      &transaction.TypeSwitcher{},
				AccountBalanceQuery:     nil,
				ParticipationScoreQuery: nil,
				NodeRegistrationQuery:   nil,
				Observer:                nil,
				SortedBlocksmiths:       nil,
			},
			args: args{
				genesisEntries: []constant.MainchainGenesisConfigEntry{
					{
						AccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
						AccountBalance: 0,
						NodePublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118,
							97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
						NodeAddress:        "0.0.0.0",
						LockedBalance:      10000000000000,
						ParticipationScore: 1000000000,
					},
					{
						AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
						AccountBalance: 0,
						NodePublicKey: []byte{0, 14, 6, 218, 170, 54, 60, 50, 2, 66, 130, 119, 226, 235, 126, 203, 5, 12, 152,
							194, 170, 146, 43, 63, 224, 101, 127, 241, 62, 152, 187, 255},
						NodeAddress:        "0.0.0.0",
						LockedBalance:      0,
						ParticipationScore: 1000000000,
					},
					{
						AccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
						AccountBalance: 0,
						NodePublicKey: []byte{140, 115, 35, 51, 159, 22, 234, 192, 38, 104, 96, 24, 80, 70, 86, 211, 123, 72, 52,
							221, 97, 121, 59, 151, 158, 90, 167, 17, 110, 253, 122, 158},
						NodeAddress:        "0.0.0.0",
						LockedBalance:      0,
						ParticipationScore: 1000000000,
					},
					{
						AccountAddress: "nK_ouxdDDwuJiogiDAi_zs1LqeN7f5ZsXbFtXGqGc0Pd",
						AccountBalance: 100000000000,
						NodePublicKey: []byte{41, 235, 184, 214, 70, 23, 153, 89, 104, 41, 250, 248, 51, 7, 69, 89, 234, 181, 100,
							163, 45, 69, 152, 70, 52, 201, 147, 70, 6, 242, 52, 220},
						NodeAddress:        "0.0.0.0",
						LockedBalance:      0,
						ParticipationScore: 1000000000,
					},
				},
			},
			wantErr: false,
			want:    -1294179708803500770,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:               tt.fields.Chaintype,
				KVExecutor:              tt.fields.KVExecutor,
				QueryExecutor:           tt.fields.QueryExecutor,
				BlockQuery:              tt.fields.BlockQuery,
				MempoolQuery:            tt.fields.MempoolQuery,
				TransactionQuery:        tt.fields.TransactionQuery,
				MerkleTreeQuery:         tt.fields.MerkleTreeQuery,
				Signature:               tt.fields.Signature,
				MempoolService:          tt.fields.MempoolService,
				ActionTypeSwitcher:      tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				Observer:                tt.fields.Observer,
				SortedBlocksmiths:       tt.fields.SortedBlocksmiths,
				Logger:                  tt.fields.Logger,
			}
			got, err := bs.GenerateGenesisBlock(tt.args.genesisEntries)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GenerateGenesisBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.ID != tt.want {
				t.Errorf("BlockService.GenerateGenesisBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}
