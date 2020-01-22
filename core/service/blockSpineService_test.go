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
	mockSpineBlockData = model.Block{
		ID:        -1701929749060110283,
		BlockHash: make([]byte, 32),
		PreviousBlockHash: []byte{204, 131, 181, 204, 170, 112, 249, 115, 172, 193, 120, 7, 166, 200, 160, 138, 32, 0, 163, 161,
			45, 128, 173, 123, 252, 203, 199, 224, 249, 124, 168, 41},
		Height:    1,
		Timestamp: 1,
		BlockSeed: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		BlockSignature:       []byte{144, 246, 37, 144, 213, 135},
		CumulativeDifficulty: "1000",
		BlocksmithPublicKey: []byte{1, 2, 3, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		TotalAmount:   0,
		TotalFee:      0,
		TotalCoinBase: 0,
		Version:       0,
		PayloadLength: 1,
		PayloadHash:   []byte{},
	}
	mockSpinePublicKey = &model.SpinePublicKey{
		NodePublicKey:   nrsNodePubKey1,
		MainBlockHeight: 1,
		PublicKeyAction: model.SpinePublicKeyAction_AddKey,
		Height:          1,
		Latest:          true,
	}
)

type (
	mockSpineSignature struct {
		crypto.Signature
	}
	mockSpineSignatureFail struct {
		crypto.Signature
	}
	mockSpineQueryExecutorSuccess struct {
		query.Executor
	}
	mockSpineQueryExecuteNotNil struct {
		query.Executor
	}
	mockSpineQueryExecutorScanFail struct {
		query.Executor
	}
	mockSpineQueryExecutorFail struct {
		query.Executor
	}
	mockSpineQueryExecutorNotFound struct {
		query.Executor
	}
	mockSpineTypeAction struct {
		transaction.SendMoney
	}
	mockSpineTypeActionSuccess struct {
		mockSpineTypeAction
	}

	mockSpineKVExecutorSuccess struct {
		kvdb.KVExecutor
	}

	mockSpineKVExecutorSuccessKeyNotFound struct {
		mockSpineKVExecutorSuccess
	}

	mockSpineKVExecutorFailOtherError struct {
		mockSpineKVExecutorSuccess
	}

	mockSpineNodeRegistrationServiceSuccess struct {
		NodeRegistrationService
	}

	mockSpineNodeRegistrationServiceFail struct {
		NodeRegistrationService
	}
)

func (*mockSpineNodeRegistrationServiceSuccess) AddParticipationScore(
	nodeID, scoreDelta int64,
	height uint32,
	tx bool,
) (newScore int64, err error) {
	return 100000, nil
}

func (*mockSpineNodeRegistrationServiceSuccess) SelectNodesToBeAdmitted(limit uint32) ([]*model.NodeRegistration, error) {
	return []*model.NodeRegistration{
		{
			AccountAddress: "TESTADMITTED",
		},
	}, nil
}

func (*mockSpineNodeRegistrationServiceSuccess) AdmitNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	return nil
}

func (*mockSpineNodeRegistrationServiceSuccess) SelectNodesToBeExpelled() ([]*model.NodeRegistration, error) {
	return []*model.NodeRegistration{
		{
			AccountAddress: "TESTEXPELLED",
		},
	}, nil
}

func (*mockSpineNodeRegistrationServiceFail) AddParticipationScore(
	nodeID, scoreDelta int64,
	height uint32,
	tx bool,
) (newScore int64, err error) {
	return 100000, nil
}

func (*mockSpineNodeRegistrationServiceFail) SelectNodesToBeExpelled() ([]*model.NodeRegistration, error) {
	return []*model.NodeRegistration{
		{
			AccountAddress: "TESTEXPELLED",
		},
	}, nil
}
func (*mockSpineNodeRegistrationServiceFail) ExpelNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	return nil
}
func (*mockSpineNodeRegistrationServiceSuccess) GetNodeAdmittanceCycle() uint32 {
	return 1
}

func (*mockSpineNodeRegistrationServiceSuccess) ExpelNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	return nil
}

func (*mockSpineNodeRegistrationServiceSuccess) BuildScrambledNodes(block *model.Block) error {
	return nil
}

func (*mockSpineNodeRegistrationServiceSuccess) GetBlockHeightToBuildScrambleNodes(lastBlockHeight uint32) uint32 {
	return lastBlockHeight
}

func (*mockSpineNodeRegistrationServiceFail) BuildScrambledNodes(block *model.Block) error {
	return errors.New("mockSpine Error")
}

func (*mockSpineNodeRegistrationServiceFail) GetBlockHeightToBuildScrambleNodes(lastBlockHeight uint32) uint32 {
	return lastBlockHeight
}

func (*mockSpineKVExecutorSuccess) Get(key string) ([]byte, error) {
	return nil, nil
}

func (*mockSpineKVExecutorSuccess) Insert(key string, value []byte, expiry int) error {
	return nil
}

func (*mockSpineKVExecutorSuccessKeyNotFound) Get(key string) ([]byte, error) {
	return nil, badger.ErrKeyNotFound
}

func (*mockSpineKVExecutorFailOtherError) Get(key string) ([]byte, error) {
	return nil, badger.ErrInvalidKey
}

func (*mockSpineKVExecutorFailOtherError) Insert(key string, value []byte, expiry int) error {
	return badger.ErrInvalidKey
}

// mockSpineTypeAction
func (*mockSpineTypeAction) ApplyConfirmed(int64) error {
	return nil
}
func (*mockSpineTypeAction) Validate(bool) error {
	return nil
}
func (*mockSpineTypeAction) GetAmount() int64 {
	return 10
}
func (*mockSpineTypeActionSuccess) GetTransactionType(tx *model.Transaction) (transaction.TypeAction, error) {
	return &mockSpineTypeAction{}, nil
}

// mockSpineSignature
func (*mockSpineSignature) SignByNode(payload []byte, nodeSeed string) []byte {
	return []byte{}
}

func (*mockSpineSignature) VerifyNodeSignature(
	payload, signature, nodePublicKey []byte,
) bool {
	return true
}

func (*mockSpineSignatureFail) VerifyNodeSignature(
	payload, signature, nodePublicKey []byte,
) bool {
	return false
}

// mockSpineQueryExecutorScanFail
func (*mockSpineQueryExecutorScanFail) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()
	mockSpine.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{
		"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
		"PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase"}))
	rows, _ := db.Query(qe)
	return rows, nil
}

// mockSpineQueryExecutorNotFound
func (*mockSpineQueryExecutorNotFound) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"registration_status, latest, height  FROM node_registry WHERE node_public_key = ? AND height <= ? " +
		"ORDER BY height DESC LIMIT 1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
			"Version"},
		))
	default:
		return nil, errors.New("mockSpineQueryExecutorNotFound - InvalidQuery")
	}
	rows, _ := db.Query(qe)
	return rows, nil
}

// mockSpineQueryExecutorNotNil
func (*mockSpineQueryExecuteNotNil) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	mockSpine.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"ID"}))
	return db.Query("")
}

// mockSpineQueryExecutorFail
func (*mockSpineQueryExecutorFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}
func (*mockSpineQueryExecutorFail) ExecuteStatement(qe string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("MockedError")
}
func (*mockSpineQueryExecutorFail) BeginTx() error { return nil }

func (*mockSpineQueryExecutorFail) RollbackTx() error { return nil }

func (*mockSpineQueryExecutorFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("mockSpineError:deleteMempoolFail")
}
func (*mockSpineQueryExecutorFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mockSpine, _ := sqlmock.New()
	mockSpineRows := mockSpine.NewRows([]string{"fake"})
	mockSpineRows.AddRow("1")
	mockSpine.ExpectQuery(qStr).WillReturnRows(mockSpineRows)
	return db.QueryRow(qStr), nil
}
func (*mockSpineQueryExecutorFail) CommitTx() error { return errors.New("mockSpineError:commitFail") }

// mockSpineQueryExecutorSuccess
func (*mockSpineQueryExecutorSuccess) BeginTx() error { return nil }

func (*mockSpineQueryExecutorSuccess) RollbackTx() error { return nil }

func (*mockSpineQueryExecutorSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}
func (*mockSpineQueryExecutorSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}
func (*mockSpineQueryExecutorSuccess) CommitTx() error { return nil }

func (*mockSpineQueryExecutorSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()

	switch qStr {
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "NodePublicKey", "AccountAddress", "RegistrationHeight", "NodeAddress", "LockedBalance", "RegistrationStatus",
			"Latest", "Height",
		}).AddRow(1, bcsNodePubKey1, bcsAddress1, 10, "10.10.10.1", 100000000, uint32(model.NodeRegistrationState_NodeQueued), true, 100))
	case "SELECT id, block_height, tree, timestamp FROM merkle_tree ORDER BY timestamp DESC LIMIT 1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "BlockHeight", "Tree", "Timestamp",
		}))
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version " +
		"FROM spine_block ORDER BY height DESC LIMIT 1":
		mockSpineRows := mockSpine.NewRows(query.NewBlockQuery(&chaintype.SpineChain{}).Fields)
		mockSpineRows.AddRow(
			mockSpineBlockData.GetID(),
			mockSpineBlockData.GetBlockHash(),
			mockSpineBlockData.GetPreviousBlockHash(),
			mockSpineBlockData.GetHeight(),
			mockSpineBlockData.GetTimestamp(),
			mockSpineBlockData.GetBlockSeed(),
			mockSpineBlockData.GetBlockSignature(),
			mockSpineBlockData.GetCumulativeDifficulty(),
			mockSpineBlockData.GetPayloadLength(),
			mockSpineBlockData.GetPayloadHash(),
			mockSpineBlockData.GetBlocksmithPublicKey(),
			mockSpineBlockData.GetTotalAmount(),
			mockSpineBlockData.GetTotalFee(),
			mockSpineBlockData.GetTotalCoinBase(),
			mockSpineBlockData.GetVersion(),
		)
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(mockSpineRows)
	default:
		return nil, fmt.Errorf("unmoked query for mockSpineQueryExecutorSuccess selectrow %s", qStr)
	}
	row := db.QueryRow(qStr)
	return row, nil
}

