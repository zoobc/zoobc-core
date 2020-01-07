package service

import (
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dgraph-io/badger"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/smith/strategy"
	"github.com/zoobc/zoobc-core/observer"
	"golang.org/x/crypto/sha3"
)

var (
	mockBlockData = model.Block{
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

	mockNodeRegistrationServiceSuccess struct {
		NodeRegistrationService
	}

	mockNodeRegistrationServiceFail struct {
		NodeRegistrationService
	}
)

func (*mockNodeRegistrationServiceSuccess) AddParticipationScore(
	nodeID, scoreDelta int64,
	height uint32,
	tx bool,
) (newScore int64, err error) {
	return 100000, nil
}

func (*mockNodeRegistrationServiceSuccess) SelectNodesToBeAdmitted(limit uint32) ([]*model.NodeRegistration, error) {
	return []*model.NodeRegistration{
		{
			AccountAddress: "TESTADMITTED",
		},
	}, nil
}

func (*mockNodeRegistrationServiceSuccess) AdmitNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	return nil
}

func (*mockNodeRegistrationServiceSuccess) SelectNodesToBeExpelled() ([]*model.NodeRegistration, error) {
	return []*model.NodeRegistration{
		{
			AccountAddress: "TESTEXPELLED",
		},
	}, nil
}

func (*mockNodeRegistrationServiceFail) AddParticipationScore(
	nodeID, scoreDelta int64,
	height uint32,
	tx bool,
) (newScore int64, err error) {
	return 100000, nil
}

func (*mockNodeRegistrationServiceFail) SelectNodesToBeExpelled() ([]*model.NodeRegistration, error) {
	return []*model.NodeRegistration{
		{
			AccountAddress: "TESTEXPELLED",
		},
	}, nil
}
func (*mockNodeRegistrationServiceFail) ExpelNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	return nil
}
func (*mockNodeRegistrationServiceSuccess) GetNodeAdmittanceCycle() uint32 {
	return 1
}

func (*mockNodeRegistrationServiceSuccess) ExpelNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	return nil
}

func (*mockNodeRegistrationServiceSuccess) BuildScrambledNodes(block *model.Block) error {
	return nil
}

func (*mockNodeRegistrationServiceSuccess) GetBlockHeightToBuildScrambleNodes(lastBlockHeight uint32) uint32 {
	return lastBlockHeight
}

func (*mockNodeRegistrationServiceFail) BuildScrambledNodes(block *model.Block) error {
	return errors.New("mock Error")
}

func (*mockNodeRegistrationServiceFail) GetBlockHeightToBuildScrambleNodes(lastBlockHeight uint32) uint32 {
	return lastBlockHeight
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
		"PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase"}))
	rows, _ := db.Query(qe)
	return rows, nil
}

// mockQueryExecutorNotFound
func (*mockQueryExecutorNotFound) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"registration_status, latest, height  FROM node_registry WHERE node_public_key = ? AND height <= ? " +
		"ORDER BY height DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
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
func (*mockQueryExecutorFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mockRows := mock.NewRows([]string{"fake"})
	mockRows.AddRow("1")
	mock.ExpectQuery(qStr).WillReturnRows(mockRows)
	return db.QueryRow(qStr), nil
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

func (*mockQueryExecutorSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	switch qStr {
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "NodePublicKey", "AccountAddress", "RegistrationHeight", "NodeAddress", "LockedBalance", "RegistrationStatus",
			"Latest", "Height",
		}).AddRow(1, bcsNodePubKey1, bcsAddress1, 10, "10.10.10.1", 100000000, uint32(model.NodeRegistrationState_NodeQueued), true, 100))
	case "SELECT id, block_height, tree, timestamp FROM merkle_tree ORDER BY timestamp DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "BlockHeight", "Tree", "Timestamp",
		}))
	default:
		mockRows := mock.NewRows(query.NewBlockQuery(chaintype.GetChainType(0)).Fields)
		mockRows.AddRow(
			mockBlockData.GetID(),
			mockBlockData.GetBlockHash(),
			mockBlockData.GetPreviousBlockHash(),
			mockBlockData.GetHeight(),
			mockBlockData.GetTimestamp(),
			mockBlockData.GetBlockSeed(),
			mockBlockData.GetBlockSignature(),
			mockBlockData.GetCumulativeDifficulty(),
			mockBlockData.GetPayloadLength(),
			mockBlockData.GetPayloadHash(),
			mockBlockData.GetBlocksmithPublicKey(),
			mockBlockData.GetTotalAmount(),
			mockBlockData.GetTotalFee(),
			mockBlockData.GetTotalCoinBase(),
			mockBlockData.GetVersion(),
		)
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(mockRows)
	}
	row := db.QueryRow(qStr)
	return row, nil
}

func (*mockQueryExecutorSuccess) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE id = ? AND latest=1":
		for idx, arg := range args {
			if idx == 0 {
				nodeID := fmt.Sprintf("%d", arg)
				switch nodeID {
				case "1":
					mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "node_address", "locked_balance", "registration_status", "latest", "height",
					}).AddRow(1, bcsNodePubKey1, bcsAddress1, 10, "10.10.10.1", 100000000, uint32(model.NodeRegistrationState_NodeRegistered), true, 100))
				case "2":
					mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "node_address", "locked_balance", "registration_status", "latest", "height",
					}).AddRow(2, bcsNodePubKey2, bcsAddress2, 20, "10.10.10.2", 100000000, uint32(model.NodeRegistrationState_NodeRegistered), true, 200))
				case "3":
					mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "node_address", "locked_balance", "registration_status", "latest", "height",
					}).AddRow(3, bcsNodePubKey3, bcsAddress3, 30, "10.10.10.3", 100000000, uint32(model.NodeRegistrationState_NodeRegistered), true, 300))
				case "4":
					mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "node_address", "locked_balance", "registration_status", "latest", "height",
					}).AddRow(3, mockGoodBlock.BlocksmithPublicKey, bcsAddress3, 30, "10.10.10.3", 100000000,
						uint32(model.NodeRegistrationState_NodeRegistered), true, 300))
				}
			}
		}
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND height <= ? " +
		"ORDER BY height DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
			"account_address", "registration_height", "node_address", "locked_balance", "registration_status", "latest", "height",
		}).AddRow(1, bcsNodePubKey1, bcsAddress1, 10, "10.10.10.10", 100000000, uint32(model.NodeRegistrationState_NodeQueued), true, 100))
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version FROM main_block WHERE height = 0":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "BlockHash", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
			"Version"},
		).AddRow(1, []byte{}, []byte{}, 1, 10000, []byte{}, []byte{}, "", 2, []byte{}, bcsNodePubKey1, 0, 0, 0, 1))
	case "SELECT A.node_id, A.score, A.latest, A.height FROM participation_score as A INNER JOIN node_registry as B " +
		"ON A.node_id = B.id WHERE B.node_public_key=? AND B.latest=1 AND B.registration_status=0 AND A.latest=1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"node_id",
			"score",
			"latest",
			"height",
		},
		).AddRow(-1, 100000, true, 0))
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version FROM main_block ORDER BY " +
		"height DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).
			WillReturnRows(sqlmock.NewRows(
				query.NewBlockQuery(&chaintype.MainChain{}).Fields,
			).AddRow(
				mockBlockData.GetID(),
				mockBlockData.GetBlockHash(),
				mockBlockData.GetPreviousBlockHash(),
				mockBlockData.GetHeight(),
				mockBlockData.GetTimestamp(),
				mockBlockData.GetBlockSeed(),
				mockBlockData.GetBlockSignature(),
				mockBlockData.GetCumulativeDifficulty(),
				mockBlockData.GetPayloadLength(),
				mockBlockData.GetPayloadHash(),
				mockBlockData.GetBlocksmithPublicKey(),
				mockBlockData.GetTotalAmount(),
				mockBlockData.GetTotalFee(),
				mockBlockData.GetTotalCoinBase(),
				mockBlockData.GetVersion(),
			))
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
	case "SELECT id, block_height, fee_per_byte, arrival_timestamp, transaction_bytes, sender_account_address, recipient_account_address " +
		"FROM mempool WHERE id = :id":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "FeePerByte", "ArrivalTimestamp", "TransactionBytes", "SenderAccountAddress", "RecipientAccountAddress",
		}))
	case "SELECT id, fee_per_byte, arrival_timestamp, transaction_bytes, sender_account_address, recipient_account_address " +
		"FROM mempool WHERE id = :id":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "FeePerByte", "ArrivalTimestamp", "TransactionBytes", "SenderAccountAddress", "RecipientAccountAddress",
		}))
	case "SELECT nr.id AS nodeID, nr.node_public_key AS node_public_key, ps.score AS participation_score FROM node_registry " +
		"AS nr INNER JOIN participation_score AS ps ON nr.id = ps.node_id WHERE nr.registration_status = 0 AND nr.latest " +
		"= 1 AND ps.score > 0 AND ps.latest = 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"node_id", "node_public_key", "score",
		}).AddRow(
			mockBlocksmiths[0].NodeID,
			mockBlocksmiths[0].NodePublicKey,
			"1000",
		).AddRow(
			mockBlocksmiths[1].NodeID,
			mockBlocksmiths[1].NodePublicKey,
			"1000",
		))
	case "SELECT blocksmith_public_key, pop_change, block_height, blocksmith_index FROM skipped_blocksmith WHERE block_height = 0":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"blocksmith_public_key", "pop_change", "block_height", "blocksmith_index",
		}).AddRow(
			mockBlocksmiths[0].NodePublicKey,
			5000,
			mockPublishedReceipt[0].BlockHeight,
			0,
		))
	case "SELECT blocksmith_public_key, pop_change, block_height, blocksmith_index FROM skipped_blocksmith WHERE block_height = 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"blocksmith_public_key", "pop_change", "block_height", "blocksmith_index",
		}).AddRow(
			mockBlocksmiths[0].NodePublicKey,
			5000,
			mockPublishedReceipt[0].BlockHeight,
			0,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
		"reference_block_hash, rmr_linked, recipient_signature, intermediate_hashes, block_height, receipt_index, " +
		"published_index FROM published_receipt WHERE block_height = ? ORDER BY published_index ASC":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"sender_public_key", "recipient_public_key", "datum_type", "datum_hash", "reference_block_height",
			"reference_block_hash", "rmr_linked", "recipient_signature", "intermediate_hashes", "block_height",
			"receipt_index", "published_index",
		}).AddRow(
			mockPublishedReceipt[0].BatchReceipt.SenderPublicKey,
			mockPublishedReceipt[0].BatchReceipt.RecipientPublicKey,
			mockPublishedReceipt[0].BatchReceipt.DatumType,
			mockPublishedReceipt[0].BatchReceipt.DatumHash,
			mockPublishedReceipt[0].BatchReceipt.ReferenceBlockHeight,
			mockPublishedReceipt[0].BatchReceipt.ReferenceBlockHash,
			mockPublishedReceipt[0].BatchReceipt.RMRLinked,
			mockPublishedReceipt[0].BatchReceipt.RecipientSignature,
			mockPublishedReceipt[0].IntermediateHashes,
			mockPublishedReceipt[0].BlockHeight,
			mockPublishedReceipt[0].ReceiptIndex,
			mockPublishedReceipt[0].PublishedIndex,
		))
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"registration_status, latest, height, max(height) AS max_height FROM node_registry where height <= 0 AND " +
		"registration_status = 0 GROUP BY id ORDER BY height DESC":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id", "node_public_key", "account_address", "registration_height", "node_address", "locked_balance",
			"registration_status", "latest", "height",
		}))
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"registration_status, latest, height, max(height) AS max_height FROM node_registry where height <= 1 " +
		"AND registration_status = 0 GROUP BY id ORDER BY height DESC":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id", "node_public_key", "account_address", "registration_height", "node_address", "locked_balance",
			"registration_status", "latest", "height",
		}))
	}
	rows, _ := db.Query(qe)
	return rows, nil
}

var mockPublishedReceipt = []*model.PublishedReceipt{
	{
		BatchReceipt: &model.BatchReceipt{
			SenderPublicKey:      make([]byte, 32),
			RecipientPublicKey:   make([]byte, 32),
			DatumType:            0,
			DatumHash:            make([]byte, 32),
			ReferenceBlockHeight: 0,
			ReferenceBlockHash:   make([]byte, 32),
			RMRLinked:            nil,
			RecipientSignature:   make([]byte, 64),
		},
		IntermediateHashes: nil,
		BlockHeight:        1,
		ReceiptIndex:       0,
		PublishedIndex:     0,
	},
}

func (*mockQueryExecutorSuccess) ExecuteStatement(qe string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func TestBlockService_NewBlock(t *testing.T) {
	var (
		mockBlock = &model.Block{
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
		}
		mockBlockHash, _ = util.GetBlockHash(mockBlock, &chaintype.MainChain{})
	)
	mockBlock.BlockHash = mockBlockHash

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
		name    string
		fields  fields
		args    args
		want    *model.Block
		wantErr bool
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
			want: mockBlock,
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
			got, err := bs.NewMainBlock(
				tt.args.version,
				tt.args.previousBlockHash,
				tt.args.blockSeed,
				tt.args.blockSmithPublicKey,
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
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.NewBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
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
		previousBlockHeight  uint32
		timestamp            int64
		totalAmount          int64
		totalFee             int64
		totalCoinBase        int64
		transactions         []*model.Transaction
		publishedReceipts    []*model.PublishedReceipt
		spinePublicKeys      []*model.SpinePublicKey
		payloadHash          []byte
		payloadLength        uint32
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
				previousBlockHeight:  0,
				timestamp:            15875392,
				totalAmount:          0,
				totalFee:             0,
				totalCoinBase:        0,
				transactions:         []*model.Transaction{},
				publishedReceipts:    []*model.PublishedReceipt{},
				spinePublicKeys:      []*model.SpinePublicKey{},
				payloadHash:          []byte{},
				payloadLength:        8,
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
				SpinePublicKeys:      []*model.SpinePublicKey{},
				PayloadHash:          []byte{},
				PayloadLength:        8,
				CumulativeDifficulty: "1",
				BlockSignature:       []byte{},
				BlockHash: []byte{63, 27, 59, 75, 41, 129, 80, 143, 205, 223, 247, 196, 135, 45, 143, 255,
					124, 196, 95, 34, 28, 147, 250, 90, 168, 247, 163, 53, 101, 53, 245, 163},
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
			if got, _ := bs.NewGenesisBlock(
				tt.args.version,
				tt.args.previousBlockHash,
				tt.args.blockSeed,
				tt.args.blockSmithPublicKey,
				tt.args.previousBlockHeight,
				tt.args.timestamp,
				tt.args.totalAmount,
				tt.args.totalFee,
				tt.args.totalCoinBase,
				tt.args.transactions,
				tt.args.publishedReceipts,
				tt.args.spinePublicKeys,
				tt.args.payloadHash,
				tt.args.payloadLength,
				tt.args.cumulativeDifficulty,
				tt.args.genesisSignature,
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.NewGenesisBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	mockBlocksmiths = []*model.Blocksmith{
		{
			NodePublicKey: bcsNodePubKey1,
			NodeID:        2,
			NodeOrder:     new(big.Int).SetInt64(1000),
			Score:         new(big.Int).SetInt64(1000),
		},
		{
			NodePublicKey: bcsNodePubKey2,
			NodeID:        3,
			NodeOrder:     new(big.Int).SetInt64(2000),
			Score:         new(big.Int).SetInt64(2000),
		},
		{
			NodePublicKey: mockBlockData.BlocksmithPublicKey,
			NodeID:        4,
			NodeOrder:     new(big.Int).SetInt64(3000),
			Score:         new(big.Int).SetInt64(3000),
		},
	}
)

type (
	mockBlocksmithServicePushBlock struct {
		strategy.BlocksmithStrategyMain
	}
)

func (*mockBlocksmithServicePushBlock) GetSortedBlocksmiths(*model.Block) []*model.Blocksmith {
	return mockBlocksmiths
}
func (*mockBlocksmithServicePushBlock) GetSortedBlocksmithsMap(*model.Block) map[string]*int64 {
	var result = make(map[string]*int64)
	for index, mock := range mockBlocksmiths {
		mockIndex := int64(index)
		result[string(mock.NodePublicKey)] = &mockIndex
	}
	return result
}
func (*mockBlocksmithServicePushBlock) SortBlocksmiths(block *model.Block) {
}
func TestBlockService_PushBlock(t *testing.T) {
	type fields struct {
		Chaintype               chaintype.ChainType
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		AccountLedgerQuery      query.AccountLedgerQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		Signature               crypto.SignatureInterface
		SkippedBlocksmithQuery  query.SkippedBlocksmithQueryInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		Observer                *observer.Observer
		NodeRegistrationService NodeRegistrationServiceInterface
		BlocksmithStrategy      strategy.BlocksmithStrategyInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
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
				Chaintype:               &chaintype.MainChain{},
				QueryExecutor:           &mockQueryExecutorSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				AccountLedgerQuery:      query.NewAccountLedgerQuery(),
				Observer:                observer.NewObserver(),
				MempoolQuery:            query.NewMempoolQuery(&chaintype.MainChain{}),
				SkippedBlocksmithQuery:  query.NewSkippedBlocksmithQuery(),
				NodeRegistrationService: &mockNodeRegistrationServiceSuccess{},
				BlocksmithStrategy:      &mockBlocksmithServicePushBlock{},
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
			},
			args: args{
				previousBlock: &model.Block{
					ID:                   0,
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
				Chaintype:               &chaintype.MainChain{},
				QueryExecutor:           &mockQueryExecutorSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				NodeRegistrationService: &mockNodeRegistrationServiceSuccess{},
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				AccountLedgerQuery:      query.NewAccountLedgerQuery(),
				MempoolQuery:            query.NewMempoolQuery(&chaintype.MainChain{}),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				SkippedBlocksmithQuery:  query.NewSkippedBlocksmithQuery(),
				Observer:                observer.NewObserver(),
				BlocksmithStrategy:      &mockBlocksmithServicePushBlock{},
			},
			args: args{
				previousBlock: &model.Block{
					ID:                   0,
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
		{
			name: "PushBlock_FAIL:BuildScrambledNodes_Fails",
			fields: fields{
				Chaintype:               &chaintype.MainChain{},
				QueryExecutor:           &mockQueryExecutorSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				NodeRegistrationService: &mockNodeRegistrationServiceFail{},
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				AccountLedgerQuery:      query.NewAccountLedgerQuery(),
				MempoolQuery:            query.NewMempoolQuery(&chaintype.MainChain{}),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				SkippedBlocksmithQuery:  query.NewSkippedBlocksmithQuery(),
				Observer:                observer.NewObserver(),
				BlocksmithStrategy:      &mockBlocksmithServicePushBlock{},
			},
			args: args{
				previousBlock: &model.Block{
					ID:                   0,
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
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				TransactionQuery:        tt.fields.TransactionQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				AccountLedgerQuery:      tt.fields.AccountLedgerQuery,
				SkippedBlocksmithQuery:  tt.fields.SkippedBlocksmithQuery,
				Signature:               tt.fields.Signature,
				ActionTypeSwitcher:      tt.fields.ActionTypeSwitcher,
				Observer:                tt.fields.Observer,
				Logger:                  logrus.New(),
				NodeRegistrationService: tt.fields.NodeRegistrationService,
				BlocksmithStrategy:      tt.fields.BlocksmithStrategy,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
			}
			if err := bs.PushBlock(tt.args.previousBlock, tt.args.block,
				tt.args.broadcast); (err != nil) != tt.wantErr {
				t.Errorf("BlockService.PushBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockService_GetLastBlock(t *testing.T) {
	var mockBlockGetLastBlock = mockBlockData
	mockBlockGetLastBlock.Transactions = []*model.Transaction{
		mockTransaction,
	}

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
			want:    &mockBlockGetLastBlock,
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
				t.Errorf("BlockService.GetLastBlock() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryExecutorGetGenesisBlockSuccess struct {
		query.Executor
	}

	mockQueryExecutorGetGenesisBlockFail struct {
		query.Executor
	}
)

func (*mockQueryExecutorGetGenesisBlockSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.MainChain{}).Fields,
		).AddRow(
			mockBlockData.GetID(),
			mockBlockData.GetBlockHash(),
			mockBlockData.GetPreviousBlockHash(),
			mockBlockData.GetHeight(),
			mockBlockData.GetTimestamp(),
			mockBlockData.GetBlockSeed(),
			mockBlockData.GetBlockSignature(),
			mockBlockData.GetCumulativeDifficulty(),
			mockBlockData.GetPayloadLength(),
			mockBlockData.GetPayloadHash(),
			mockBlockData.GetBlocksmithPublicKey(),
			mockBlockData.GetTotalAmount(),
			mockBlockData.GetTotalFee(),
			mockBlockData.GetTotalCoinBase(),
			mockBlockData.GetVersion(),
		))
	return db.QueryRow(qStr), nil
}

func (*mockQueryExecutorGetGenesisBlockFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, nil
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
				QueryExecutor: &mockQueryExecutorGetGenesisBlockSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			want:    &mockBlockData,
			wantErr: false,
		},
		{
			name: "GetGenesis:fail",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorGetGenesisBlockFail{},
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

type (
	mockQueryExecutorGetBlocksSuccess struct {
		query.Executor
	}

	mockQueryExecutorGetBlocksFail struct {
		query.Executor
	}
)

func (*mockQueryExecutorGetBlocksSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	mock.ExpectQuery(qStr).WillReturnRows(sqlmock.NewRows(
		query.NewBlockQuery(&chaintype.MainChain{}).Fields,
	).AddRow(
		mockBlockData.GetID(),
		mockBlockData.GetBlockHash(),
		mockBlockData.GetPreviousBlockHash(),
		mockBlockData.GetHeight(),
		mockBlockData.GetTimestamp(),
		mockBlockData.GetBlockSeed(),
		mockBlockData.GetBlockSignature(),
		mockBlockData.GetCumulativeDifficulty(),
		mockBlockData.GetPayloadLength(),
		mockBlockData.GetPayloadHash(),
		mockBlockData.GetBlocksmithPublicKey(),
		mockBlockData.GetTotalAmount(),
		mockBlockData.GetTotalFee(),
		mockBlockData.GetTotalCoinBase(),
		mockBlockData.GetVersion(),
	))
	return db.Query(qStr)
}

func (*mockQueryExecutorGetBlocksFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
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
				QueryExecutor: &mockQueryExecutorGetBlocksSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			want: []*model.Block{
				&mockBlockData,
			},
			wantErr: false,
		},
		{
			name: "GetBlocks:fail",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorGetBlocksFail{},
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
	numberOfReceipt, lastBlockHeight uint32,
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
func (*mockMempoolServiceSelectSuccess) SelectTransactionsFromMempool(blockTimestamp int64) ([]*model.Transaction, error) {
	txByte := getTestSignedMempoolTransaction(1, 1562893305).TransactionBytes
	txHash := sha3.Sum256(txByte)
	return []*model.Transaction{
		{
			ID:              1,
			TransactionHash: txHash[:],
		},
	}, nil
}

// mockMempoolServiceSelectFail
func (*mockMempoolServiceSelectFail) SelectTransactionsFromMempool(blockTimestamp int64) ([]*model.Transaction, error) {
	return nil, errors.New("want error on select")
}

// mockMempoolServiceSelectSuccess
func (*mockMempoolServiceSelectWrongTransactionBytes) SelectTransactionsFromMempool(
	blockTimestamp int64,
) ([]*model.Transaction, error) {
	return []*model.Transaction{
		{
			ID: 1,
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
		BlocksmithStrategy strategy.BlocksmithStrategyInterface
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
				BlocksmithStrategy: &mockBlocksmithServicePushBlock{},
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
				BlocksmithStrategy: tt.fields.BlocksmithStrategy,
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

type (
	mockAddGenesisExecutor struct {
		query.Executor
	}
)

func (*mockAddGenesisExecutor) BeginTx() error    { return nil }
func (*mockAddGenesisExecutor) RollbackTx() error { return nil }
func (*mockAddGenesisExecutor) CommitTx() error   { return nil }
func (*mockAddGenesisExecutor) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}
func (*mockAddGenesisExecutor) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
		sqlmock.NewRows(query.NewMempoolQuery(chaintype.GetChainType(0)).Fields),
	)
	return db.Query(qStr)
}

type (
	mockBlocksmithServiceAddGenesisSuccess struct {
		strategy.BlocksmithStrategyMain
	}
)

func (*mockBlocksmithServiceAddGenesisSuccess) SortBlocksmiths(block *model.Block) {

}

func TestBlockService_AddGenesis(t *testing.T) {
	type fields struct {
		Chaintype               chaintype.ChainType
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		Observer                *observer.Observer
		NodeRegistrationService NodeRegistrationServiceInterface
		BlocksmithStrategy      strategy.BlocksmithStrategyInterface
		Logger                  *logrus.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Chaintype:               &chaintype.MainChain{},
				Signature:               &mockSignature{},
				MempoolQuery:            query.NewMempoolQuery(&chaintype.MainChain{}),
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				MempoolService:          &mockMempoolServiceSelectFail{},
				ActionTypeSwitcher:      &mockTypeActionSuccess{},
				QueryExecutor:           &mockAddGenesisExecutor{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				TransactionQuery:        query.NewTransactionQuery(&chaintype.MainChain{}),
				Observer:                observer.NewObserver(),
				NodeRegistrationService: &mockNodeRegistrationServiceSuccess{},
				BlocksmithStrategy:      &mockBlocksmithServiceAddGenesisSuccess{},
				Logger:                  log.New(),
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
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				TransactionQuery:        tt.fields.TransactionQuery,
				Signature:               tt.fields.Signature,
				MempoolService:          tt.fields.MempoolService,
				ActionTypeSwitcher:      tt.fields.ActionTypeSwitcher,
				Observer:                tt.fields.Observer,
				NodeRegistrationService: tt.fields.NodeRegistrationService,
				BlocksmithStrategy:      tt.fields.BlocksmithStrategy,
				Logger:                  tt.fields.Logger,
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
		"PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
		"Version",
	}))
	return db.Query("")
}

func (*mockQueryExecutorCheckGenesisFalse) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, nil
}

func (*mockQueryExecutorCheckGenesisTrue) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(
		query.NewBlockQuery(&chaintype.MainChain{}).Fields,
	).AddRow(
		mockBlockData.GetID(),
		mockBlockData.GetBlockHash(),
		mockBlockData.GetPreviousBlockHash(),
		mockBlockData.GetHeight(),
		mockBlockData.GetTimestamp(),
		mockBlockData.GetBlockSeed(),
		mockBlockData.GetBlockSignature(),
		mockBlockData.GetCumulativeDifficulty(),
		mockBlockData.GetPayloadLength(),
		mockBlockData.GetPayloadHash(),
		mockBlockData.GetBlocksmithPublicKey(),
		mockBlockData.GetTotalAmount(),
		mockBlockData.GetTotalFee(),
		mockBlockData.GetTotalCoinBase(),
		mockBlockData.GetVersion(),
	))
	return db.Query("")
}

func (*mockQueryExecutorCheckGenesisTrue) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.MainChain{}).Fields,
		).AddRow(
			mockBlockData.GetID(),
			mockBlockData.GetBlockHash(),
			mockBlockData.GetPreviousBlockHash(),
			mockBlockData.GetHeight(),
			mockBlockData.GetTimestamp(),
			mockBlockData.GetBlockSeed(),
			mockBlockData.GetBlockSignature(),
			mockBlockData.GetCumulativeDifficulty(),
			mockBlockData.GetPayloadLength(),
			mockBlockData.GetPayloadHash(),
			mockBlockData.GetBlocksmithPublicKey(),
			mockBlockData.GetTotalAmount(),
			mockBlockData.GetTotalFee(),
			mockBlockData.GetTotalCoinBase(),
			mockBlockData.GetVersion(),
		))
	return db.QueryRow(qStr), nil
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
		Logger             *log.Logger
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
				Logger:        log.New(),
			},
			want: true,
		},
		{
			name: "wantFalse",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorCheckGenesisFalse{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
				Logger:        log.New(),
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
				Logger:             tt.fields.Logger,
			}
			if got := bs.CheckGenesis(); got != tt.want {
				t.Errorf("BlockService.CheckGenesis() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryExecutorGetBlockByHeightSuccess struct {
		query.Executor
	}
	mockQueryExecutorGetBlockByHeightFail struct {
		query.Executor
	}
)

func (*mockQueryExecutorGetBlockByHeightSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	switch qStr {
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version FROM main_block WHERE height = 0":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.MainChain{}).Fields).AddRow(
			mockBlockData.GetID(),
			mockBlockData.GetBlockHash(),
			mockBlockData.GetPreviousBlockHash(),
			mockBlockData.GetHeight(),
			mockBlockData.GetTimestamp(),
			mockBlockData.GetBlockSeed(),
			mockBlockData.GetBlockSignature(),
			mockBlockData.GetCumulativeDifficulty(),
			mockBlockData.GetPayloadLength(),
			mockBlockData.GetPayloadHash(),
			mockBlockData.GetBlocksmithPublicKey(),
			mockBlockData.GetTotalAmount(),
			mockBlockData.GetTotalFee(),
			mockBlockData.GetTotalCoinBase(),
			mockBlockData.GetVersion(),
		))
	case "SELECT id, block_id, block_height, sender_account_address, recipient_account_address, transaction_type, " +
		"fee, timestamp, transaction_hash, transaction_body_length, transaction_body_bytes, " +
		"signature, version, transaction_index FROM \"transaction\" WHERE block_id = ? ORDER BY transaction_index ASC":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
			query.NewTransactionQuery(&chaintype.MainChain{}).Fields))
	}
	return db.Query(qStr)
}