func (*mockSpineQueryExecutorSuccess) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE id = ? AND latest=1":
		for idx, arg := range args {
			if idx == 0 {
				nodeID := fmt.Sprintf("%d", arg)
				switch nodeID {
				case "1":
					mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "node_address", "locked_balance", "registration_status", "latest", "height",
					}).AddRow(1, bcsNodePubKey1, bcsAddress1, 10, "10.10.10.1", 100000000, uint32(model.NodeRegistrationState_NodeRegistered), true, 100))
				case "2":
					mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "node_address", "locked_balance", "registration_status", "latest", "height",
					}).AddRow(2, bcsNodePubKey2, bcsAddress2, 20, "10.10.10.2", 100000000, uint32(model.NodeRegistrationState_NodeRegistered), true, 200))
				case "3":
					mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "node_address", "locked_balance", "registration_status", "latest", "height",
					}).AddRow(3, bcsNodePubKey3, bcsAddress3, 30, "10.10.10.3", 100000000, uint32(model.NodeRegistrationState_NodeRegistered), true, 300))
				case "4":
					mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "node_address", "locked_balance", "registration_status", "latest", "height",
					}).AddRow(3, mockSpineGoodBlock.BlocksmithPublicKey, bcsAddress3, 30, "10.10.10.3", 100000000,
						uint32(model.NodeRegistrationState_NodeRegistered), true, 300))
				}
			}
		}
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE height >= (SELECT MIN(height) " +
		"FROM main_block AS mb1 WHERE mb1.timestamp >= 12345600) AND height <= (SELECT MAX(height) " +
		"FROM main_block AS mb2 WHERE mb2.timestamp < 12345678) AND registration_status != 1 AND latest=1 ORDER BY height":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(query.NewNodeRegistrationQuery().Fields))
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE height >= (SELECT MIN(height) " +
		"FROM main_block AS mb1 WHERE mb1.timestamp >= 0) AND height <= (SELECT MAX(height) " +
		"FROM main_block AS mb2 WHERE mb2.timestamp < 12345678) AND registration_status != 1 AND latest=1 ORDER BY height":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(query.NewNodeRegistrationQuery().Fields))
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND height <= ? " +
		"ORDER BY height DESC LIMIT 1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
			"account_address", "registration_height", "node_address", "locked_balance", "registration_status", "latest", "height",
		}).AddRow(1, bcsNodePubKey1, bcsAddress1, 10, "10.10.10.10", 100000000, uint32(model.NodeRegistrationState_NodeQueued), true, 100))
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version FROM spine_block WHERE height = 0":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "BlockHash", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
			"Version"},
		).AddRow(1, []byte{}, []byte{}, 1, 10000, []byte{}, []byte{}, "", 2, []byte{}, bcsNodePubKey1, 0, 0, 0, 1))
	case "SELECT A.node_id, A.score, A.latest, A.height FROM participation_score as A INNER JOIN node_registry as B " +
		"ON A.node_id = B.id WHERE B.node_public_key=? AND B.latest=1 AND B.registration_status=0 AND A.latest=1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"node_id",
			"score",
			"latest",
			"height",
		},
		).AddRow(-1, 100000, true, 0))
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version FROM spine_block ORDER BY " +
		"height DESC LIMIT 1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).
			WillReturnRows(sqlmock.NewRows(
				query.NewBlockQuery(&chaintype.SpineChain{}).Fields,
			).AddRow(
				mockSpineBlockData.GetID(),
				mockSpineBlockData.GetBlockHash(),
				mockSpineBlockData.GetPreviousBlockHash(),
				mockSpineBlockData.GetHeight(),
				mockSpineBlockData.GetTimestamp(),
				mockSpineBlockData.GetBlockSeed(),
				mockSpineBlockData.GetBlockSignature(),
				mockSpineBlockData.GetCumulativeDifficulty(),
				mockSpineBlockData.GetPayloadLength(),
				mockSpineBlockData.GetPayloadHash(),
				mockSpineBlockData.GetBlocksmithPublicKey(),
				mockSpineBlockData.GetTotalAmount(),
				mockSpineBlockData.GetTotalFee(),
				mockSpineBlockData.GetTotalCoinBase(),
				mockSpineBlockData.GetVersion(),
			))
	case "SELECT node_public_key, public_key_action, main_block_height, latest, height FROM spine_public_key WHERE height = 1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).
			WillReturnRows(sqlmock.NewRows(
				query.NewSpinePublicKeyQuery().Fields,
			).AddRow(
				mockSpinePublicKey.NodePublicKey,
				mockSpinePublicKey.PublicKeyAction,
				mockSpinePublicKey.MainBlockHeight,
				mockSpinePublicKey.Latest,
				mockSpinePublicKey.Height,
			))
	case "SELECT id, fee_per_byte, arrival_timestamp, transaction_bytes, sender_account_address, recipient_account_address " +
		"FROM mempool WHERE id = :id":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "FeePerByte", "ArrivalTimestamp", "TransactionBytes", "SenderAccountAddress", "RecipientAccountAddress",
		}))
	case "SELECT nr.id AS nodeID, nr.node_public_key AS node_public_key, ps.score AS participation_score FROM node_registry " +
		"AS nr INNER JOIN participation_score AS ps ON nr.id = ps.node_id WHERE nr.registration_status = 0 AND nr.latest " +
		"= 1 AND ps.score > 0 AND ps.latest = 1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"node_id", "node_public_key", "score",
		}).AddRow(
			mockSpineBlocksmiths[0].NodeID,
			mockSpineBlocksmiths[0].NodePublicKey,
			"1000",
		).AddRow(
			mockSpineBlocksmiths[1].NodeID,
			mockSpineBlocksmiths[1].NodePublicKey,
			"1000",
		))
	case "SELECT blocksmith_public_key, pop_change, block_height, blocksmith_index FROM skipped_blocksmith WHERE block_height = 0":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"blocksmith_public_key", "pop_change", "block_height", "blocksmith_index",
		}).AddRow(
			mockSpineBlocksmiths[0].NodePublicKey,
			5000,
			mockSpinePublishedReceipt[0].BlockHeight,
			0,
		))
	case "SELECT blocksmith_public_key, pop_change, block_height, blocksmith_index FROM skipped_blocksmith WHERE block_height = 1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"blocksmith_public_key", "pop_change", "block_height", "blocksmith_index",
		}).AddRow(
			mockSpineBlocksmiths[0].NodePublicKey,
			5000,
			mockSpinePublishedReceipt[0].BlockHeight,
			0,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
		"reference_block_hash, rmr_linked, recipient_signature, intermediate_hashes, block_height, receipt_index, " +
		"published_index FROM published_receipt WHERE block_height = ? ORDER BY published_index ASC":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"sender_public_key", "recipient_public_key", "datum_type", "datum_hash", "reference_block_height",
			"reference_block_hash", "rmr_linked", "recipient_signature", "intermediate_hashes", "block_height",
			"receipt_index", "published_index",
		}).AddRow(
			mockSpinePublishedReceipt[0].BatchReceipt.SenderPublicKey,
			mockSpinePublishedReceipt[0].BatchReceipt.RecipientPublicKey,
			mockSpinePublishedReceipt[0].BatchReceipt.DatumType,
			mockSpinePublishedReceipt[0].BatchReceipt.DatumHash,
			mockSpinePublishedReceipt[0].BatchReceipt.ReferenceBlockHeight,
			mockSpinePublishedReceipt[0].BatchReceipt.ReferenceBlockHash,
			mockSpinePublishedReceipt[0].BatchReceipt.RMRLinked,
			mockSpinePublishedReceipt[0].BatchReceipt.RecipientSignature,
			mockSpinePublishedReceipt[0].IntermediateHashes,
			mockSpinePublishedReceipt[0].BlockHeight,
			mockSpinePublishedReceipt[0].ReceiptIndex,
			mockSpinePublishedReceipt[0].PublishedIndex,
		))
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"registration_status, latest, height, max(height) AS max_height FROM node_registry where height <= 0 AND " +
		"registration_status = 0 GROUP BY id ORDER BY height DESC":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id", "node_public_key", "account_address", "registration_height", "node_address", "locked_balance",
			"registration_status", "latest", "height",
		}))
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"registration_status, latest, height, max(height) AS max_height FROM node_registry where height <= 1 " +
		"AND registration_status = 0 GROUP BY id ORDER BY height DESC":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id", "node_public_key", "account_address", "registration_height", "node_address", "locked_balance",
			"registration_status", "latest", "height",
		}))
	}
	rows, _ := db.Query(qe)
	return rows, nil
}

var mockSpinePublishedReceipt = []*model.PublishedReceipt{
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

func (*mockSpineQueryExecutorSuccess) ExecuteStatement(qe string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func TestBlockSpineService_NewSpineBlock(t *testing.T) {
	var (
		mockSpineBlock = &model.Block{
			Version:             1,
			PreviousBlockHash:   []byte{},
			BlockSeed:           []byte{},
			BlocksmithPublicKey: bcsNodePubKey1,
			Timestamp:           15875392,
			SpinePublicKeys:     []*model.SpinePublicKey{},
			PayloadHash:         []byte{},
			PayloadLength:       0,
			BlockSignature:      []byte{},
		}
		mockSpineBlockHash, _ = util.GetBlockHash(mockSpineBlock, &chaintype.SpineChain{})
	)
	mockSpineBlock.BlockHash = mockSpineBlockHash

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
		spinePublicKeys     []*model.SpinePublicKey
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
				Chaintype: &chaintype.SpineChain{},
				Signature: &mockSpineSignature{},
			},
			args: args{
				version:             1,
				previousBlockHash:   []byte{},
				blockSeed:           []byte{},
				blockSmithPublicKey: bcsNodePubKey1,
				previousBlockHeight: 0,
				timestamp:           15875392,
				spinePublicKeys:     []*model.SpinePublicKey{},
				payloadHash:         []byte{},
				payloadLength:       0,
				secretPhrase:        "secretphrase",
			},
			want: mockSpineBlock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:     tt.fields.Chaintype,
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
				Signature:     tt.fields.Signature,
			}
			got, err := bs.NewSpineBlock(
				tt.args.version,
				tt.args.previousBlockHash,
				tt.args.blockSeed,
				tt.args.blockSmithPublicKey,
				tt.args.previousBlockHeight,
				tt.args.timestamp,
				tt.args.spinePublicKeys,
				tt.args.payloadHash,
				tt.args.payloadLength,
				tt.args.secretPhrase,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineService.NewBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockSpineService.NewBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockSpineService_NewGenesisBlock(t *testing.T) {
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
				Chaintype: &chaintype.SpineChain{},
				Signature: &mockSpineSignature{},
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
				BlockHash: []byte{222, 81, 44, 228, 147, 156, 145, 104, 1, 97, 62, 138, 253, 90, 55, 41, 29, 150, 230, 196,
					68, 216, 14, 244, 224, 161, 132, 157, 229, 68, 33, 147},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:     tt.fields.Chaintype,
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
				Signature:     tt.fields.Signature,
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
				t.Errorf("BlockSpineService.NewGenesisBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	mockSpineBlocksmiths = []*model.Blocksmith{
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
			NodePublicKey: mockSpineBlockData.BlocksmithPublicKey,
			NodeID:        4,
			NodeOrder:     new(big.Int).SetInt64(3000),
			Score:         new(big.Int).SetInt64(3000),
		},
	}
)

type (
	mockSpineBlocksmithServicePushBlock struct {
		strategy.BlocksmithStrategyMain
	}
)

func (*mockSpineBlocksmithServicePushBlock) GetSortedBlocksmiths(*model.Block) []*model.Blocksmith {
	return mockSpineBlocksmiths
}
func (*mockSpineBlocksmithServicePushBlock) GetSortedBlocksmithsMap(*model.Block) map[string]*int64 {
	var result = make(map[string]*int64)
	for index, mockSpine := range mockSpineBlocksmiths {
		mockSpineIndex := int64(index)
		result[string(mockSpine.NodePublicKey)] = &mockSpineIndex
	}
	return result
}
func (*mockSpineBlocksmithServicePushBlock) SortBlocksmiths(block *model.Block) {
}
func (*mockSpineBlocksmithServicePushBlock) GetSmithTime(blocksmithIndex int64, previousBlock *model.Block) int64 {
	return 0
}
func TestBlockSpineService_PushBlock(t *testing.T) {
	type fields struct {
		Chaintype               chaintype.ChainType
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		AccountBalanceQuery     query.AccountBalanceQueryInterface
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
				Chaintype:               &chaintype.SpineChain{},
				QueryExecutor:           &mockSpineQueryExecutorSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.SpineChain{}),
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				Observer:                observer.NewObserver(),
				MempoolQuery:            query.NewMempoolQuery(&chaintype.SpineChain{}),
				SkippedBlocksmithQuery:  query.NewSkippedBlocksmithQuery(),
				NodeRegistrationService: &mockSpineNodeRegistrationServiceSuccess{},
				BlocksmithStrategy:      &mockSpineBlocksmithServicePushBlock{},
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
				Chaintype:               &chaintype.SpineChain{},
				QueryExecutor:           &mockSpineQueryExecutorSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.SpineChain{}),
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				NodeRegistrationService: &mockSpineNodeRegistrationServiceSuccess{},
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				MempoolQuery:            query.NewMempoolQuery(&chaintype.SpineChain{}),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				SkippedBlocksmithQuery:  query.NewSkippedBlocksmithQuery(),
				Observer:                observer.NewObserver(),
				BlocksmithStrategy:      &mockSpineBlocksmithServicePushBlock{},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:          tt.fields.Chaintype,
				QueryExecutor:      tt.fields.QueryExecutor,
				BlockQuery:         tt.fields.BlockQuery,
				Signature:          tt.fields.Signature,
				Observer:           tt.fields.Observer,
				Logger:             logrus.New(),
				BlocksmithStrategy: tt.fields.BlocksmithStrategy,
			}
			if err := bs.PushBlock(tt.args.previousBlock, tt.args.block, tt.args.broadcast, true); (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineService.PushBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockSpineService_GetLastBlock(t *testing.T) {
	var (
		mockSpineBlockGetLastBlock = mockSpineBlockData
	)
	mockSpineBlockGetLastBlock.SpinePublicKeys = []*model.SpinePublicKey{
		mockSpinePublicKey,
	}

	type fields struct {
		Chaintype           chaintype.ChainType
		QueryExecutor       query.ExecutorInterface
		BlockQuery          query.BlockQueryInterface
		MempoolQuery        query.MempoolQueryInterface
		TransactionQuery    query.TransactionQueryInterface
		SpinePublicKeyQuery query.SpinePublicKeyQueryInterface
		Signature           crypto.SignatureInterface
		ActionTypeSwitcher  transaction.TypeActionSwitcher
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
				Chaintype:           &chaintype.SpineChain{},
				QueryExecutor:       &mockSpineQueryExecutorSuccess{},
				TransactionQuery:    query.NewTransactionQuery(&chaintype.SpineChain{}),
				SpinePublicKeyQuery: query.NewSpinePublicKeyQuery(),
				BlockQuery:          query.NewBlockQuery(&chaintype.SpineChain{}),
			},
			want:    &mockSpineBlockGetLastBlock,
			wantErr: false,
		},
		{
			name: "GetLastBlock:SelectFail",
			fields: fields{
				Chaintype:           &chaintype.SpineChain{},
				QueryExecutor:       &mockSpineQueryExecutorFail{},
				BlockQuery:          query.NewBlockQuery(&chaintype.SpineChain{}),
				SpinePublicKeyQuery: query.NewSpinePublicKeyQuery(),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:           tt.fields.Chaintype,
				QueryExecutor:       tt.fields.QueryExecutor,
				BlockQuery:          tt.fields.BlockQuery,
				SpinePublicKeyQuery: tt.fields.SpinePublicKeyQuery,
				Signature:           tt.fields.Signature,
			}
			got, err := bs.GetLastBlock()
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineService.GetLastBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockSpineService.GetLastBlock() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

type (
	mockSpineQueryExecutorGetGenesisBlockSuccess struct {
		query.Executor
	}

	mockSpineQueryExecutorGetGenesisBlockFail struct {
		query.Executor
	}
)

func (*mockSpineQueryExecutorGetGenesisBlockSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mockSpine, _ := sqlmock.New()
	mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.SpineChain{}).Fields,
		).AddRow(
			mockSpineBlockData.GetID(),
			mockSpineBlockData.GetBlockHash(),
			mockSpineBlockData.GetPreviousBlockHash(),
			mockSpineBlockData.GetHeight(),
			mockSpineBlockData.GetTimestamp(),
			mockSpineBlockData.GetBlockSeed(),
			mockSpineBlockData.GetBlockSignature(),
			mockSpineBlockData.GetCumulativeDifficulty(),
			mockSpineBlockData.GetPayloadLength(),
			mockSpineBlockData.GetPayloadHash(),
			mockSpineBlockData.GetBlocksmithPublicKey(),
			mockSpineBlockData.GetTotalAmount(),
			mockSpineBlockData.GetTotalFee(),
			mockSpineBlockData.GetTotalCoinBase(),
			mockSpineBlockData.GetVersion(),
		))
	return db.QueryRow(qStr), nil
}

func (*mockSpineQueryExecutorGetGenesisBlockFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, nil
}

func TestBlockSpineService_GetGenesisBlock(t *testing.T) {
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
				Chaintype:     &chaintype.SpineChain{},
				QueryExecutor: &mockSpineQueryExecutorGetGenesisBlockSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.SpineChain{}),
			},
			want:    &mockSpineBlockData,
			wantErr: false,
		},
		{
			name: "GetGenesis:fail",
			fields: fields{
				Chaintype:     &chaintype.SpineChain{},
				QueryExecutor: &mockSpineQueryExecutorGetGenesisBlockFail{},
				BlockQuery:    query.NewBlockQuery(&chaintype.SpineChain{}),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:     tt.fields.Chaintype,
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
				Signature:     tt.fields.Signature,
			}
			got, err := bs.GetGenesisBlock()
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineService.GetGenesisBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockSpineService.GetGenesisBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockSpineQueryExecutorGetBlocksSuccess struct {
		query.Executor
	}

	mockSpineQueryExecutorGetBlocksFail struct {
		query.Executor
	}
)

func (*mockSpineQueryExecutorGetBlocksSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	mockSpine.ExpectQuery(qStr).WillReturnRows(sqlmock.NewRows(
		query.NewBlockQuery(&chaintype.SpineChain{}).Fields,
	).AddRow(
		mockSpineBlockData.GetID(),
		mockSpineBlockData.GetBlockHash(),
		mockSpineBlockData.GetPreviousBlockHash(),
		mockSpineBlockData.GetHeight(),
		mockSpineBlockData.GetTimestamp(),
		mockSpineBlockData.GetBlockSeed(),
		mockSpineBlockData.GetBlockSignature(),
		mockSpineBlockData.GetCumulativeDifficulty(),
		mockSpineBlockData.GetPayloadLength(),
		mockSpineBlockData.GetPayloadHash(),
		mockSpineBlockData.GetBlocksmithPublicKey(),
		mockSpineBlockData.GetTotalAmount(),
		mockSpineBlockData.GetTotalFee(),
		mockSpineBlockData.GetTotalCoinBase(),
		mockSpineBlockData.GetVersion(),
	))
	return db.Query(qStr)
}

func (*mockSpineQueryExecutorGetBlocksFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}

func TestBlockSpineService_GetBlocks(t *testing.T) {
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
				Chaintype:     &chaintype.SpineChain{},
				QueryExecutor: &mockSpineQueryExecutorGetBlocksSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.SpineChain{}),
			},
			want: []*model.Block{
				&mockSpineBlockData,
			},
			wantErr: false,
		},
		{
			name: "GetBlocks:fail",
			fields: fields{
				Chaintype:     &chaintype.SpineChain{},
				QueryExecutor: &mockSpineQueryExecutorGetBlocksFail{},
				BlockQuery:    query.NewBlockQuery(&chaintype.SpineChain{}),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:     tt.fields.Chaintype,
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
				Signature:     tt.fields.Signature,
			}
			got, err := bs.GetBlocks()
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineService.GetBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockSpineService.GetBlocks() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockSpineMempoolServiceSelectFail struct {
		MempoolService
	}
	mockSpineMempoolServiceSelectWrongTransactionBytes struct {
		MempoolService
	}
	mockSpineMempoolServiceSelectSuccess struct {
		MempoolService
	}
	mockSpineQueryExecutorMempoolSuccess struct {
		query.Executor
	}
	mockSpineReceiptServiceReturnEmpty struct {
		ReceiptService
	}
)

func (*mockSpineReceiptServiceReturnEmpty) SelectReceipts(int64, uint32, uint32) ([]*model.PublishedReceipt, error) {
	return []*model.PublishedReceipt{}, nil
}

// mockSpineQueryExecutorMempoolSuccess
func (*mockSpineQueryExecutorMempoolSuccess) ExecuteSelect(string, bool, ...interface{}) (*sql.Rows, error) {
	db, mockSpine, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	mockSpine.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
		"id",
		"fee_per_byte",
		"arrival_timestamp",
		"transaction_bytes",
	}).AddRow(
		1,
		1,
		123456,
		transaction.GetFixturesForSignedMempoolTransaction(
			1,
			1562893305,
			"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			false,
		).TransactionBytes),
	)
	return db.Query("")
}

// mockSpineMempoolServiceSelectSuccess
func (*mockSpineMempoolServiceSelectSuccess) SelectTransactionFromMempool() ([]*model.MempoolTransaction, error) {
	return []*model.MempoolTransaction{
		{
			FeePerByte: 1,
			TransactionBytes: transaction.GetFixturesForSignedMempoolTransaction(
				1,
				1562893305,
				"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				false,
			).TransactionBytes,
		},
	}, nil
}

// mockSpineMempoolServiceSelectSuccess
func (*mockSpineMempoolServiceSelectSuccess) SelectTransactionsFromMempool(int64) ([]*model.Transaction, error) {
	txByte := transaction.GetFixturesForSignedMempoolTransaction(
		1,
		1562893305,
		"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		false,
	).TransactionBytes
	txHash := sha3.Sum256(txByte)
	return []*model.Transaction{
		{
			ID:              1,
			TransactionHash: txHash[:],
		},
	}, nil
}

// mockSpineMempoolServiceSelectFail
func (*mockSpineMempoolServiceSelectFail) SelectTransactionsFromMempool(int64) ([]*model.Transaction, error) {
	return nil, errors.New("want error on select")
}

// mockSpineMempoolServiceSelectSuccess
func (*mockSpineMempoolServiceSelectWrongTransactionBytes) SelectTransactionsFromMempool(int64) ([]*model.Transaction, error) {
	return []*model.Transaction{
		{
			ID: 1,
		},
	}, nil
}

func TestBlockSpineService_GenerateBlock(t *testing.T) {
	type fields struct {
		Chaintype             chaintype.ChainType
		QueryExecutor         query.ExecutorInterface
		BlockQuery            query.BlockQueryInterface
		MempoolQuery          query.MempoolQueryInterface
		TransactionQuery      query.TransactionQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		Signature             crypto.SignatureInterface
		MempoolService        MempoolServiceInterface
		ReceiptService        ReceiptServiceInterface
		BlocksmithStrategy    strategy.BlocksmithStrategyInterface
		ActionTypeSwitcher    transaction.TypeActionSwitcher
	}
	type args struct {
		previousBlock *model.Block
		secretPhrase  string
		timestamp     int64
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
				Chaintype:             &chaintype.SpineChain{},
				Signature:             &mockSpineSignature{},
				BlockQuery:            query.NewBlockQuery(&chaintype.SpineChain{}),
				MempoolQuery:          query.NewMempoolQuery(&chaintype.SpineChain{}),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &mockSpineQueryExecutorSuccess{},
				MempoolService: &mockSpineMempoolServiceSelectSuccess{
					MempoolService{
						QueryExecutor:      &mockSpineQueryExecutorMempoolSuccess{},
						ActionTypeSwitcher: &mockSpineTypeActionSuccess{},
					},
				},
				BlocksmithStrategy: &mockSpineBlocksmithServicePushBlock{},
				ReceiptService:     &mockSpineReceiptServiceReturnEmpty{},
				ActionTypeSwitcher: &mockSpineTypeActionSuccess{},
			},
			args: args{
				previousBlock: &model.Block{
					Version:             1,
					PreviousBlockHash:   []byte{},
					BlockSeed:           []byte{},
					BlocksmithPublicKey: bcsNodePubKey1,
					Timestamp:           12345600,
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
			bs := &BlockSpineService{
				Chaintype:             tt.fields.Chaintype,
				QueryExecutor:         tt.fields.QueryExecutor,
				BlockQuery:            tt.fields.BlockQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				Signature:             tt.fields.Signature,
				BlocksmithStrategy:    tt.fields.BlocksmithStrategy,
			}
			_, err := bs.GenerateBlock(
				tt.args.previousBlock,
				tt.args.secretPhrase,
				tt.args.timestamp,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineService.GenerateBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

type (
	mockSpineAddGenesisExecutor struct {
		query.Executor
	}
)

func (*mockSpineAddGenesisExecutor) BeginTx() error    { return nil }
func (*mockSpineAddGenesisExecutor) RollbackTx() error { return nil }
func (*mockSpineAddGenesisExecutor) CommitTx() error   { return nil }
func (*mockSpineAddGenesisExecutor) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}
func (*mockSpineAddGenesisExecutor) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}
func (*mockSpineAddGenesisExecutor) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()
	mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
		sqlmock.NewRows(query.NewMempoolQuery(chaintype.GetChainType(0)).Fields),
	)
	return db.Query(qStr)
}

type (
	mockSpineBlocksmithServiceAddGenesisSuccess struct {
		strategy.BlocksmithStrategyMain
	}
)

func (*mockSpineBlocksmithServiceAddGenesisSuccess) SortBlocksmiths(block *model.Block) {

}