func (*mockQueryExecutorGetBlockByHeightFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
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
				Chaintype:        &chaintype.MainChain{},
				QueryExecutor:    &mockQueryExecutorGetBlockByHeightSuccess{},
				BlockQuery:       query.NewBlockQuery(&chaintype.MainChain{}),
				TransactionQuery: query.NewTransactionQuery(&chaintype.MainChain{}),
			},
			want:    &mockBlockData,
			wantErr: false,
		},
		{
			name: "GetBlockByHeight:FailNoEntryFound", // All is good
			fields: fields{
				Chaintype:        &chaintype.MainChain{},
				QueryExecutor:    &mockQueryExecutorGetBlockByHeightFail{},
				BlockQuery:       query.NewBlockQuery(&chaintype.MainChain{}),
				TransactionQuery: query.NewTransactionQuery(&chaintype.MainChain{}),
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

type (
	mockQueryExecutorGetBlockByIDSuccess struct {
		query.Executor
	}
	mockQueryExecutorGetBlockByIDFail struct {
		query.Executor
	}
)

func (*mockQueryExecutorGetBlockByIDSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	switch qStr {
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, " +
		"version FROM main_block WHERE id = 1":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.MainChain{}).Fields).AddRow(
			mockBlockData.GetID(),
			mockBlockData.GetBlockHash(),
			mockBlockData.GetPreviousBlockHash(),
			mockBlockData.GetHeight(),
			mockBlockData.GetTimestamp(),
			mockBlockData.GetBlockSeed(),
			mockBlockData.GetBlockSignature(),
			mockBlockData.GetCumulativeDifficulty(),
			mockBlockData.GetPayloadLength(),
			mockBlockData.GetPayloadHash(),
			mockBlockData.GetBlocksmithPublicKey(),
			mockBlockData.GetTotalAmount(),
			mockBlockData.GetTotalFee(),
			mockBlockData.GetTotalCoinBase(),
			mockBlockData.GetVersion(),
		))
	case "SELECT id, block_id, block_height, sender_account_address, recipient_account_address, transaction_type, " +
		"fee, timestamp, transaction_hash, transaction_body_length, transaction_body_bytes, " +
		"signature, version, transaction_index FROM \"transaction\" WHERE block_id = ? ORDER BY transaction_index ASC":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
			query.NewTransactionQuery(&chaintype.MainChain{}).Fields))
	}
	return db.Query(qStr)
}

func (*mockQueryExecutorGetBlockByIDFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}

func (*mockQueryExecutorGetBlockByIDSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(query.NewBlockQuery(&chaintype.MainChain{}).Fields).AddRow(
			mockBlockData.GetID(),
			mockBlockData.GetBlockHash(),
			mockBlockData.GetPreviousBlockHash(),
			mockBlockData.GetHeight(),
			mockBlockData.GetTimestamp(),
			mockBlockData.GetBlockSeed(),
			mockBlockData.GetBlockSignature(),
			mockBlockData.GetCumulativeDifficulty(),
			mockBlockData.GetPayloadLength(),
			mockBlockData.GetPayloadHash(),
			mockBlockData.GetBlocksmithPublicKey(),
			mockBlockData.GetTotalAmount(),
			mockBlockData.GetTotalFee(),
			mockBlockData.GetTotalCoinBase(),
			mockBlockData.GetVersion(),
		))
	return db.QueryRow(qStr), nil
}

func (*mockQueryExecutorGetBlockByIDFail) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, errors.New("MockedError")
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
		ID               int64
		withAttachedData bool
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
				Chaintype:        &chaintype.MainChain{},
				QueryExecutor:    &mockQueryExecutorGetBlockByIDSuccess{},
				BlockQuery:       query.NewBlockQuery(&chaintype.MainChain{}),
				TransactionQuery: query.NewTransactionQuery(&chaintype.MainChain{}),
			},
			args: args{
				ID:               int64(1),
				withAttachedData: true,
			},
			want:    &mockBlockData,
			wantErr: false,
		},
		{
			name: "GetBlockByID:FailNoEntryFound", // All is good
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorGetBlockByIDFail{},
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
			got, err := bs.GetBlockByID(tt.args.ID, tt.args.withAttachedData)
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

type (
	mockQueryExecutorGetBlocksFromHeightSuccess struct {
		query.Executor
	}

	mockQueryExecutorGetBlocksFromHeightFail struct {
		query.Executor
	}
)

func (*mockQueryExecutorGetBlocksFromHeightSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	mock.ExpectQuery(qStr).WillReturnRows(sqlmock.NewRows(
		query.NewBlockQuery(&chaintype.MainChain{}).Fields,
	).AddRow(
		mockBlockData.GetID(),
		mockBlockData.GetBlockHash(),
		mockBlockData.GetPreviousBlockHash(),
		mockBlockData.GetHeight(),
		mockBlockData.GetTimestamp(),
		mockBlockData.GetBlockSeed(),
		mockBlockData.GetBlockSignature(),
		mockBlockData.GetCumulativeDifficulty(),
		mockBlockData.GetPayloadLength(),
		mockBlockData.GetPayloadHash(),
		mockBlockData.GetBlocksmithPublicKey(),
		mockBlockData.GetTotalAmount(),
		mockBlockData.GetTotalFee(),
		mockBlockData.GetTotalCoinBase(),
		mockBlockData.GetVersion(),
	).AddRow(
		mockBlockData.GetID(),
		mockBlockData.GetBlockHash(),
		mockBlockData.GetPreviousBlockHash(),
		mockBlockData.GetHeight(),
		mockBlockData.GetTimestamp(),
		mockBlockData.GetBlockSeed(),
		mockBlockData.GetBlockSignature(),
		mockBlockData.GetCumulativeDifficulty(),
		mockBlockData.GetPayloadLength(),
		mockBlockData.GetPayloadHash(),
		mockBlockData.GetBlocksmithPublicKey(),
		mockBlockData.GetTotalAmount(),
		mockBlockData.GetTotalFee(),
		mockBlockData.GetTotalCoinBase(),
		mockBlockData.GetVersion(),
	),
	)
	return db.Query(qStr)
}

func (*mockQueryExecutorGetBlocksFromHeightFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
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
				QueryExecutor: &mockQueryExecutorGetBlocksFromHeightSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			args: args{
				startHeight: 0,
				limit:       2,
			},
			want: []*model.Block{
				&mockBlockData,
				&mockBlockData,
			},
			wantErr: false,
		},
		{
			name: "GetBlocksFromHeight:FailNoEntryFound", // All is good
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				QueryExecutor: &mockQueryExecutorGetBlocksFromHeightFail{},
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
	var (
		mockLastBlockData = model.Block{
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
			Transactions: []*model.Transaction{
				mockTransaction,
			},
		}

		mockGoodLastBlockHash, _ = util.GetBlockHash(&mockLastBlockData, &chaintype.MainChain{})
		mockGoodIncomingBlock    = &model.Block{
			PreviousBlockHash:    mockGoodLastBlockHash,
			BlockSignature:       nil,
			CumulativeDifficulty: "200",
			Timestamp:            10000,
			BlocksmithPublicKey:  mockBlocksmiths[0].NodePublicKey,
			Transactions: []*model.Transaction{
				mockTransaction,
			},
		}
		successBlockHash = []byte{
			197, 250, 152, 172, 169, 236, 102, 225, 55, 58, 90, 101, 214, 217, 209, 67, 185, 183, 116, 101, 64, 47, 196,
			207, 27, 173, 3, 141, 12, 163, 245, 254,
		}
		mockBlockSuccess = &model.Block{
			BlockSignature:    []byte{},
			BlockHash:         successBlockHash,
			PreviousBlockHash: make([]byte, 32),
			Transactions:      make([]*model.Transaction, 0),
		}
	)
	mockBlockData.BlockHash = mockGoodLastBlockHash

	type fields struct {
		Chaintype               chaintype.ChainType
		KVExecutor              kvdb.KVExecutorInterface
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		MerkleTreeQuery         query.MerkleTreeQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		SkippedBlocksmithQuery  query.SkippedBlocksmithQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		AccountLedgerQuery      query.AccountLedgerQueryInterface
		BlocksmithStrategy      strategy.BlocksmithStrategyInterface
		Observer                *observer.Observer
		NodeRegistrationService NodeRegistrationServiceInterface
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
				Chaintype:               &chaintype.MainChain{},
				QueryExecutor:           nil,
				BlockQuery:              nil,
				MempoolQuery:            query.NewMempoolQuery(&chaintype.MainChain{}),
				TransactionQuery:        nil,
				Signature:               nil,
				MempoolService:          nil,
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				AccountLedgerQuery:      nil,
				Observer:                nil,
				NodeRegistrationService: nil,
				BlocksmithStrategy:      &mockBlocksmithService{},
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "ReceiveBlock:fail - {last block hash != previousBlockHash}",
			args: args{
				senderPublicKey: nil,
				lastBlock: &model.Block{
					BlockHash:      []byte{1},
					BlockSignature: []byte{},
				},
				block: &model.Block{
					PreviousBlockHash: []byte{},
					BlockSignature:    nil,
				},
				nodeSecretPhrase: "",
			},
			fields: fields{
				Chaintype:               &chaintype.MainChain{},
				KVExecutor:              &mockKVExecutorSuccess{},
				QueryExecutor:           nil,
				BlockQuery:              nil,
				MempoolQuery:            query.NewMempoolQuery(&chaintype.MainChain{}),
				TransactionQuery:        nil,
				Signature:               &mockSignature{},
				MempoolService:          nil,
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				AccountLedgerQuery:      nil,
				Observer:                nil,
				BlocksmithStrategy:      &mockBlocksmithService{},
				NodeRegistrationService: nil,
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "ReceiveBlock:fail - {last block hash != previousBlockHash - kvExecutor KeyNotFound - generate batch receipt success}",
			args: args{
				senderPublicKey:  []byte{1, 3, 4, 5, 6},
				lastBlock:        mockBlockSuccess,
				block:            mockBlockSuccess,
				nodeSecretPhrase: "",
			},
			fields: fields{
				Chaintype:               &chaintype.MainChain{},
				KVExecutor:              &mockKVExecutorSuccessKeyNotFound{},
				QueryExecutor:           &mockQueryExecutorSuccess{},
				BlockQuery:              nil,
				MempoolQuery:            query.NewMempoolQuery(&chaintype.MainChain{}),
				MerkleTreeQuery:         query.NewMerkleTreeQuery(),
				TransactionQuery:        nil,
				Signature:               &mockSignature{},
				MempoolService:          nil,
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				AccountLedgerQuery:      nil,
				Observer:                nil,
				NodeRegistrationService: nil,
				BlocksmithStrategy:      &mockBlocksmithService{},
			},
			wantErr: false,
			want: &model.BatchReceipt{
				SenderPublicKey: []byte{1, 3, 4, 5, 6},
				RecipientPublicKey: []byte{
					88, 220, 21, 76, 132, 107, 209, 213, 213, 206, 112, 50, 201, 183, 134, 250, 90, 163, 91, 63, 176,
					223, 177, 77, 197, 161, 178, 55, 31, 225, 233, 115,
				},
				DatumHash:            successBlockHash,
				DatumType:            constant.ReceiptDatumTypeBlock,
				ReferenceBlockHeight: 0,
				ReferenceBlockHash:   successBlockHash,
				RMRLinked:            nil,
				RecipientSignature:   []byte{},
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
					BlocksmithPublicKey: []byte{1, 3, 4, 5, 6},
				},
				nodeSecretPhrase: "",
			},
			fields: fields{
				Chaintype:               &chaintype.MainChain{},
				KVExecutor:              &mockKVExecutorFailOtherError{},
				QueryExecutor:           &mockQueryExecutorSuccess{},
				BlockQuery:              nil,
				MempoolQuery:            query.NewMempoolQuery(&chaintype.MainChain{}),
				TransactionQuery:        nil,
				Signature:               &mockSignature{},
				MempoolService:          nil,
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				AccountLedgerQuery:      nil,
				Observer:                nil,
				NodeRegistrationService: nil,
				BlocksmithStrategy:      &mockBlocksmithService{},
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "ReceiveBlock:pushBlockFail",
			args: args{
				senderPublicKey:  []byte{1, 3, 4, 5, 6},
				lastBlock:        &mockBlockData,
				block:            mockGoodIncomingBlock,
				nodeSecretPhrase: "",
			},
			fields: fields{
				Chaintype:               &chaintype.MainChain{},
				QueryExecutor:           &mockQueryExecutorFail{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:            query.NewMempoolQuery(&chaintype.MainChain{}),
				TransactionQuery:        nil,
				Signature:               &mockSignature{},
				MempoolService:          nil,
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				AccountLedgerQuery:      nil,
				Observer:                observer.NewObserver(),
				NodeRegistrationService: nil,
				BlocksmithStrategy:      &mockBlocksmithService{},
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "ReceiveBlock:success",
			args: args{
				senderPublicKey:  []byte{1, 3, 4, 5, 6},
				lastBlock:        &mockBlockData,
				block:            mockGoodIncomingBlock,
				nodeSecretPhrase: "",
			},
			fields: fields{
				Chaintype:               &chaintype.MainChain{},
				KVExecutor:              &mockKVExecutorSuccess{},
				QueryExecutor:           &mockQueryExecutorSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:            query.NewMempoolQuery(&chaintype.MainChain{}),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				TransactionQuery:        query.NewTransactionQuery(&chaintype.MainChain{}),
				MerkleTreeQuery:         query.NewMerkleTreeQuery(),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				SkippedBlocksmithQuery:  query.NewSkippedBlocksmithQuery(),
				Signature:               &mockSignature{},
				MempoolService: &mockMempoolServiceSelectSuccess{
					MempoolService{
						QueryExecutor:      &mockQueryExecutorMempoolSuccess{},
						ActionTypeSwitcher: &mockTypeActionSuccess{},
					},
				},
				ActionTypeSwitcher:      &mockTypeActionSuccess{},
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				AccountLedgerQuery:      query.NewAccountLedgerQuery(),
				Observer:                observer.NewObserver(),
				BlocksmithStrategy:      &mockBlocksmithServicePushBlock{},
				NodeRegistrationService: &mockNodeRegistrationServiceSuccess{},
			},
			wantErr: false,
			want: &model.BatchReceipt{
				SenderPublicKey: []byte{1, 3, 4, 5, 6},
				RecipientPublicKey: []byte{
					88, 220, 21, 76, 132, 107, 209, 213, 213, 206, 112, 50, 201, 183, 134, 250, 90, 163, 91, 63, 176,
					223, 177, 77, 197, 161, 178, 55, 31, 225, 233, 115,
				},
				DatumType:            constant.ReceiptDatumTypeBlock,
				ReferenceBlockHeight: mockBlockData.GetHeight(),
				ReferenceBlockHash:   mockGoodLastBlockHash,
				RMRLinked:            nil,
				RecipientSignature:   []byte{},
			},
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
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				SkippedBlocksmithQuery:  tt.fields.SkippedBlocksmithQuery,
				Signature:               tt.fields.Signature,
				MempoolService:          tt.fields.MempoolService,
				ActionTypeSwitcher:      tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				AccountLedgerQuery:      tt.fields.AccountLedgerQuery,
				Observer:                tt.fields.Observer,
				BlocksmithStrategy:      tt.fields.BlocksmithStrategy,
				Logger:                  logrus.New(),
				NodeRegistrationService: tt.fields.NodeRegistrationService,
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
		PublishedReceiptQuery   query.PublishedReceiptQueryInterface
		SkippedBlocksmithQuery  query.SkippedBlocksmithQueryInterface
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
				QueryExecutor:          &mockQueryExecutorNotFound{},
				NodeRegistrationQuery:  query.NewNodeRegistrationQuery(),
				PublishedReceiptQuery:  query.NewPublishedReceiptQuery(),
				SkippedBlocksmithQuery: query.NewSkippedBlocksmithQuery(),
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
				QueryExecutor:          &mockQueryExecutorSuccess{},
				NodeRegistrationQuery:  query.NewNodeRegistrationQuery(),
				PublishedReceiptQuery:  query.NewPublishedReceiptQuery(),
				SkippedBlocksmithQuery: query.NewSkippedBlocksmithQuery(),
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
					PayloadLength:        0,
					PayloadHash:          []byte{},
					BlocksmithPublicKey:  bcsNodePubKey1,
					TotalAmount:          100000000,
					TotalFee:             10000000,
					TotalCoinBase:        1,
					Version:              0,
				},
				BlocksmithAccountAddress: constant.MainchainGenesisAccountAddress,
				TotalReceipts:            1,
				ReceiptValue:             50000000,
				PopChange:                1000000000,
				SkippedBlocksmiths: []*model.SkippedBlocksmith{
					{
						BlocksmithPublicKey: mockBlocksmiths[0].NodePublicKey,
						POPChange:           5000,
						BlockHeight:         1,
					},
				},
			},
		},
		{
			name: "GetBlockExtendedInfo:success",
			args: args{
				block: block,
			},
			fields: fields{
				QueryExecutor:          &mockQueryExecutorSuccess{},
				NodeRegistrationQuery:  query.NewNodeRegistrationQuery(),
				PublishedReceiptQuery:  query.NewPublishedReceiptQuery(),
				SkippedBlocksmithQuery: query.NewSkippedBlocksmithQuery(),
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
					PayloadLength:        0,
					PayloadHash:          []byte{},
					BlocksmithPublicKey:  bcsNodePubKey1,
					TotalAmount:          100000000,
					TotalFee:             10000000,
					TotalCoinBase:        1,
					Version:              0,
				},
				BlocksmithAccountAddress: bcsAddress1,
				TotalReceipts:            int64(len(mockPublishedReceipt)),
				ReceiptValue:             50000000,
				PopChange:                1000000000,
				SkippedBlocksmiths: []*model.SkippedBlocksmith{
					{
						BlocksmithPublicKey: mockBlocksmiths[0].NodePublicKey,
						POPChange:           5000,
						BlockHeight:         1,
					},
				},
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
				PublishedReceiptQuery:   tt.fields.PublishedReceiptQuery,
				SkippedBlocksmithQuery:  tt.fields.SkippedBlocksmithQuery,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				Observer:                tt.fields.Observer,
			}
			got, err := bs.GetBlockExtendedInfo(tt.args.block, false)
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
		AccountLedgerQuery      query.AccountLedgerQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		Observer                *observer.Observer
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
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
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
				AccountLedgerQuery:      tt.fields.AccountLedgerQuery,
				Signature:               tt.fields.Signature,
				MempoolService:          tt.fields.MempoolService,
				ActionTypeSwitcher:      tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				Observer:                tt.fields.Observer,
			}
			if err := bs.RewardBlocksmithAccountAddresses(tt.args.blocksmithAccountAddresses, tt.args.totalReward,
				tt.args.height); (err != nil) != tt.wantErr {
				t.Errorf("BlockService.RewardBlocksmithAccountAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockBlocksmithService struct {
		strategy.BlocksmithStrategyMain
	}
)

func (*mockBlocksmithService) GetSortedBlocksmiths(block *model.Block) []*model.Blocksmith {
	return []*model.Blocksmith{
		{
			NodeID:        1,
			NodeOrder:     new(big.Int).SetInt64(8000),
			NodePublicKey: []byte{1, 3, 4, 5, 6},
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
}
func TestBlockService_CoinbaseLotteryWinners(t *testing.T) {

	type fields struct {
		Chaintype               chaintype.ChainType
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		BlocksmithStrategy      strategy.BlocksmithStrategyInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		Observer                *observer.Observer
	}
	type args struct {
		blocksmiths []*model.Blocksmith
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		args    args
		wantErr bool
	}{
		{
			name: "CoinbaseLotteryWinners:success",
			fields: fields{
				QueryExecutor:         &mockQueryExecutorSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			wantErr: false,
			args: args{blocksmiths: []*model.Blocksmith{
				{
					NodeID:        1,
					NodeOrder:     new(big.Int).SetInt64(8000),
					NodePublicKey: []byte{1, 3, 4, 5, 6},
				},
				{
					NodeID:    2,
					NodeOrder: new(big.Int).SetInt64(1000),
				},
				{
					NodeID:    3,
					NodeOrder: new(big.Int).SetInt64(5000),
				},
			}},
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
				BlocksmithStrategy:      tt.fields.BlocksmithStrategy,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				Observer:                tt.fields.Observer,
			}
			got, err := bs.CoinbaseLotteryWinners(tt.args.blocksmiths)
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
		Logger                  *log.Logger
	}
	type args struct {
		genesisEntries []constant.GenesisConfigEntry
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
			},
			args: args{
				genesisEntries: []constant.GenesisConfigEntry{
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
			want:    4070746053101615238,
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
				Logger:                  tt.fields.Logger,
			}
			got, err := bs.GenerateGenesisBlock(tt.args.genesisEntries)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GenerateGenesisBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.ID != tt.want {
				t.Errorf("BlockService.GenerateGenesisBlock() got %v, want %v", got.GetID(), tt.want)
			}
		})
	}
}

type mockQueryExecutorValidateBlockSuccess struct {
	query.Executor
}

func (*mockQueryExecutorValidateBlockSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.MainChain{}).Fields,
		).AddRow(
			mockBlockData.GetID(),
			mockBlockData.GetBlockHash(),
			mockBlockData.GetPreviousBlockHash(),
			mockBlockData.GetHeight(),
			mockBlockData.GetTimestamp(),
			mockBlockData.GetBlockSeed(),
			mockBlockData.GetBlockSignature(),
			mockBlockData.GetCumulativeDifficulty(),
			mockBlockData.GetPayloadLength(),
			mockBlockData.GetPayloadHash(),
			mockBlockData.GetBlocksmithPublicKey(),
			mockBlockData.GetTotalAmount(),
			mockBlockData.GetTotalFee(),
			mockBlockData.GetTotalCoinBase(),
			mockBlockData.GetVersion(),
		))
	rows, _ := db.Query(qStr)
	return rows, nil
}