func TestBlockSpineService_AddGenesis(t *testing.T) {
	type fields struct {
		Chaintype               chaintype.ChainType
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		SpinePublicKeyQuery     query.SpinePublicKeyQueryInterface
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
				Chaintype:               &chaintype.SpineChain{},
				Signature:               &mockSpineSignature{},
				MempoolQuery:            query.NewMempoolQuery(&chaintype.SpineChain{}),
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				MempoolService:          &mockSpineMempoolServiceSelectFail{},
				ActionTypeSwitcher:      &mockSpineTypeActionSuccess{},
				QueryExecutor:           &mockSpineAddGenesisExecutor{},
				BlockQuery:              query.NewBlockQuery(&chaintype.SpineChain{}),
				TransactionQuery:        query.NewTransactionQuery(&chaintype.SpineChain{}),
				SpinePublicKeyQuery:     query.NewSpinePublicKeyQuery(),
				Observer:                observer.NewObserver(),
				NodeRegistrationService: &mockSpineNodeRegistrationServiceSuccess{},
				BlocksmithStrategy:      &mockSpineBlocksmithServiceAddGenesisSuccess{},
				Logger:                  log.New(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:           tt.fields.Chaintype,
				QueryExecutor:       tt.fields.QueryExecutor,
				BlockQuery:          tt.fields.BlockQuery,
				SpinePublicKeyQuery: tt.fields.SpinePublicKeyQuery,
				Signature:           tt.fields.Signature,
				Observer:            tt.fields.Observer,
				BlocksmithStrategy:  tt.fields.BlocksmithStrategy,
				Logger:              tt.fields.Logger,
			}
			if err := bs.AddGenesis(); (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineService.AddGenesis() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockSpineQueryExecutorCheckGenesisTrue struct {
		query.Executor
	}
	mockSpineQueryExecutorCheckGenesisFalse struct {
		query.Executor
	}
)

func (*mockSpineQueryExecutorCheckGenesisFalse) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	mockSpine.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
		"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
		"PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
		"Version",
	}))
	return db.Query("")
}

func (*mockSpineQueryExecutorCheckGenesisFalse) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, nil
}

func (*mockSpineQueryExecutorCheckGenesisTrue) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	mockSpine.ExpectQuery("").WillReturnRows(sqlmock.NewRows(
		query.NewBlockQuery(&chaintype.SpineChain{}).Fields,
	).AddRow(
		mockSpineBlockData.GetID(),
		mockSpineBlockData.GetBlockHash(),
		mockSpineBlockData.GetPreviousBlockHash(),
		mockSpineBlockData.GetHeight(),
		mockSpineBlockData.GetTimestamp(),
		mockSpineBlockData.GetBlockSeed(),
		mockSpineBlockData.GetBlockSignature(),
		mockSpineBlockData.GetCumulativeDifficulty(),
		mockSpineBlockData.GetPayloadLength(),
		mockSpineBlockData.GetPayloadHash(),
		mockSpineBlockData.GetBlocksmithPublicKey(),
		mockSpineBlockData.GetTotalAmount(),
		mockSpineBlockData.GetTotalFee(),
		mockSpineBlockData.GetTotalCoinBase(),
		mockSpineBlockData.GetVersion(),
	))
	return db.Query("")
}

func (*mockSpineQueryExecutorCheckGenesisTrue) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mockSpine, _ := sqlmock.New()
	mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.SpineChain{}).Fields,
		).AddRow(
			constant.SpinechainGenesisBlockID,
			mockSpineBlockData.GetBlockHash(),
			mockSpineBlockData.GetPreviousBlockHash(),
			mockSpineBlockData.GetHeight(),
			mockSpineBlockData.GetTimestamp(),
			mockSpineBlockData.GetBlockSeed(),
			mockSpineBlockData.GetBlockSignature(),
			mockSpineBlockData.GetCumulativeDifficulty(),
			mockSpineBlockData.GetPayloadLength(),
			mockSpineBlockData.GetPayloadHash(),
			mockSpineBlockData.GetBlocksmithPublicKey(),
			mockSpineBlockData.GetTotalAmount(),
			mockSpineBlockData.GetTotalFee(),
			mockSpineBlockData.GetTotalCoinBase(),
			mockSpineBlockData.GetVersion(),
		))
	return db.QueryRow(qStr), nil
}

func TestBlockSpineService_CheckGenesis(t *testing.T) {
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
				Chaintype:     &chaintype.SpineChain{},
				QueryExecutor: &mockSpineQueryExecutorCheckGenesisTrue{},
				BlockQuery:    query.NewBlockQuery(&chaintype.SpineChain{}),
				Logger:        log.New(),
			},
			want: true,
		},
		{
			name: "wantFalse",
			fields: fields{
				Chaintype:     &chaintype.SpineChain{},
				QueryExecutor: &mockSpineQueryExecutorCheckGenesisFalse{},
				BlockQuery:    query.NewBlockQuery(&chaintype.SpineChain{}),
				Logger:        log.New(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:     tt.fields.Chaintype,
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
				Signature:     tt.fields.Signature,
				Logger:        tt.fields.Logger,
			}
			if got := bs.CheckGenesis(); got != tt.want {
				t.Errorf("BlockSpineService.CheckGenesis() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockSpineQueryExecutorGetBlockByHeightSuccess struct {
		query.Executor
	}
	mockSpineQueryExecutorGetBlockByHeightFail struct {
		query.Executor
	}
)

func (*mockSpineQueryExecutorGetBlockByHeightSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()

	switch qStr {
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, " +
		"total_coinbase, version FROM spine_block WHERE height = 0":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.SpineChain{}).Fields).AddRow(
			mockSpineBlockData.GetID(),
			mockSpineBlockData.GetBlockHash(),
			mockSpineBlockData.GetPreviousBlockHash(),
			mockSpineBlockData.GetHeight(),
			mockSpineBlockData.GetTimestamp(),
			mockSpineBlockData.GetBlockSeed(),
			mockSpineBlockData.GetBlockSignature(),
			mockSpineBlockData.GetCumulativeDifficulty(),
			mockSpineBlockData.GetPayloadLength(),
			mockSpineBlockData.GetPayloadHash(),
			mockSpineBlockData.GetBlocksmithPublicKey(),
			mockSpineBlockData.GetTotalAmount(),
			mockSpineBlockData.GetTotalFee(),
			mockSpineBlockData.GetTotalCoinBase(),
			mockSpineBlockData.GetVersion(),
		))
	case "SELECT node_public_key, public_key_action, latest, height FROM spine_public_key " +
		"WHERE height >= 0 AND height <= 0 AND public_key_action=0 AND latest=1 ORDER BY height":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
			query.NewSpinePublicKeyQuery().Fields))
	case "SELECT node_public_key, public_key_action, main_block_height, latest, height FROM spine_public_key WHERE height = 1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
			query.NewSpinePublicKeyQuery().Fields))
	case "SELECT id, block_id, block_height, sender_account_address, recipient_account_address, transaction_type, " +
		"fee, timestamp, transaction_hash, transaction_body_length, transaction_body_bytes, " +
		"signature, version, transaction_index FROM \"transaction\" WHERE block_id = ? ORDER BY transaction_index ASC":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
			query.NewTransactionQuery(&chaintype.SpineChain{}).Fields))
	}
	return db.Query(qStr)
}

func (*mockSpineQueryExecutorGetBlockByHeightFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}

func TestBlockSpineService_GetBlockByHeight(t *testing.T) {
	type fields struct {
		Chaintype           chaintype.ChainType
		QueryExecutor       query.ExecutorInterface
		BlockQuery          query.BlockQueryInterface
		MempoolQuery        query.MempoolQueryInterface
		TransactionQuery    query.TransactionQueryInterface
		SpinePublicKeyQuery query.SpinePublicKeyQueryInterface
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
				Chaintype:           &chaintype.SpineChain{},
				QueryExecutor:       &mockSpineQueryExecutorGetBlockByHeightSuccess{},
				BlockQuery:          query.NewBlockQuery(&chaintype.SpineChain{}),
				TransactionQuery:    query.NewTransactionQuery(&chaintype.SpineChain{}),
				SpinePublicKeyQuery: query.NewSpinePublicKeyQuery(),
			},
			want:    &mockSpineBlockData,
			wantErr: false,
		},
		{
			name: "GetBlockByHeight:FailNoEntryFound", // All is good
			fields: fields{
				Chaintype:        &chaintype.SpineChain{},
				QueryExecutor:    &mockSpineQueryExecutorGetBlockByHeightFail{},
				BlockQuery:       query.NewBlockQuery(&chaintype.SpineChain{}),
				TransactionQuery: query.NewTransactionQuery(&chaintype.SpineChain{}),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:           tt.fields.Chaintype,
				QueryExecutor:       tt.fields.QueryExecutor,
				BlockQuery:          tt.fields.BlockQuery,
				SpinePublicKeyQuery: tt.fields.SpinePublicKeyQuery,
				Signature:           tt.fields.Signature,
				Observer:            tt.fields.Observer,
			}
			got, err := bs.GetBlockByHeight(tt.args.height)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineService.GetBlockByHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockSpineService.GetBlockByHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockSpineQueryExecutorGetBlockByIDSuccess struct {
		query.Executor
	}
	mockSpineQueryExecutorGetBlockByIDFail struct {
		query.Executor
	}
)

func (*mockSpineQueryExecutorGetBlockByIDSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()

	switch qStr {
	case "SELECT node_public_key, public_key_action, latest, height FROM spine_public_key " +
		"WHERE height >= 0 AND height <= 1 AND public_key_action=0 AND latest=1 ORDER BY height":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(query.NewSpinePublicKeyQuery().Fields))
	case "SELECT node_public_key, public_key_action, main_block_height, latest, height FROM spine_public_key WHERE height = 1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(query.NewSpinePublicKeyQuery().Fields))
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, " +
		"version FROM spine_block WHERE id = 1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.SpineChain{}).Fields).AddRow(
			mockSpineBlockData.GetID(),
			mockSpineBlockData.GetBlockHash(),
			mockSpineBlockData.GetPreviousBlockHash(),
			mockSpineBlockData.GetHeight(),
			mockSpineBlockData.GetTimestamp(),
			mockSpineBlockData.GetBlockSeed(),
			mockSpineBlockData.GetBlockSignature(),
			mockSpineBlockData.GetCumulativeDifficulty(),
			mockSpineBlockData.GetPayloadLength(),
			mockSpineBlockData.GetPayloadHash(),
			mockSpineBlockData.GetBlocksmithPublicKey(),
			mockSpineBlockData.GetTotalAmount(),
			mockSpineBlockData.GetTotalFee(),
			mockSpineBlockData.GetTotalCoinBase(),
			mockSpineBlockData.GetVersion(),
		))
	case "SELECT id, block_id, block_height, sender_account_address, recipient_account_address, transaction_type, " +
		"fee, timestamp, transaction_hash, transaction_body_length, transaction_body_bytes, " +
		"signature, version, transaction_index FROM \"transaction\" WHERE block_id = ? ORDER BY transaction_index ASC":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
			query.NewTransactionQuery(&chaintype.SpineChain{}).Fields))
	}
	return db.Query(qStr)
}

func (*mockSpineQueryExecutorGetBlockByIDFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}

func (*mockSpineQueryExecutorGetBlockByIDSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()
	mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(query.NewBlockQuery(&chaintype.SpineChain{}).Fields).AddRow(
			mockSpineBlockData.GetID(),
			mockSpineBlockData.GetBlockHash(),
			mockSpineBlockData.GetPreviousBlockHash(),
			mockSpineBlockData.GetHeight(),
			mockSpineBlockData.GetTimestamp(),
			mockSpineBlockData.GetBlockSeed(),
			mockSpineBlockData.GetBlockSignature(),
			mockSpineBlockData.GetCumulativeDifficulty(),
			mockSpineBlockData.GetPayloadLength(),
			mockSpineBlockData.GetPayloadHash(),
			mockSpineBlockData.GetBlocksmithPublicKey(),
			mockSpineBlockData.GetTotalAmount(),
			mockSpineBlockData.GetTotalFee(),
			mockSpineBlockData.GetTotalCoinBase(),
			mockSpineBlockData.GetVersion(),
		))
	return db.QueryRow(qStr), nil
}

func (*mockSpineQueryExecutorGetBlockByIDFail) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, errors.New("MockedError")
}