var (
	mockValidateBadBlockInvalidBlockHash = &model.Block{
		Timestamp:           1572246820,
		BlockSignature:      []byte{},
		BlocksmithPublicKey: []byte{1, 2, 3, 4},
		PreviousBlockHash:   []byte{},
	}

	mockValidateBlockSuccess = &model.Block{
		Timestamp: 1572246820,
		ID:        constant.MainchainGenesisBlockID,
		BlockHash: make([]byte, 32),
		PreviousBlockHash: []byte{167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160,
			97, 214, 98, 245, 128, 255, 77, 228, 59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74},
		Height: 1,
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
	mockBlocksmithServiceValidateBlockSuccess struct {
		strategy.BlocksmithStrategyMain
	}
)

func (*mockBlocksmithServiceValidateBlockSuccess) GetSortedBlocksmithsMap(*model.Block) map[string]*int64 {
	firstIndex := int64(0)
	secondIndex := int64(1)
	return map[string]*int64{
		string(mockValidateBadBlockInvalidBlockHash.BlocksmithPublicKey): &firstIndex,
		string(mockBlockData.BlocksmithPublicKey):                        &secondIndex,
	}
}

func TestBlockService_ValidateBlock(t *testing.T) {
	type fields struct {
		Chaintype               chaintype.ChainType
		KVExecutor              kvdb.KVExecutorInterface
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		MerkleTreeQuery         query.MerkleTreeQueryInterface
		PublishedReceiptQuery   query.PublishedReceiptQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		ReceiptService          ReceiptServiceInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlocksmithStrategy      strategy.BlocksmithStrategyInterface
		Observer                *observer.Observer
		Logger                  *log.Logger
	}
	type args struct {
		block             *model.Block
		previousLastBlock *model.Block
		curTime           int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ValidateBlock:fail-{InvalidTimestamp}",
			args: args{
				block: &model.Block{
					Timestamp: 1572246820 + constant.GenerateBlockTimeoutSec + 1,
				},
				curTime: 1572246820,
			},
			fields:  fields{},
			wantErr: true,
		},
		{
			name: "ValidateBlock:fail-{notInBlocksmithList}",
			args: args{
				block: &model.Block{
					Timestamp:           1572246820,
					BlockSignature:      []byte{},
					BlocksmithPublicKey: []byte{},
				},
				curTime: 1572246820,
			},
			fields: fields{
				Signature:          &mockSignatureFail{},
				BlocksmithStrategy: &mockBlocksmithServiceValidateBlockSuccess{},
			},
			wantErr: true,
		},
		{
			name: "ValidateBlock:fail-{InvalidSignature}",
			args: args{
				block:   mockValidateBadBlockInvalidBlockHash,
				curTime: 1572246820,
			},
			fields: fields{
				Signature:          &mockSignatureFail{},
				BlocksmithStrategy: &mockBlocksmithServiceValidateBlockSuccess{},
			},
			wantErr: true,
		},
		{
			name: "ValidateBlock:fail-{InvalidBlockHash}",
			args: args{
				block:             mockValidateBadBlockInvalidBlockHash,
				previousLastBlock: &model.Block{},
				curTime:           1572246820,
			},
			fields: fields{
				Signature:          &mockSignature{},
				BlocksmithStrategy: &mockBlocksmithServiceValidateBlockSuccess{},
			},
			wantErr: true,
		},
		{
			name: "ValidateBlock:fail-{InvalidCumulativeDifficulty}",
			args: args{
				block: &model.Block{
					Timestamp:           1572246820,
					BlockSignature:      []byte{},
					BlocksmithPublicKey: []byte{},
					PreviousBlockHash: []byte{167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160,
						97, 214, 98, 245, 128, 255, 77, 228, 59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74},
					CumulativeDifficulty: "10",
				},
				previousLastBlock: &model.Block{},
				curTime:           1572246820,
			},
			fields: fields{
				Signature:          &mockSignature{},
				BlockQuery:         query.NewBlockQuery(&chaintype.MainChain{}),
				QueryExecutor:      &mockQueryExecutorValidateBlockSuccess{},
				BlocksmithStrategy: &mockBlocksmithServiceValidateBlockSuccess{},
			},
			wantErr: true,
		},
		{
			name: "ValidateBlock:success",
			args: args{
				block:             mockValidateBlockSuccess,
				previousLastBlock: &model.Block{},
				curTime:           mockValidateBlockSuccess.Timestamp,
			},
			fields: fields{
				Signature:          &mockSignature{},
				BlockQuery:         query.NewBlockQuery(&chaintype.MainChain{}),
				QueryExecutor:      &mockQueryExecutorValidateBlockSuccess{},
				BlocksmithStrategy: &mockBlocksmithServiceValidateBlockSuccess{},
			},
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
				PublishedReceiptQuery:   tt.fields.PublishedReceiptQuery,
				Signature:               tt.fields.Signature,
				MempoolService:          tt.fields.MempoolService,
				ReceiptService:          tt.fields.ReceiptService,
				ActionTypeSwitcher:      tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlocksmithStrategy:      tt.fields.BlocksmithStrategy,
				Observer:                tt.fields.Observer,
				Logger:                  tt.fields.Logger,
			}
			if err := bs.ValidateBlock(tt.args.block, tt.args.previousLastBlock, tt.args.curTime); (err != nil) != tt.wantErr {
				t.Errorf("BlockService.ValidateBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockPopOffToBlockReturnCommonBlock struct {
		query.Executor
	}
	mockPopOffToBlockReturnBeginTxFunc struct {
		query.Executor
	}
	mockPopOffToBlockReturnWantFailOnCommit struct {
		query.Executor
	}
	mockPopOffToBlockReturnWantFailOnExecuteTransactions struct {
		query.Executor
	}
)

func (*mockPopOffToBlockReturnCommonBlock) BeginTx() error {
	return nil
}
func (*mockPopOffToBlockReturnCommonBlock) CommitTx() error {
	return nil
}
func (*mockPopOffToBlockReturnCommonBlock) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}
func (*mockPopOffToBlockReturnCommonBlock) ExecuteSelect(qSrt string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(query.NewMempoolQuery(chaintype.GetChainType(0)).Fields).AddRow(
			1,
			0,
			10,
			1000,
			[]byte{2, 0, 0, 0, 1, 112, 240, 249, 74, 0, 0, 0, 0, 44, 0, 0, 0, 66, 67, 90, 69, 71, 79, 98, 51, 87, 78, 120, 51,
				102, 68, 79, 86, 102, 57, 90, 83, 52, 69, 106, 118, 79, 73, 118, 95, 85, 101, 87, 52, 84, 86, 66, 81, 74, 95, 54,
				116, 72, 75, 108, 69, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 201, 0, 0, 0, 153, 58, 50, 200, 7, 61,
				108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213,
				135, 0, 0, 0, 0, 9, 0, 0, 0, 49, 50, 55, 46, 48, 46, 48, 46, 49, 0, 202, 154, 59, 0, 0, 0, 0, 86, 90, 118, 89, 100,
				56, 48, 112, 53, 83, 45, 114, 120, 83, 78, 81, 109, 77, 90, 119, 89, 88, 67, 55, 76, 121, 65, 122, 66, 109, 99, 102,
				99, 106, 52, 77, 85, 85, 65, 100, 117, 100, 87, 77, 198, 224, 91, 94, 235, 56, 96, 236, 211, 155, 119, 159, 171, 196,
				10, 175, 144, 215, 90, 167, 3, 27, 88, 212, 233, 202, 31, 112, 45, 147, 34, 18, 1, 0, 0, 0, 48, 128, 236, 38, 196, 0,
				66, 232, 114, 70, 30, 220, 206, 222, 141, 50, 152, 151, 150, 235, 72, 86, 150, 96, 70, 162, 253, 128, 108, 95, 26, 175,
				178, 108, 74, 76, 98, 68, 141, 131, 57, 209, 224, 251, 129, 224, 47, 156, 120, 9, 77, 251, 236, 230, 212, 109, 193, 67,
				250, 166, 49, 249, 198, 11, 0, 0, 0, 0, 162, 190, 223, 52, 221, 118, 195, 111, 129, 166, 99, 216, 213, 202, 203, 118, 28,
				231, 39, 137, 123, 228, 86, 52, 100, 8, 124, 254, 19, 181, 202, 139, 211, 184, 202, 54, 8, 166, 131, 96, 244, 101, 76,
				167, 176, 172, 85, 88, 93, 32, 173, 123, 229, 109, 128, 26, 192, 70, 155, 217, 107, 210, 254, 15},
			"BCZ",
			"ZCB",
		),
	)
	return db.Query("")
}
func (*mockPopOffToBlockReturnCommonBlock) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}
func (*mockPopOffToBlockReturnBeginTxFunc) BeginTx() error {
	return errors.New("i want this")
}
func (*mockPopOffToBlockReturnBeginTxFunc) CommitTx() error {
	return nil
}
func (*mockPopOffToBlockReturnWantFailOnCommit) BeginTx() error {
	return nil
}
func (*mockPopOffToBlockReturnWantFailOnCommit) CommitTx() error {
	return errors.New("i want this")
}
func (*mockPopOffToBlockReturnWantFailOnCommit) ExecuteSelect(qSrt string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(query.NewMempoolQuery(chaintype.GetChainType(0)).Fields).AddRow(
			1,
			0,
			10,
			1000,
			[]byte{1, 2, 3, 4, 5},
			"BCZ",
			"ZCB",
		),
	)
	return db.Query("")

}
func (*mockPopOffToBlockReturnWantFailOnExecuteTransactions) BeginTx() error {
	return nil
}
func (*mockPopOffToBlockReturnWantFailOnExecuteTransactions) CommitTx() error {
	return nil
}
func (*mockPopOffToBlockReturnWantFailOnExecuteTransactions) ExecuteTransactions(queries [][]interface{}) error {
	return errors.New("i want this")
}
func (*mockPopOffToBlockReturnWantFailOnExecuteTransactions) RollbackTx() error {
	return nil
}

var (
	mockGoodBlock = &model.Block{
		ID:                   1,
		BlockHash:            nil,
		PreviousBlockHash:    nil,
		Height:               1000,
		Timestamp:            0,
		BlockSeed:            nil,
		BlockSignature:       nil,
		CumulativeDifficulty: "",
		BlocksmithPublicKey:  nil,
		TotalAmount:          0,
		TotalFee:             0,
		TotalCoinBase:        0,
		Version:              0,
		PayloadLength:        0,
		PayloadHash:          nil,
		Transactions:         nil,
		PublishedReceipts:    nil,
	}
	mockGoodCommonBlock = &model.Block{
		ID:                   1,
		BlockHash:            nil,
		PreviousBlockHash:    nil,
		Height:               900,
		Timestamp:            0,
		BlockSeed:            nil,
		BlockSignature:       nil,
		CumulativeDifficulty: "",
		BlocksmithPublicKey:  nil,
		TotalAmount:          0,
		TotalFee:             0,
		TotalCoinBase:        0,
		Version:              0,
		PayloadLength:        0,
		PayloadHash:          nil,
		Transactions:         nil,
		PublishedReceipts:    nil,
	}
	mockBadCommonBlockHardFork = &model.Block{
		ID:                   1,
		BlockHash:            nil,
		PreviousBlockHash:    nil,
		Height:               100,
		Timestamp:            0,
		BlockSeed:            nil,
		BlockSignature:       nil,
		CumulativeDifficulty: "",
		BlocksmithPublicKey:  nil,
		TotalAmount:          0,
		TotalFee:             0,
		TotalCoinBase:        0,
		Version:              0,
		PayloadLength:        0,
		PayloadHash:          nil,
		Transactions:         nil,
		PublishedReceipts:    nil,
	}
)