func TestBlockSpineService_GetBlockByID(t *testing.T) {
	var mockData = mockSpineBlockData
	type fields struct {
		Chaintype           chaintype.ChainType
		QueryExecutor       query.ExecutorInterface
		BlockQuery          query.BlockQueryInterface
		MempoolQuery        query.MempoolQueryInterface
		TransactionQuery    query.TransactionQueryInterface
		SpinePublicKeyQuery query.SpinePublicKeyQueryInterface
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
				Chaintype:           &chaintype.SpineChain{},
				QueryExecutor:       &mockSpineQueryExecutorGetBlockByIDSuccess{},
				BlockQuery:          query.NewBlockQuery(&chaintype.SpineChain{}),
				TransactionQuery:    query.NewTransactionQuery(&chaintype.SpineChain{}),
				SpinePublicKeyQuery: query.NewSpinePublicKeyQuery(),
			},
			args: args{
				ID:               int64(1),
				withAttachedData: true,
			},
			want:    &mockData,
			wantErr: false,
		},
		{
			name: "GetBlockByID:FailNoEntryFound", // All is good
			fields: fields{
				Chaintype:           &chaintype.SpineChain{},
				QueryExecutor:       &mockSpineQueryExecutorGetBlockByIDFail{},
				BlockQuery:          query.NewBlockQuery(&chaintype.SpineChain{}),
				SpinePublicKeyQuery: query.NewSpinePublicKeyQuery(),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:           tt.fields.Chaintype,
				QueryExecutor:       tt.fields.QueryExecutor,
				BlockQuery:          tt.fields.BlockQuery,
				SpinePublicKeyQuery: tt.fields.SpinePublicKeyQuery,
				Signature:           tt.fields.Signature,
				Observer:            tt.fields.Observer,
			}
			got, err := bs.GetBlockByID(tt.args.ID, tt.args.withAttachedData)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineService.GetBlockByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockSpineService.GetBlockByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockSpineQueryExecutorGetBlocksFromHeightSuccess struct {
		query.Executor
	}

	mockSpineQueryExecutorGetBlocksFromHeightFail struct {
		query.Executor
	}
)

func (*mockSpineQueryExecutorGetBlocksFromHeightSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	mockSpine.ExpectQuery(qStr).WillReturnRows(sqlmock.NewRows(
		query.NewBlockQuery(&chaintype.SpineChain{}).Fields,
	).AddRow(
		mockSpineBlockData.GetID(),
		mockSpineBlockData.GetBlockHash(),
		mockSpineBlockData.GetPreviousBlockHash(),
		mockSpineBlockData.GetHeight(),
		mockSpineBlockData.GetTimestamp(),
		mockSpineBlockData.GetBlockSeed(),
		mockSpineBlockData.GetBlockSignature(),
		mockSpineBlockData.GetCumulativeDifficulty(),
		mockSpineBlockData.GetPayloadLength(),
		mockSpineBlockData.GetPayloadHash(),
		mockSpineBlockData.GetBlocksmithPublicKey(),
		mockSpineBlockData.GetTotalAmount(),
		mockSpineBlockData.GetTotalFee(),
		mockSpineBlockData.GetTotalCoinBase(),
		mockSpineBlockData.GetVersion(),
	).AddRow(
		mockSpineBlockData.GetID(),
		mockSpineBlockData.GetBlockHash(),
		mockSpineBlockData.GetPreviousBlockHash(),
		mockSpineBlockData.GetHeight(),
		mockSpineBlockData.GetTimestamp(),
		mockSpineBlockData.GetBlockSeed(),
		mockSpineBlockData.GetBlockSignature(),
		mockSpineBlockData.GetCumulativeDifficulty(),
		mockSpineBlockData.GetPayloadLength(),
		mockSpineBlockData.GetPayloadHash(),
		mockSpineBlockData.GetBlocksmithPublicKey(),
		mockSpineBlockData.GetTotalAmount(),
		mockSpineBlockData.GetTotalFee(),
		mockSpineBlockData.GetTotalCoinBase(),
		mockSpineBlockData.GetVersion(),
	),
	)
	return db.Query(qStr)
}

func (*mockSpineQueryExecutorGetBlocksFromHeightFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}