type (
	mockExecutorBlockPopGetLastBlockFail struct {
		query.Executor
	}
	mockExecutorBlockPopSuccess struct {
		query.Executor
	}
	mockExecutorBlockPopFailCommonNotFound struct {
		mockExecutorBlockPopSuccess
	}
	mockReceiptSuccess struct {
		ReceiptService
	}
	mockReceiptFail struct {
		ReceiptService
	}
	mockMempoolServiceBlockPopSuccess struct {
		MempoolService
	}
	mockMempoolServiceBlockPopFail struct {
		MempoolService
	}
	mockNodeRegistrationServiceBlockPopSuccess struct {
		NodeRegistrationService
	}
)

func (*mockExecutorBlockPopFailCommonNotFound) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	transactionQ := query.NewTransactionQuery(&chaintype.MainChain{})
	blockQ := query.NewBlockQuery(&chaintype.MainChain{})
	switch qStr {
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version FROM main_block WHERE id = 0":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields))
	case "SELECT id, block_id, block_height, sender_account_address, recipient_account_address, transaction_type, fee, " +
		"timestamp, transaction_hash, transaction_body_length, transaction_body_bytes, signature, version, " +
		"transaction_index FROM \"transaction\" WHERE block_id = ? ORDER BY transaction_index ASC":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(transactionQ.Fields))
	}

	return db.Query(qStr)
}

func (*mockExecutorBlockPopFailCommonNotFound) ExecuteSelectRow(
	qStr string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	blockQ := query.NewBlockQuery(&chaintype.MainChain{})
	switch qStr {
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version FROM main_block ORDER BY height DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields[:len(blockQ.Fields)-1]).AddRow(
				mockGoodBlock.GetID(),
				mockGoodBlock.GetBlockHash(),
				mockGoodBlock.GetPreviousBlockHash(),
				mockGoodBlock.GetHeight(),
				mockGoodBlock.GetTimestamp(),
				mockGoodBlock.GetBlockSeed(),
				mockGoodBlock.GetBlockSignature(),
				mockGoodBlock.GetCumulativeDifficulty(),
				mockGoodBlock.GetPayloadLength(),
				mockGoodBlock.GetPayloadHash(),
				mockGoodBlock.GetBlocksmithPublicKey(),
				mockGoodBlock.GetTotalAmount(),
				mockGoodBlock.GetTotalFee(),
				mockGoodBlock.GetTotalCoinBase(),
			),
		)
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version FROM main_block WHERE id = 1":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields))
	default:
		return nil, fmt.Errorf("unmocked query: %s", qStr)
	}
	return db.QueryRow(qStr), nil
}

func (*mockExecutorBlockPopGetLastBlockFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	blockQ := query.NewBlockQuery(&chaintype.MainChain{})
	switch qStr {
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version FROM main_block WHERE id = 0":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields))
	default:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields[:len(blockQ.Fields)-1]).AddRow(
				mockGoodBlock.GetID(),
				mockGoodBlock.GetBlockHash(),
				mockGoodBlock.GetPreviousBlockHash(),
				mockGoodBlock.GetHeight(),
				mockGoodBlock.GetTimestamp(),
				mockGoodBlock.GetBlockSeed(),
				mockGoodBlock.GetBlockSignature(),
				mockGoodBlock.GetCumulativeDifficulty(),
				mockGoodBlock.GetPayloadLength(),
				mockGoodBlock.GetPayloadHash(),
				mockGoodBlock.GetBlocksmithPublicKey(),
				mockGoodBlock.GetTotalAmount(),
				mockGoodBlock.GetTotalFee(),
				mockGoodBlock.GetTotalCoinBase(),
			),
		)
	}

	return db.QueryRow(qStr), nil
}

func (*mockNodeRegistrationServiceBlockPopSuccess) ResetScrambledNodes() {

}

func (*mockMempoolServiceBlockPopSuccess) GetMempoolTransactionsWantToBackup(
	height uint32,
) ([]*model.MempoolTransaction, error) {
	return make([]*model.MempoolTransaction, 0), nil
}

func (*mockMempoolServiceBlockPopFail) GetMempoolTransactionsWantToBackup(
	height uint32,
) ([]*model.MempoolTransaction, error) {
	return nil, errors.New("mockedError")
}

func (*mockReceiptSuccess) GetPublishedReceiptsByHeight(blockHeight uint32) ([]*model.PublishedReceipt, error) {
	return make([]*model.PublishedReceipt, 0), nil
}

func (*mockReceiptFail) GetPublishedReceiptsByHeight(blockHeight uint32) ([]*model.PublishedReceipt, error) {
	return nil, errors.New("mockError")
}

func (*mockExecutorBlockPopSuccess) BeginTx() error {
	return nil
}

func (*mockExecutorBlockPopSuccess) CommitTx() error {
	return nil
}

func (*mockExecutorBlockPopSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}
func (*mockExecutorBlockPopSuccess) RollbackTx() error {
	return nil
}
func (*mockExecutorBlockPopSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	transactionQ := query.NewTransactionQuery(&chaintype.MainChain{})
	blockQ := query.NewBlockQuery(&chaintype.MainChain{})
	switch qStr {
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version FROM main_block WHERE height = 999":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields).AddRow(
				mockGoodCommonBlock.GetID(),
				mockGoodCommonBlock.GetBlockHash(),
				mockGoodCommonBlock.GetPreviousBlockHash(),
				mockGoodCommonBlock.GetHeight(),
				mockGoodCommonBlock.GetTimestamp(),
				mockGoodCommonBlock.GetBlockSeed(),
				mockGoodCommonBlock.GetBlockSignature(),
				mockGoodCommonBlock.GetCumulativeDifficulty(),
				mockGoodCommonBlock.GetPayloadLength(),
				mockGoodCommonBlock.GetPayloadHash(),
				mockGoodCommonBlock.GetBlocksmithPublicKey(),
				mockGoodCommonBlock.GetTotalAmount(),
				mockGoodCommonBlock.GetTotalFee(),
				mockGoodCommonBlock.GetTotalCoinBase(),
				mockGoodCommonBlock.GetVersion(),
			),
		)
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version FROM main_block WHERE id = 0":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields).AddRow(
				mockGoodCommonBlock.GetID(),
				mockGoodCommonBlock.GetBlockHash(),
				mockGoodCommonBlock.GetPreviousBlockHash(),
				mockGoodCommonBlock.GetHeight(),
				mockGoodCommonBlock.GetTimestamp(),
				mockGoodCommonBlock.GetBlockSeed(),
				mockGoodCommonBlock.GetBlockSignature(),
				mockGoodCommonBlock.GetCumulativeDifficulty(),
				mockGoodCommonBlock.GetPayloadLength(),
				mockGoodCommonBlock.GetPayloadHash(),
				mockGoodCommonBlock.GetBlocksmithPublicKey(),
				mockGoodCommonBlock.GetTotalAmount(),
				mockGoodCommonBlock.GetTotalFee(),
				mockGoodCommonBlock.GetTotalCoinBase(),
				mockGoodCommonBlock.GetVersion(),
			),
		)
	case "SELECT id, block_id, block_height, sender_account_address, recipient_account_address, transaction_type, fee, " +
		"timestamp, transaction_hash, transaction_body_length, transaction_body_bytes, signature, version, " +
		"transaction_index FROM \"transaction\" WHERE block_id = ? ORDER BY transaction_index ASC":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(transactionQ.Fields))
	}

	return db.Query(qStr)
}

func (*mockExecutorBlockPopSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	blockQ := query.NewBlockQuery(&chaintype.MainChain{})

	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
		sqlmock.NewRows(blockQ.Fields).AddRow(
			mockGoodBlock.GetID(),
			mockGoodBlock.GetBlockHash(),
			mockGoodBlock.GetPreviousBlockHash(),
			mockGoodBlock.GetHeight(),
			mockGoodBlock.GetTimestamp(),
			mockGoodBlock.GetBlockSeed(),
			mockGoodBlock.GetBlockSignature(),
			mockGoodBlock.GetCumulativeDifficulty(),
			mockGoodBlock.GetPayloadLength(),
			mockGoodBlock.GetPayloadHash(),
			mockGoodBlock.GetBlocksmithPublicKey(),
			mockGoodBlock.GetTotalAmount(),
			mockGoodBlock.GetTotalFee(),
			mockGoodBlock.GetTotalCoinBase(),
			mockGoodBlock.GetVersion(),
		),
	)
	return db.QueryRow(qStr), nil
}