func TestBlockSpineService_GetBlocksFromHeight(t *testing.T) {
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
				Chaintype:     &chaintype.SpineChain{},
				QueryExecutor: &mockSpineQueryExecutorGetBlocksFromHeightSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.SpineChain{}),
			},
			args: args{
				startHeight: 0,
				limit:       2,
			},
			want: []*model.Block{
				&mockSpineBlockData,
				&mockSpineBlockData,
			},
			wantErr: false,
		},
		{
			name: "GetBlocksFromHeight:FailNoEntryFound", // All is good
			fields: fields{
				Chaintype:     &chaintype.SpineChain{},
				QueryExecutor: &mockSpineQueryExecutorGetBlocksFromHeightFail{},
				BlockQuery:    query.NewBlockQuery(&chaintype.SpineChain{}),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:     tt.fields.Chaintype,
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
				Signature:     tt.fields.Signature,
				Observer:      tt.fields.Observer,
			}
			got, err := bs.GetBlocksFromHeight(tt.args.startHeight, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineService.GetBlocksFromHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockSpineService.GetBlocksFromHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockSpineService_ReceiveBlock(t *testing.T) {

	mockSpineLastBlockData := model.Block{
		ID: -1701929749060110283,
		BlockHash: []byte{131, 164, 247, 141, 242, 130, 3, 197, 8, 43, 22, 189, 169, 240, 6, 44, 150, 12, 173, 148, 255, 230, 50,
			16, 166, 136, 75, 12, 106, 33, 93, 78},
		PreviousBlockHash: []byte{204, 131, 181, 204, 170, 112, 249, 115, 172, 193, 120, 7, 166, 200, 160, 138, 32, 0, 163, 161,
			45, 128, 173, 123, 252, 203, 199, 224, 249, 124, 168, 41},
		Height:    1,
		Timestamp: 1,
		BlockSeed: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		BlockSignature:       []byte{144, 246, 37, 144, 213, 135},
		CumulativeDifficulty: "1000",
		BlocksmithPublicKey: []byte{1, 2, 3, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		Version:       0,
		PayloadLength: 1,
		PayloadHash:   []byte{},
		SpinePublicKeys: []*model.SpinePublicKey{
			mockSpinePublicKey,
		},
	}

	mockSpineGoodLastBlockHash, _ := util.GetBlockHash(&mockSpineLastBlockData, &chaintype.SpineChain{})
	mockSpineGoodIncomingBlock := &model.Block{
		PreviousBlockHash:    mockSpineGoodLastBlockHash,
		BlockSignature:       nil,
		CumulativeDifficulty: "200",
		Timestamp:            10000,
		BlocksmithPublicKey:  mockSpineBlocksmiths[0].NodePublicKey,
		SpinePublicKeys: []*model.SpinePublicKey{
			mockSpinePublicKey,
		},
	}
	// TODO: remove this if unused
	// successBlockHash := []byte{197, 250, 152, 172, 169, 236, 102, 225, 55, 58, 90, 101, 214, 217,
	// 	209, 67, 185, 183, 116, 101, 64, 47, 196, 207, 27, 173, 3, 141, 12, 163, 245, 254}
	// mockSpineBlockSuccess := &model.Block{
	// 	BlockSignature:    []byte{},
	// 	BlockHash:         successBlockHash,
	// 	PreviousBlockHash: make([]byte, 32),
	// 	SpinePublicKeys:   []*model.SpinePublicKey{},
	// }

	mockSpineBlockData.BlockHash = mockSpineGoodLastBlockHash

	type fields struct {
		Chaintype               chaintype.ChainType
		KVExecutor              kvdb.KVExecutorInterface
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		SpinePublicKeyQuery     query.SpinePublicKeyQueryInterface
		MerkleTreeQuery         query.MerkleTreeQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		SkippedBlocksmithQuery  query.SkippedBlocksmithQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		AccountBalanceQuery     query.AccountBalanceQueryInterface
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
		// {
		// 	name: "ReceiveBlock:fail - {incoming block.previousBlockHash == nil}",
		// 	args: args{
		// 		senderPublicKey: nil,
		// 		lastBlock:       nil,
		// 		block: &model.Block{
		// 			PreviousBlockHash: nil,
		// 		},
		// 		nodeSecretPhrase: "",
		// 	},
		// 	fields: fields{
		// 		Chaintype:               &chaintype.SpineChain{},
		// 		QueryExecutor:           nil,
		// 		BlockQuery:              nil,
		// 		MempoolQuery:            query.NewMempoolQuery(&chaintype.SpineChain{}),
		// 		SpinePublicKeyQuery:     query.NewSpinePublicKeyQuery(),
		// 		TransactionQuery:        nil,
		// 		Signature:               nil,
		// 		MempoolService:          nil,
		// 		ActionTypeSwitcher:      nil,
		// 		AccountBalanceQuery:     nil,
		// 		Observer:                nil,
		// 		NodeRegistrationService: nil,
		// 		BlocksmithStrategy:      &mockSpineBlocksmithService{},
		// 	},
		// 	wantErr: true,
		// 	want:    nil,
		// },
		// {
		// 	name: "ReceiveBlock:fail - {last block hash != previousBlockHash}",
		// 	args: args{
		// 		senderPublicKey: nil,
		// 		lastBlock: &model.Block{
		// 			BlockHash:      []byte{1},
		// 			BlockSignature: []byte{},
		// 		},
		// 		block: &model.Block{
		// 			PreviousBlockHash: []byte{},
		// 			BlockSignature:    nil,
		// 		},
		// 		nodeSecretPhrase: "",
		// 	},
		// 	fields: fields{
		// 		Chaintype:               &chaintype.SpineChain{},
		// 		KVExecutor:              &mockSpineKVExecutorSuccess{},
		// 		QueryExecutor:           nil,
		// 		BlockQuery:              nil,
		// 		MempoolQuery:            query.NewMempoolQuery(&chaintype.SpineChain{}),
		// 		TransactionQuery:        nil,
		// 		Signature:               &mockSpineSignature{},
		// 		MempoolService:          nil,
		// 		ActionTypeSwitcher:      nil,
		// 		AccountBalanceQuery:     nil,
		// 		Observer:                nil,
		// 		BlocksmithStrategy:      &mockSpineBlocksmithService{},
		// 		NodeRegistrationService: nil,
		// 	},
		// 	wantErr: true,
		// 	want:    nil,
		// },
		// {
		// 	name: "ReceiveBlock:fail - {last block hash != previousBlockHash - kvExecutor KeyNotFound - generate batch receipt success}",
		// 	args: args{
		// 		senderPublicKey:  []byte{1, 3, 4, 5, 6},
		// 		lastBlock:        mockSpineBlockSuccess,
		// 		block:            mockSpineBlockSuccess,
		// 		nodeSecretPhrase: "",
		// 	},
		// 	fields: fields{
		// 		Chaintype:               &chaintype.SpineChain{},
		// 		KVExecutor:              &mockSpineKVExecutorSuccessKeyNotFound{},
		// 		QueryExecutor:           &mockSpineQueryExecutorSuccess{},
		// 		BlockQuery:              nil,
		// 		MempoolQuery:            query.NewMempoolQuery(&chaintype.SpineChain{}),
		// 		SpinePublicKeyQuery:     query.NewSpinePublicKeyQuery(),
		// 		MerkleTreeQuery:         query.NewMerkleTreeQuery(),
		// 		TransactionQuery:        nil,
		// 		Signature:               &mockSpineSignature{},
		// 		MempoolService:          nil,
		// 		ActionTypeSwitcher:      nil,
		// 		AccountBalanceQuery:     nil,
		// 		Observer:                nil,
		// 		NodeRegistrationService: nil,
		// 		BlocksmithStrategy:      &mockSpineBlocksmithService{},
		// 	},
		// 	wantErr: false,
		// 	want: &model.BatchReceipt{
		// 		SenderPublicKey: []byte{1, 3, 4, 5, 6},
		// 		RecipientPublicKey: []byte{
		// 			88, 220, 21, 76, 132, 107, 209, 213, 213, 206, 112, 50, 201, 183, 134, 250, 90, 163, 91, 63, 176,
		// 			223, 177, 77, 197, 161, 178, 55, 31, 225, 233, 115,
		// 		},
		// 		DatumHash:            successBlockHash,
		// 		DatumType:            constant.ReceiptDatumTypeBlock,
		// 		ReferenceBlockHeight: 0,
		// 		ReferenceBlockHash:   successBlockHash,
		// 		RMRLinked:            nil,
		// 		RecipientSignature:   []byte{},
		// 	},
		// },
		// {
		// 	name: "ReceiveBlock:fail - {last block hash != previousBlockHash - kvExecutor other error - generate batch receipt success}",
		// 	args: args{
		// 		senderPublicKey: []byte{1, 3, 4, 5, 6},
		// 		lastBlock: &model.Block{
		// 			BlockSignature: []byte{},
		// 		},
		// 		block: &model.Block{
		// 			PreviousBlockHash: []byte{133, 198, 93, 19, 200, 113, 155, 159, 136, 63, 230, 29, 21, 173, 160, 40,
		// 				169, 25, 61, 85, 203, 79, 43, 182, 5, 236, 141, 124, 46, 193, 223, 255, 0},
		// 			BlockSignature:      nil,
		// 			BlocksmithPublicKey: []byte{1, 3, 4, 5, 6},
		// 		},
		// 		nodeSecretPhrase: "",
		// 	},
		// 	fields: fields{
		// 		Chaintype:               &chaintype.SpineChain{},
		// 		KVExecutor:              &mockSpineKVExecutorFailOtherError{},
		// 		QueryExecutor:           &mockSpineQueryExecutorSuccess{},
		// 		BlockQuery:              nil,
		// 		MempoolQuery:            query.NewMempoolQuery(&chaintype.SpineChain{}),
		// 		SpinePublicKeyQuery:     query.NewSpinePublicKeyQuery(),
		// 		TransactionQuery:        nil,
		// 		Signature:               &mockSpineSignature{},
		// 		MempoolService:          nil,
		// 		ActionTypeSwitcher:      nil,
		// 		AccountBalanceQuery:     nil,
		// 		Observer:                nil,
		// 		NodeRegistrationService: nil,
		// 		BlocksmithStrategy:      &mockSpineBlocksmithService{},
		// 	},
		// 	wantErr: true,
		// 	want:    nil,
		// },
		// {
		// 	name: "ReceiveBlock:pushBlockFail",
		// 	args: args{
		// 		senderPublicKey:  []byte{1, 3, 4, 5, 6},
		// 		lastBlock:        &mockSpineBlockData,
		// 		block:            mockSpineGoodIncomingBlock,
		// 		nodeSecretPhrase: "",
		// 	},
		// 	fields: fields{
		// 		Chaintype:               &chaintype.SpineChain{},
		// 		QueryExecutor:           &mockSpineQueryExecutorFail{},
		// 		BlockQuery:              query.NewBlockQuery(&chaintype.SpineChain{}),
		// 		MempoolQuery:            query.NewMempoolQuery(&chaintype.SpineChain{}),
		// 		SpinePublicKeyQuery:     query.NewSpinePublicKeyQuery(),
		// 		TransactionQuery:        nil,
		// 		Signature:               &mockSpineSignature{},
		// 		MempoolService:          nil,
		// 		ActionTypeSwitcher:      nil,
		// 		AccountBalanceQuery:     nil,
		// 		Observer:                observer.NewObserver(),
		// 		NodeRegistrationService: nil,
		// 		BlocksmithStrategy:      &mockSpineBlocksmithService{},
		// 	},
		// 	wantErr: true,
		// 	want:    nil,
		// },
		// {
		// 	name: "ReceiveBlock:fail - {last block hash != previousBlockHash - kvExecutor other error - generate batch receipt success}",
		// 	args: args{
		// 		senderPublicKey: []byte{1, 3, 4, 5, 6},
		// 		lastBlock: &model.Block{
		// 			BlockSignature: []byte{},
		// 		},
		// 		block: &model.Block{
		// 			PreviousBlockHash: []byte{133, 198, 93, 19, 200, 113, 155, 159, 136, 63, 230, 29, 21, 173, 160, 40,
		// 				169, 25, 61, 85, 203, 79, 43, 182, 5, 236, 141, 124, 46, 193, 223, 255, 0},
		// 			BlockSignature:      nil,
		// 			BlocksmithPublicKey: []byte{1, 3, 4, 5, 6},
		// 		},
		// 		nodeSecretPhrase: "",
		// 	},
		// 	fields: fields{
		// 		Chaintype:               &chaintype.SpineChain{},
		// 		KVExecutor:              &mockSpineKVExecutorFailOtherError{},
		// 		QueryExecutor:           &mockSpineQueryExecutorSuccess{},
		// 		BlockQuery:              nil,
		// 		MempoolQuery:            query.NewMempoolQuery(&chaintype.SpineChain{}),
		// 		SpinePublicKeyQuery:     query.NewSpinePublicKeyQuery(),
		// 		TransactionQuery:        nil,
		// 		Signature:               &mockSpineSignature{},
		// 		MempoolService:          nil,
		// 		ActionTypeSwitcher:      nil,
		// 		AccountBalanceQuery:     nil,
		// 		Observer:                nil,
		// 		NodeRegistrationService: nil,
		// 		BlocksmithStrategy:      &mockSpineBlocksmithService{},
		// 	},
		// 	wantErr: true,
		// 	want:    nil,
		// },
		// {
		// 	name: "ReceiveBlock:pushBlockFail",
		// 	args: args{
		// 		senderPublicKey:  []byte{1, 3, 4, 5, 6},
		// 		lastBlock:        &mockSpineBlockData,
		// 		block:            mockSpineGoodIncomingBlock,
		// 		nodeSecretPhrase: "",
		// 	},
		// 	fields: fields{
		// 		Chaintype:               &chaintype.SpineChain{},
		// 		QueryExecutor:           &mockSpineQueryExecutorFail{},
		// 		BlockQuery:              query.NewBlockQuery(&chaintype.SpineChain{}),
		// 		MempoolQuery:            query.NewMempoolQuery(&chaintype.SpineChain{}),
		// 		SpinePublicKeyQuery:     query.NewSpinePublicKeyQuery(),
		// 		TransactionQuery:        nil,
		// 		Signature:               &mockSpineSignature{},
		// 		MempoolService:          nil,
		// 		ActionTypeSwitcher:      nil,
		// 		AccountBalanceQuery:     nil,
		// 		Observer:                observer.NewObserver(),
		// 		NodeRegistrationService: nil,
		// 		BlocksmithStrategy:      &mockSpineBlocksmithService{},
		// 	},
		// 	wantErr: true,
		// 	want:    nil,
		// },
		{
			name: "ReceiveBlock:success",
			args: args{
				senderPublicKey:  []byte{1, 3, 4, 5, 6},
				lastBlock:        &mockSpineLastBlockData,
				block:            mockSpineGoodIncomingBlock,
				nodeSecretPhrase: "",
			},
			fields: fields{
				Chaintype:               &chaintype.SpineChain{},
				KVExecutor:              &mockSpineKVExecutorSuccess{},
				QueryExecutor:           &mockSpineQueryExecutorSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.SpineChain{}),
				MempoolQuery:            query.NewMempoolQuery(&chaintype.SpineChain{}),
				SpinePublicKeyQuery:     query.NewSpinePublicKeyQuery(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				TransactionQuery:        query.NewTransactionQuery(&chaintype.SpineChain{}),
				MerkleTreeQuery:         query.NewMerkleTreeQuery(),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				SkippedBlocksmithQuery:  query.NewSkippedBlocksmithQuery(),
				Signature:               &mockSpineSignature{},
				MempoolService:          nil,
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				Observer:                observer.NewObserver(),
				BlocksmithStrategy:      &mockSpineBlocksmithServicePushBlock{},
				NodeRegistrationService: &mockSpineNodeRegistrationServiceSuccess{},
			},
			wantErr: false,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:           tt.fields.Chaintype,
				KVExecutor:          tt.fields.KVExecutor,
				QueryExecutor:       tt.fields.QueryExecutor,
				BlockQuery:          tt.fields.BlockQuery,
				SpinePublicKeyQuery: tt.fields.SpinePublicKeyQuery,
				Signature:           tt.fields.Signature,
				Observer:            tt.fields.Observer,
				BlocksmithStrategy:  tt.fields.BlocksmithStrategy,
				Logger:              logrus.New(),
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

func TestBlockSpineService_GetBlockExtendedInfo(t *testing.T) {
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
			name: "GetBlockExtendedInfo:success-{genesisBlock}",
			args: args{
				block: genesisBlock,
			},
			fields: fields{
				QueryExecutor:          &mockSpineQueryExecutorSuccess{},
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
					Version:              0,
				},
			},
		},
		{
			name: "GetBlockExtendedInfo:success",
			args: args{
				block: block,
			},
			fields: fields{
				QueryExecutor:          &mockSpineQueryExecutorSuccess{},
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
					Version:              0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:     tt.fields.Chaintype,
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
				Signature:     tt.fields.Signature,
				Observer:      tt.fields.Observer,
			}
			got, err := bs.GetBlockExtendedInfo(tt.args.block, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineService.GetBlockExtendedInfo() error = \n%v, wantErr \n%v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockSpineService.GetBlockExtendedInfo() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

type (
	mockSpineBlocksmithService struct {
		strategy.BlocksmithStrategyMain
	}
)

func (*mockSpineBlocksmithService) GetSortedBlocksmiths(block *model.Block) []*model.Blocksmith {
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
func TestBlockSpineService_GenerateGenesisBlock(t *testing.T) {
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
				Chaintype:               &chaintype.SpineChain{},
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
				genesisEntries: constant.GenesisConfig,
			},
			wantErr: false,
			want:    constant.SpinechainGenesisBlockID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:     tt.fields.Chaintype,
				KVExecutor:    tt.fields.KVExecutor,
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
				Signature:     tt.fields.Signature,
				Observer:      tt.fields.Observer,
				Logger:        tt.fields.Logger,
			}
			got, err := bs.GenerateGenesisBlock(tt.args.genesisEntries)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineService.GenerateGenesisBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.ID != tt.want {
				t.Errorf("BlockSpineService.GenerateGenesisBlock() got %v, want %v", got.GetID(), tt.want)
			}
		})
	}
}

type mockSpineQueryExecutorValidateBlockSuccess struct {
	query.Executor
}

func (*mockSpineQueryExecutorValidateBlockSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()
	mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.SpineChain{}).Fields,
		).AddRow(
			mockSpineBlockData.GetID(),
			mockSpineBlockData.GetBlockHash(),
			mockSpineBlockData.GetPreviousBlockHash(),
			mockSpineBlockData.GetHeight(),
			mockSpineBlockData.GetTimestamp(),
			mockSpineBlockData.GetBlockSeed(),
			mockSpineBlockData.GetBlockSignature(),
			mockSpineBlockData.GetCumulativeDifficulty(),
			mockSpineBlockData.GetPayloadLength(),
			mockSpineBlockData.GetPayloadHash(),
			mockSpineBlockData.GetBlocksmithPublicKey(),
			mockSpineBlockData.GetTotalAmount(),
			mockSpineBlockData.GetTotalFee(),
			mockSpineBlockData.GetTotalCoinBase(),
			mockSpineBlockData.GetVersion(),
		))
	rows, _ := db.Query(qStr)
	return rows, nil
}

var (
	mockSpineValidateBadBlockInvalidBlockHash = &model.Block{
		Timestamp:           1572246820,
		BlockSignature:      []byte{},
		BlocksmithPublicKey: []byte{1, 2, 3, 4},
		PreviousBlockHash:   []byte{},
	}

	mockSpineValidateBlockSuccess = &model.Block{
		Timestamp: 1572246820,
		ID:        constant.MainchainGenesisBlockID,
		BlockHash: make([]byte, 32),
		PreviousBlockHash: []byte{167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160, 97, 214, 98, 245, 128, 255, 77, 228,
			59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74},
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
	mockSpineBlocksmithServiceValidateBlockSuccess struct {
		strategy.BlocksmithStrategyMain
	}
)

func (*mockSpineBlocksmithServiceValidateBlockSuccess) GetSortedBlocksmithsMap(*model.Block) map[string]*int64 {
	firstIndex := int64(0)
	secondIndex := int64(1)
	return map[string]*int64{
		string(mockSpineValidateBadBlockInvalidBlockHash.BlocksmithPublicKey): &firstIndex,
		string(mockSpineBlockData.BlocksmithPublicKey):                        &secondIndex,
	}
}
func (*mockSpineBlocksmithServiceValidateBlockSuccess) GetSmithTime(blocksmithIndex int64, previousBlock *model.Block) int64 {
	return 0
}

func TestBlockSpineService_ValidateBlock(t *testing.T) {
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
				Signature:          &mockSpineSignatureFail{},
				BlocksmithStrategy: &mockSpineBlocksmithServiceValidateBlockSuccess{},
			},
			wantErr: true,
		},
		{
			name: "ValidateBlock:fail-{InvalidSignature}",
			args: args{
				block:   mockSpineValidateBadBlockInvalidBlockHash,
				curTime: 1572246820,
			},
			fields: fields{
				Signature:          &mockSpineSignatureFail{},
				BlocksmithStrategy: &mockSpineBlocksmithServiceValidateBlockSuccess{},
			},
			wantErr: true,
		},
		{
			name: "ValidateBlock:fail-{InvalidBlockHash}",
			args: args{
				block:             mockSpineValidateBadBlockInvalidBlockHash,
				previousLastBlock: &model.Block{},
				curTime:           1572246820,
			},
			fields: fields{
				Signature:          &mockSpineSignature{},
				BlocksmithStrategy: &mockSpineBlocksmithServiceValidateBlockSuccess{},
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
					PreviousBlockHash: []byte{204, 131, 181, 204, 170, 112, 249, 115, 172, 193, 120, 7, 166, 200, 160,
						138, 32, 0, 163, 161, 45, 128, 173, 123, 252, 203, 199, 224, 249, 124, 168, 41},
					CumulativeDifficulty: "10",
				},
				previousLastBlock: &model.Block{},
				curTime:           1572246820,
			},
			fields: fields{
				Signature:          &mockSpineSignature{},
				BlockQuery:         query.NewBlockQuery(&chaintype.SpineChain{}),
				QueryExecutor:      &mockSpineQueryExecutorValidateBlockSuccess{},
				BlocksmithStrategy: &mockSpineBlocksmithServiceValidateBlockSuccess{},
			},
			wantErr: true,
		},
		{
			name: "ValidateBlock:success",
			args: args{
				block:             mockSpineValidateBlockSuccess,
				previousLastBlock: &model.Block{},
				curTime:           mockSpineValidateBlockSuccess.Timestamp,
			},
			fields: fields{
				Signature:          &mockSpineSignature{},
				BlockQuery:         query.NewBlockQuery(&chaintype.SpineChain{}),
				QueryExecutor:      &mockSpineQueryExecutorValidateBlockSuccess{},
				BlocksmithStrategy: &mockSpineBlocksmithServiceValidateBlockSuccess{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:          tt.fields.Chaintype,
				KVExecutor:         tt.fields.KVExecutor,
				QueryExecutor:      tt.fields.QueryExecutor,
				BlockQuery:         tt.fields.BlockQuery,
				Signature:          tt.fields.Signature,
				BlocksmithStrategy: tt.fields.BlocksmithStrategy,
				Observer:           tt.fields.Observer,
				Logger:             tt.fields.Logger,
			}
			if err := bs.ValidateBlock(tt.args.block, tt.args.previousLastBlock, tt.args.curTime); (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineService.ValidateBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockSpinePopOffToBlockReturnCommonBlock struct {
		query.Executor
	}
	mockSpinePopOffToBlockReturnBeginTxFunc struct {
		query.Executor
	}
	mockSpinePopOffToBlockReturnWantFailOnCommit struct {
		query.Executor
	}
	mockSpinePopOffToBlockReturnWantFailOnExecuteTransactions struct {
		query.Executor
	}
)

func (*mockSpinePopOffToBlockReturnCommonBlock) BeginTx() error {
	return nil
}
func (*mockSpinePopOffToBlockReturnCommonBlock) CommitTx() error {
	return nil
}
func (*mockSpinePopOffToBlockReturnCommonBlock) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}
func (*mockSpinePopOffToBlockReturnCommonBlock) ExecuteSelect(qSrt string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()

	mockSpine.ExpectQuery("").WillReturnRows(
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
func (*mockSpinePopOffToBlockReturnCommonBlock) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}
func (*mockSpinePopOffToBlockReturnBeginTxFunc) BeginTx() error {
	return errors.New("i want this")
}
func (*mockSpinePopOffToBlockReturnBeginTxFunc) CommitTx() error {
	return nil
}
func (*mockSpinePopOffToBlockReturnWantFailOnCommit) BeginTx() error {
	return nil
}
func (*mockSpinePopOffToBlockReturnWantFailOnCommit) CommitTx() error {
	return errors.New("i want this")
}
func (*mockSpinePopOffToBlockReturnWantFailOnCommit) ExecuteSelect(qSrt string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()
	mockSpine.ExpectQuery("").WillReturnRows(
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
func (*mockSpinePopOffToBlockReturnWantFailOnExecuteTransactions) BeginTx() error {
	return nil
}
func (*mockSpinePopOffToBlockReturnWantFailOnExecuteTransactions) CommitTx() error {
	return nil
}
func (*mockSpinePopOffToBlockReturnWantFailOnExecuteTransactions) ExecuteTransactions(queries [][]interface{}) error {
	return errors.New("i want this")
}
func (*mockSpinePopOffToBlockReturnWantFailOnExecuteTransactions) RollbackTx() error {
	return nil
}

var (
	mockSpineGoodBlock = &model.Block{
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
	mockSpineGoodCommonBlock = &model.Block{
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
	mockSpineBadCommonBlockHardFork = &model.Block{
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
	mockSpineExecutorBlockPopGetLastBlockFail struct {
		query.Executor
	}
	mockSpineExecutorBlockPopSuccess struct {
		query.Executor
	}
	mockSpineExecutorBlockPopFailCommonNotFound struct {
		mockSpineExecutorBlockPopSuccess
	}
	mockSpineReceiptSuccess struct {
		ReceiptService
	}
	mockSpineReceiptFail struct {
		ReceiptService
	}
	mockSpineMempoolServiceBlockPopSuccess struct {
		MempoolService
	}
	mockSpineMempoolServiceBlockPopFail struct {
		MempoolService
	}
	mockSpineNodeRegistrationServiceBlockPopSuccess struct {
		NodeRegistrationService
	}
)

func (*mockSpineExecutorBlockPopFailCommonNotFound) ExecuteSelectRow(
	qStr string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	blockQ := query.NewBlockQuery(&chaintype.SpineChain{})
	switch qStr {
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version FROM spine_block ORDER BY height DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields))
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version FROM spine_block WHERE id = 1":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields))
	default:
		return nil, fmt.Errorf("unmocked query: %s", qStr)
	}
	return db.QueryRow(qStr), nil
}

func (*mockSpineExecutorBlockPopFailCommonNotFound) ExecuteSelect(
	qStr string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	transactionQ := query.NewTransactionQuery(&chaintype.SpineChain{})
	blockQ := query.NewBlockQuery(&chaintype.SpineChain{})
	switch qStr {
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version FROM spine_block ORDER BY height DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields).AddRow(
				mockSpineGoodBlock.GetID(),
				mockSpineGoodBlock.GetBlockHash(),
				mockSpineGoodBlock.GetPreviousBlockHash(),
				mockSpineGoodBlock.GetHeight(),
				mockSpineGoodBlock.GetTimestamp(),
				mockSpineGoodBlock.GetBlockSeed(),
				mockSpineGoodBlock.GetBlockSignature(),
				mockSpineGoodBlock.GetCumulativeDifficulty(),
				mockSpineGoodBlock.GetPayloadLength(),
				mockSpineGoodBlock.GetPayloadHash(),
				mockSpineGoodBlock.GetBlocksmithPublicKey(),
				mockSpineGoodBlock.GetTotalAmount(),
				mockSpineGoodBlock.GetTotalFee(),
				mockSpineGoodBlock.GetTotalCoinBase(),
			),
		)
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version FROM spine_block WHERE id = 0":
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

func (*mockSpineExecutorBlockPopGetLastBlockFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	blockQ := query.NewBlockQuery(&chaintype.SpineChain{})
	switch qStr {
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version FROM main_block WHERE id = 0":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields))
	default:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields[:len(blockQ.Fields)-1]).AddRow(
				mockSpineGoodBlock.GetID(),
				mockSpineGoodBlock.GetBlockHash(),
				mockSpineGoodBlock.GetPreviousBlockHash(),
				mockSpineGoodBlock.GetHeight(),
				mockSpineGoodBlock.GetTimestamp(),
				mockSpineGoodBlock.GetBlockSeed(),
				mockSpineGoodBlock.GetBlockSignature(),
				mockSpineGoodBlock.GetCumulativeDifficulty(),
				mockSpineGoodBlock.GetPayloadLength(),
				mockSpineGoodBlock.GetPayloadHash(),
				mockSpineGoodBlock.GetBlocksmithPublicKey(),
				mockSpineGoodBlock.GetTotalAmount(),
				mockSpineGoodBlock.GetTotalFee(),
				mockSpineGoodBlock.GetTotalCoinBase(),
			),
		)
	}
	return db.QueryRow(qStr), nil
}

func (*mockSpineNodeRegistrationServiceBlockPopSuccess) ResetScrambledNodes() {

}

func (*mockSpineMempoolServiceBlockPopSuccess) GetMempoolTransactionsWantToBackup(
	height uint32,
) ([]*model.MempoolTransaction, error) {
	return make([]*model.MempoolTransaction, 0), nil
}

func (*mockSpineMempoolServiceBlockPopFail) GetMempoolTransactionsWantToBackup(
	height uint32,
) ([]*model.MempoolTransaction, error) {
	return nil, errors.New("mockSpineedError")
}

func (*mockSpineReceiptSuccess) GetPublishedReceiptsByHeight(blockHeight uint32) ([]*model.PublishedReceipt, error) {
	return make([]*model.PublishedReceipt, 0), nil
}

func (*mockSpineReceiptFail) GetPublishedReceiptsByHeight(blockHeight uint32) ([]*model.PublishedReceipt, error) {
	return nil, errors.New("mockSpineError")
}

func (*mockSpineExecutorBlockPopSuccess) BeginTx() error {
	return nil
}

func (*mockSpineExecutorBlockPopSuccess) CommitTx() error {
	return nil
}

func (*mockSpineExecutorBlockPopSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}
func (*mockSpineExecutorBlockPopSuccess) RollbackTx() error {
	return nil
}
func (*mockSpineExecutorBlockPopSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()

	transactionQ := query.NewTransactionQuery(&chaintype.SpineChain{})
	blockQ := query.NewBlockQuery(&chaintype.SpineChain{})
	spinePubKeyQ := query.NewSpinePublicKeyQuery()
	switch qStr {
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version FROM spine_block WHERE id = 0":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields).AddRow(
				mockSpineGoodCommonBlock.GetID(),
				mockSpineGoodCommonBlock.GetBlockHash(),
				mockSpineGoodCommonBlock.GetPreviousBlockHash(),
				mockSpineGoodCommonBlock.GetHeight(),
				mockSpineGoodCommonBlock.GetTimestamp(),
				mockSpineGoodCommonBlock.GetBlockSeed(),
				mockSpineGoodCommonBlock.GetBlockSignature(),
				mockSpineGoodCommonBlock.GetCumulativeDifficulty(),
				mockSpineGoodCommonBlock.GetPayloadLength(),
				mockSpineGoodCommonBlock.GetPayloadHash(),
				mockSpineGoodCommonBlock.GetBlocksmithPublicKey(),
				mockSpineGoodCommonBlock.GetTotalAmount(),
				mockSpineGoodCommonBlock.GetTotalFee(),
				mockSpineGoodCommonBlock.GetTotalCoinBase(),
				mockSpineGoodCommonBlock.GetVersion(),
			),
		)
	case "SELECT node_public_key, public_key_action, latest, height FROM spine_public_key " +
		"WHERE height >= 0 AND height <= 1000 AND public_key_action=0 AND latest=1 ORDER BY height":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(spinePubKeyQ.Fields))
	case "SELECT node_public_key, public_key_action, main_block_height, latest, height FROM spine_public_key WHERE height = 1000":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(spinePubKeyQ.Fields))
	case "SELECT id, block_id, block_height, sender_account_address, recipient_account_address, transaction_type, fee, " +
		"timestamp, transaction_hash, transaction_body_length, transaction_body_bytes, signature, version, " +
		"transaction_index FROM \"transaction\" WHERE block_id = ? ORDER BY transaction_index ASC":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(transactionQ.Fields))
	}

	return db.Query(qStr)
}

func (*mockSpineExecutorBlockPopSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()

	blockQ := query.NewBlockQuery(&chaintype.SpineChain{})

	mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
		sqlmock.NewRows(blockQ.Fields).AddRow(
			mockSpineGoodBlock.GetID(),
			mockSpineGoodBlock.GetBlockHash(),
			mockSpineGoodBlock.GetPreviousBlockHash(),
			mockSpineGoodBlock.GetHeight(),
			mockSpineGoodBlock.GetTimestamp(),
			mockSpineGoodBlock.GetBlockSeed(),
			mockSpineGoodBlock.GetBlockSignature(),
			mockSpineGoodBlock.GetCumulativeDifficulty(),
			mockSpineGoodBlock.GetPayloadLength(),
			mockSpineGoodBlock.GetPayloadHash(),
			mockSpineGoodBlock.GetBlocksmithPublicKey(),
			mockSpineGoodBlock.GetTotalAmount(),
			mockSpineGoodBlock.GetTotalFee(),
			mockSpineGoodBlock.GetTotalCoinBase(),
			mockSpineGoodBlock.GetVersion(),
		),
	)
	return db.QueryRow(qStr), nil
}