func TestBlockService_PopOffToBlock(t *testing.T) {
	type fields struct {
		RWMutex                 sync.RWMutex
		Chaintype               chaintype.ChainType
		KVExecutor              kvdb.KVExecutorInterface
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		MerkleTreeQuery         query.MerkleTreeQueryInterface
		PublishedReceiptQuery   query.PublishedReceiptQueryInterface
		SkippedBlocksmithQuery  query.SkippedBlocksmithQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		ReceiptService          ReceiptServiceInterface
		NodeRegistrationService NodeRegistrationServiceInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		Observer                *observer.Observer
		Logger                  *log.Logger
	}
	type args struct {
		commonBlock *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Block
		wantErr bool
	}{
		{
			name: "Fail - GetLastBlock",
			fields: fields{
				Chaintype:               &chaintype.MainChain{},
				KVExecutor:              nil,
				QueryExecutor:           &mockExecutorBlockPopGetLastBlockFail{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.MainChain{}),
				MerkleTreeQuery:         nil,
				PublishedReceiptQuery:   nil,
				SkippedBlocksmithQuery:  nil,
				Signature:               nil,
				MempoolService:          &mockMempoolServiceBlockPopSuccess{},
				ReceiptService:          &mockReceiptSuccess{},
				NodeRegistrationService: &mockNodeRegistrationServiceBlockPopSuccess{},
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				ParticipationScoreQuery: nil,
				NodeRegistrationQuery:   nil,
				Observer:                nil,
				Logger:                  logrus.New(),
			},
			args: args{
				commonBlock: mockGoodCommonBlock,
			},
			want:    make([]*model.Block, 0),
			wantErr: true,
		},
		{
			name: "Fail - HardFork",
			fields: fields{
				RWMutex:                 sync.RWMutex{},
				Chaintype:               &chaintype.MainChain{},
				KVExecutor:              nil,
				QueryExecutor:           &mockExecutorBlockPopSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.MainChain{}),
				MerkleTreeQuery:         nil,
				PublishedReceiptQuery:   nil,
				SkippedBlocksmithQuery:  nil,
				Signature:               nil,
				MempoolService:          &mockMempoolServiceBlockPopSuccess{},
				ReceiptService:          &mockReceiptSuccess{},
				NodeRegistrationService: &mockNodeRegistrationServiceBlockPopSuccess{},
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				ParticipationScoreQuery: nil,
				NodeRegistrationQuery:   nil,
				Observer:                nil,
				Logger:                  logrus.New(),
			},
			args: args{
				commonBlock: mockBadCommonBlockHardFork,
			},
			want:    make([]*model.Block, 0),
			wantErr: false,
		},
		{
			name: "Fail - CommonBlockNotFound",
			fields: fields{
				Chaintype:               &chaintype.MainChain{},
				KVExecutor:              nil,
				QueryExecutor:           &mockExecutorBlockPopFailCommonNotFound{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.MainChain{}),
				MerkleTreeQuery:         nil,
				PublishedReceiptQuery:   nil,
				SkippedBlocksmithQuery:  nil,
				Signature:               nil,
				MempoolService:          &mockMempoolServiceBlockPopSuccess{},
				ReceiptService:          &mockReceiptSuccess{},
				NodeRegistrationService: &mockNodeRegistrationServiceBlockPopSuccess{},
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				ParticipationScoreQuery: nil,
				NodeRegistrationQuery:   nil,
				Observer:                nil,
				Logger:                  logrus.New(),
			},
			args: args{
				commonBlock: mockGoodCommonBlock,
			},
			want:    make([]*model.Block, 0),
			wantErr: true,
		},
		{
			name: "Fail - GetPublishedReceiptError",
			fields: fields{
				RWMutex:                 sync.RWMutex{},
				Chaintype:               &chaintype.MainChain{},
				KVExecutor:              nil,
				QueryExecutor:           &mockExecutorBlockPopSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.MainChain{}),
				MerkleTreeQuery:         nil,
				PublishedReceiptQuery:   nil,
				SkippedBlocksmithQuery:  nil,
				Signature:               nil,
				MempoolService:          &mockMempoolServiceBlockPopSuccess{},
				ReceiptService:          &mockReceiptSuccess{},
				NodeRegistrationService: &mockNodeRegistrationServiceBlockPopSuccess{},
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				ParticipationScoreQuery: nil,
				NodeRegistrationQuery:   nil,
				Observer:                nil,
				Logger:                  logrus.New(),
			},
			args: args{
				commonBlock: mockGoodCommonBlock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Fail - GetMempoolToBackupFail",
			fields: fields{
				RWMutex:                 sync.RWMutex{},
				Chaintype:               &chaintype.MainChain{},
				KVExecutor:              nil,
				QueryExecutor:           &mockExecutorBlockPopSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.MainChain{}),
				MerkleTreeQuery:         nil,
				PublishedReceiptQuery:   nil,
				SkippedBlocksmithQuery:  nil,
				Signature:               nil,
				MempoolService:          &mockMempoolServiceBlockPopFail{},
				ReceiptService:          &mockReceiptSuccess{},
				NodeRegistrationService: &mockNodeRegistrationServiceBlockPopSuccess{},
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				ParticipationScoreQuery: nil,
				NodeRegistrationQuery:   nil,
				Observer:                nil,
				Logger:                  logrus.New(),
			},
			args: args{
				commonBlock: mockGoodCommonBlock,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Success",
			fields: fields{
				RWMutex:                 sync.RWMutex{},
				Chaintype:               &chaintype.MainChain{},
				KVExecutor:              nil,
				QueryExecutor:           &mockExecutorBlockPopSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.MainChain{}),
				MerkleTreeQuery:         nil,
				PublishedReceiptQuery:   nil,
				SkippedBlocksmithQuery:  nil,
				Signature:               nil,
				MempoolService:          &mockMempoolServiceBlockPopSuccess{},
				ReceiptService:          &mockReceiptSuccess{},
				NodeRegistrationService: &mockNodeRegistrationServiceBlockPopSuccess{},
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				ParticipationScoreQuery: nil,
				NodeRegistrationQuery:   nil,
				Observer:                nil,
				Logger:                  logrus.New(),
			},
			args: args{
				commonBlock: mockGoodCommonBlock,
			},
			want:    nil,
			wantErr: false,
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
				PublishedReceiptQuery:   tt.fields.PublishedReceiptQuery,
				SkippedBlocksmithQuery:  tt.fields.SkippedBlocksmithQuery,
				Signature:               tt.fields.Signature,
				MempoolService:          tt.fields.MempoolService,
				ReceiptService:          tt.fields.ReceiptService,
				NodeRegistrationService: tt.fields.NodeRegistrationService,
				ActionTypeSwitcher:      tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				Observer:                tt.fields.Observer,
				Logger:                  tt.fields.Logger,
			}
			got, err := bs.PopOffToBlock(tt.args.commonBlock)
			if (err != nil) != tt.wantErr {
				t.Errorf("PopOffToBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PopOffToBlock() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockMainExecutorPopulateBlockDataFail struct {
		query.Executor
	}
	mockMainExecutorPopulateBlockDataSuccess struct {
		query.Executor
	}
)

func (*mockMainExecutorPopulateBlockDataFail) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("Mock Error")
}

func (*mockMainExecutorPopulateBlockDataSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockMain, _ := sqlmock.New()
	defer db.Close()
	switch qStr {
	case "SELECT id, block_id, block_height, sender_account_address, recipient_account_address, transaction_type, " +
		"fee, timestamp, transaction_hash, transaction_body_length, transaction_body_bytes, signature, version, " +
		"transaction_index FROM \"transaction\" WHERE block_id = ? ORDER BY transaction_index ASC":
		mockMain.ExpectQuery(regexp.QuoteMeta(qStr)).
			WillReturnRows(sqlmock.NewRows(
				query.NewTransactionQuery(&chaintype.MainChain{}).Fields,
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
				mockTransaction.TransactionIndex,
			))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
		"reference_block_hash, rmr_linked, recipient_signature, intermediate_hashes, block_height, receipt_index, " +
		"published_index FROM published_receipt WHERE block_height = ? ORDER BY published_index ASC":
		mockMain.ExpectQuery(regexp.QuoteMeta(qStr)).
			WillReturnRows(sqlmock.NewRows(
				query.NewPublishedReceiptQuery().Fields,
			).AddRow(
				mockPublishedReceipt[0].BatchReceipt.SenderPublicKey,
				mockPublishedReceipt[0].BatchReceipt.RecipientPublicKey,
				mockPublishedReceipt[0].BatchReceipt.DatumType,
				mockPublishedReceipt[0].BatchReceipt.DatumHash,
				mockPublishedReceipt[0].BatchReceipt.ReferenceBlockHeight,
				mockPublishedReceipt[0].BatchReceipt.ReferenceBlockHash,
				mockPublishedReceipt[0].BatchReceipt.RMRLinked,
				mockPublishedReceipt[0].BatchReceipt.RecipientSignature,
				mockPublishedReceipt[0].IntermediateHashes,
				mockPublishedReceipt[0].BlockHeight,
				mockPublishedReceipt[0].ReceiptIndex,
				mockPublishedReceipt[0].PublishedIndex,
			))

	}
	rows, _ := db.Query(qStr)
	return rows, nil
}

func TestBlockMainService_PopulateBlockData(t *testing.T) {
	type fields struct {
		Chaintype               chaintype.ChainType
		KVExecutor              kvdb.KVExecutorInterface
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		MerkleTreeQuery         query.MerkleTreeQueryInterface
		PublishedReceiptQuery   query.PublishedReceiptQueryInterface
		SkippedBlocksmithQuery  query.SkippedBlocksmithQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		ReceiptService          ReceiptServiceInterface
		NodeRegistrationService NodeRegistrationServiceInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		Observer                *observer.Observer
		Logger                  *log.Logger
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		expects *model.Block
	}{
		{
			name: "PopulateBlockData:fail-{dbErr}",
			fields: fields{
				Chaintype:             &chaintype.SpineChain{},
				QueryExecutor:         &mockMainExecutorPopulateBlockDataFail{},
				TransactionQuery:      query.NewTransactionQuery(&chaintype.MainChain{}),
				PublishedReceiptQuery: query.NewPublishedReceiptQuery(),
				Logger:                logrus.New(),
			},
			args: args{
				block: &model.Block{},
			},
			wantErr: true,
		},
		{
			name: "PopulateBlockData:success",
			fields: fields{
				Chaintype:             &chaintype.SpineChain{},
				QueryExecutor:         &mockMainExecutorPopulateBlockDataSuccess{},
				TransactionQuery:      query.NewTransactionQuery(&chaintype.MainChain{}),
				PublishedReceiptQuery: query.NewPublishedReceiptQuery(),
				Logger:                logrus.New(),
			},
			args: args{
				block: &model.Block{
					ID: int64(-1701929749060110283),
				},
			},
			wantErr: false,
			expects: &model.Block{
				ID: int64(-1701929749060110283),
				Transactions: []*model.Transaction{
					mockTransaction,
				},
				PublishedReceipts: []*model.PublishedReceipt{
					mockPublishedReceipt[0],
				},
			},
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
				PublishedReceiptQuery:   tt.fields.PublishedReceiptQuery,
				SkippedBlocksmithQuery:  tt.fields.SkippedBlocksmithQuery,
				Signature:               tt.fields.Signature,
				MempoolService:          tt.fields.MempoolService,
				ReceiptService:          tt.fields.ReceiptService,
				NodeRegistrationService: tt.fields.NodeRegistrationService,
				ActionTypeSwitcher:      tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				Observer:                tt.fields.Observer,
				Logger:                  tt.fields.Logger,
			}
			if err := bs.PopulateBlockData(tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("BlockMainService.PopulateBlockData() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.expects != nil && !reflect.DeepEqual(tt.args.block, tt.expects) {
				t.Errorf("BlockMainService.PopulateBlockData() = %v, want %v", tt.expects, tt.args.block)
			}
		})
	}
}