func TestBlockSpineService_PopOffToBlock(t *testing.T) {
	type fields struct {
		RWMutex                 sync.RWMutex
		Chaintype               chaintype.ChainType
		KVExecutor              kvdb.KVExecutorInterface
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		SpinePublicKeyQuery     query.SpinePublicKeyQueryInterface
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
				Chaintype:               &chaintype.SpineChain{},
				KVExecutor:              nil,
				QueryExecutor:           &mockSpineExecutorBlockPopGetLastBlockFail{},
				BlockQuery:              query.NewBlockQuery(&chaintype.SpineChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.SpineChain{}),
				SpinePublicKeyQuery:     query.NewSpinePublicKeyQuery(),
				MerkleTreeQuery:         nil,
				PublishedReceiptQuery:   nil,
				SkippedBlocksmithQuery:  nil,
				Signature:               nil,
				MempoolService:          &mockSpineMempoolServiceBlockPopSuccess{},
				ReceiptService:          &mockSpineReceiptSuccess{},
				NodeRegistrationService: &mockSpineNodeRegistrationServiceBlockPopSuccess{},
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				ParticipationScoreQuery: nil,
				NodeRegistrationQuery:   nil,
				Observer:                nil,
				Logger:                  logrus.New(),
			},
			args: args{
				commonBlock: mockSpineGoodCommonBlock,
			},
			want:    make([]*model.Block, 0),
			wantErr: true,
		},
		{
			name: "Fail - HardFork",
			fields: fields{
				RWMutex:                 sync.RWMutex{},
				Chaintype:               &chaintype.SpineChain{},
				KVExecutor:              nil,
				QueryExecutor:           &mockSpineExecutorBlockPopSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.SpineChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.SpineChain{}),
				SpinePublicKeyQuery:     query.NewSpinePublicKeyQuery(),
				MerkleTreeQuery:         nil,
				PublishedReceiptQuery:   nil,
				SkippedBlocksmithQuery:  nil,
				Signature:               nil,
				MempoolService:          &mockSpineMempoolServiceBlockPopSuccess{},
				ReceiptService:          &mockSpineReceiptSuccess{},
				NodeRegistrationService: &mockSpineNodeRegistrationServiceBlockPopSuccess{},
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				ParticipationScoreQuery: nil,
				NodeRegistrationQuery:   nil,
				Observer:                nil,
				Logger:                  logrus.New(),
			},
			args: args{
				commonBlock: mockSpineBadCommonBlockHardFork,
			},
			want:    make([]*model.Block, 0),
			wantErr: false,
		},
		{
			name: "Fail - CommonBlockNotFound",
			fields: fields{
				Chaintype:               &chaintype.SpineChain{},
				KVExecutor:              nil,
				QueryExecutor:           &mockSpineExecutorBlockPopFailCommonNotFound{},
				BlockQuery:              query.NewBlockQuery(&chaintype.SpineChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.SpineChain{}),
				SpinePublicKeyQuery:     query.NewSpinePublicKeyQuery(),
				MerkleTreeQuery:         nil,
				PublishedReceiptQuery:   nil,
				SkippedBlocksmithQuery:  nil,
				Signature:               nil,
				MempoolService:          &mockSpineMempoolServiceBlockPopSuccess{},
				ReceiptService:          &mockSpineReceiptSuccess{},
				NodeRegistrationService: &mockSpineNodeRegistrationServiceBlockPopSuccess{},
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				ParticipationScoreQuery: nil,
				NodeRegistrationQuery:   nil,
				Observer:                nil,
				Logger:                  logrus.New(),
			},
			args: args{
				commonBlock: mockSpineGoodCommonBlock,
			},
			want:    make([]*model.Block, 0),
			wantErr: true,
		},
		{
			name: "Success",
			fields: fields{
				RWMutex:                 sync.RWMutex{},
				Chaintype:               &chaintype.SpineChain{},
				KVExecutor:              nil,
				QueryExecutor:           &mockSpineExecutorBlockPopSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.SpineChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.SpineChain{}),
				SpinePublicKeyQuery:     query.NewSpinePublicKeyQuery(),
				MerkleTreeQuery:         nil,
				PublishedReceiptQuery:   nil,
				SkippedBlocksmithQuery:  nil,
				Signature:               nil,
				MempoolService:          &mockSpineMempoolServiceBlockPopSuccess{},
				ReceiptService:          &mockSpineReceiptSuccess{},
				NodeRegistrationService: &mockSpineNodeRegistrationServiceBlockPopSuccess{},
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				ParticipationScoreQuery: nil,
				NodeRegistrationQuery:   nil,
				Observer:                nil,
				Logger:                  logrus.New(),
			},
			args: args{
				commonBlock: mockSpineGoodCommonBlock,
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:           tt.fields.Chaintype,
				KVExecutor:          tt.fields.KVExecutor,
				QueryExecutor:       tt.fields.QueryExecutor,
				BlockQuery:          tt.fields.BlockQuery,
				SpinePublicKeyQuery: tt.fields.SpinePublicKeyQuery,
				Signature:           tt.fields.Signature,
				Observer:            tt.fields.Observer,
				Logger:              tt.fields.Logger,
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
	mockSpineExecutorPopulateBlockDataFail struct {
		query.Executor
	}
	mockSpineExecutorPopulateBlockDataSuccess struct {
		query.Executor
	}
)

func (*mockSpineExecutorPopulateBlockDataFail) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("Mock Error")
}

func (*mockSpineExecutorPopulateBlockDataSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()
	switch qStr {
	case "SELECT node_public_key, public_key_action, main_block_height, latest, height FROM spine_public_key WHERE height = 0":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).
			WillReturnRows(sqlmock.NewRows(
				query.NewSpinePublicKeyQuery().Fields,
			).AddRow(
				mockSpinePublicKey.NodePublicKey,
				mockSpinePublicKey.PublicKeyAction,
				mockSpinePublicKey.MainBlockHeight,
				mockSpinePublicKey.Latest,
				mockSpinePublicKey.Height,
			))
	default:
		return nil, fmt.Errorf("unmocked sql query in mockSpineExecutorPopulateBlockDataSuccess: %s", qStr)
	}
	rows, _ := db.Query(qStr)
	return rows, nil
}

func TestBlockSpineService_PopulateBlockData(t *testing.T) {
	type fields struct {
		Chaintype             chaintype.ChainType
		KVExecutor            kvdb.KVExecutorInterface
		QueryExecutor         query.ExecutorInterface
		BlockQuery            query.BlockQueryInterface
		SpinePublicKeyQuery   query.SpinePublicKeyQueryInterface
		Signature             crypto.SignatureInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlocksmithStrategy    strategy.BlocksmithStrategyInterface
		Observer              *observer.Observer
		Logger                *log.Logger
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
				Chaintype:           &chaintype.SpineChain{},
				QueryExecutor:       &mockSpineExecutorPopulateBlockDataFail{},
				SpinePublicKeyQuery: query.NewSpinePublicKeyQuery(),
				Logger:              logrus.New(),
			},
			args: args{
				block: &model.Block{},
			},
			wantErr: true,
		},
		{
			name: "PopulateBlockData:success",
			fields: fields{
				Chaintype:           &chaintype.SpineChain{},
				QueryExecutor:       &mockSpineExecutorPopulateBlockDataSuccess{},
				SpinePublicKeyQuery: query.NewSpinePublicKeyQuery(),
				Logger:              logrus.New(),
			},
			args: args{
				block: &model.Block{
					ID: int64(-1701929749060110283),
				},
			},
			wantErr: false,
			expects: &model.Block{
				ID: int64(-1701929749060110283),
				SpinePublicKeys: []*model.SpinePublicKey{
					mockSpinePublicKey,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:             tt.fields.Chaintype,
				KVExecutor:            tt.fields.KVExecutor,
				QueryExecutor:         tt.fields.QueryExecutor,
				BlockQuery:            tt.fields.BlockQuery,
				SpinePublicKeyQuery:   tt.fields.SpinePublicKeyQuery,
				Signature:             tt.fields.Signature,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlocksmithStrategy:    tt.fields.BlocksmithStrategy,
				Observer:              tt.fields.Observer,
				Logger:                tt.fields.Logger,
			}
			if err := bs.PopulateBlockData(tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineService.PopulateBlockData() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.expects != nil && !reflect.DeepEqual(tt.args.block, tt.expects) {
				t.Errorf("BlockSpineService.PopulateBlockData() = %v, want %v", tt.expects, tt.args.block)
			}
		})
	}
}

type (
	mockNodeRegistationQueryExecutorSuccess struct {
		query.Executor
	}
)

func (*mockNodeRegistationQueryExecutorSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()
	switch qStr {
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE height >= (SELECT MIN(height) " +
		"FROM main_block AS mb1 WHERE mb1.timestamp >= 1) AND height <= (SELECT MAX(height) " +
		"FROM main_block AS mb2 WHERE mb2.timestamp < 2) AND registration_status != 1 AND latest=1 ORDER BY height":
		mockNodeRegistrationRows := mockSpine.NewRows(query.NewNodeRegistrationQuery().Fields)
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(mockNodeRegistrationRows)
	default:
		return nil, fmt.Errorf("unmocked query for mockNodeRegistationQueryExecutorSuccess: %s", qStr)
	}
	rows, _ := db.Query(qStr)
	return rows, nil
}

func TestBlockSpineService_BuildSpinePublicKeysFromNodeRegistry(t *testing.T) {
	type fields struct {
		Chaintype             chaintype.ChainType
		KVExecutor            kvdb.KVExecutorInterface
		QueryExecutor         query.ExecutorInterface
		BlockQuery            query.BlockQueryInterface
		SpinePublicKeyQuery   query.SpinePublicKeyQueryInterface
		Signature             crypto.SignatureInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlocksmithStrategy    strategy.BlocksmithStrategyInterface
		Observer              *observer.Observer
		Logger                *log.Logger
	}
	type args struct {
		fromTimestamp int64
		toTimestamp   int64
		spineHeight   uint32
	}
	tests := []struct {
		name                string
		fields              fields
		args                args
		wantSpinePublicKeys []*model.SpinePublicKey
		wantErr             bool
	}{
		{
			name: "BuildSpinePublicKeysFromNodeRegistry:success",
			fields: fields{
				Chaintype:             &chaintype.SpineChain{},
				KVExecutor:            nil,
				QueryExecutor:         &mockNodeRegistationQueryExecutorSuccess{},
				BlockQuery:            query.NewBlockQuery(&chaintype.SpineChain{}),
				SpinePublicKeyQuery:   query.NewSpinePublicKeyQuery(),
				Signature:             nil,
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Logger:                logrus.New(),
			},
			args: args{
				fromTimestamp: 1,
				toTimestamp:   2,
				spineHeight:   1,
			},
			wantSpinePublicKeys: []*model.SpinePublicKey{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:             tt.fields.Chaintype,
				KVExecutor:            tt.fields.KVExecutor,
				QueryExecutor:         tt.fields.QueryExecutor,
				BlockQuery:            tt.fields.BlockQuery,
				SpinePublicKeyQuery:   tt.fields.SpinePublicKeyQuery,
				Signature:             tt.fields.Signature,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlocksmithStrategy:    tt.fields.BlocksmithStrategy,
				Observer:              tt.fields.Observer,
				Logger:                tt.fields.Logger,
			}
			gotSpinePublicKeys, err := bs.BuildSpinePublicKeysFromNodeRegistry(tt.args.fromTimestamp, tt.args.toTimestamp, tt.args.spineHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineService.BuildSpinePublicKeysFromNodeRegistry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotSpinePublicKeys, tt.wantSpinePublicKeys) {
				t.Errorf("BlockSpineService.BuildSpinePublicKeysFromNodeRegistry() = %v, want %v", gotSpinePublicKeys, tt.wantSpinePublicKeys)
			}
		})
	}
}
