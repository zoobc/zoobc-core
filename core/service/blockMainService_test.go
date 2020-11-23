package service

import (
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/zoobc/zoobc-core/common/crypto"

	"github.com/zoobc/zoobc-core/common/blocker"

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/smith/strategy"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
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

	mockNodeRegistrationServiceSuccess struct {
		NodeRegistrationService
	}

	mockNodeRegistrationServiceFail struct {
		NodeRegistrationService
	}
)

func (*mockNodeRegistrationServiceSuccess) GetActiveRegistryNodeWithTotalParticipationScore() ([]storage.NodeRegistry, int64, error) {
	return []storage.NodeRegistry{}, 0, nil
}

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
			AccountAddress: []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126,
				202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
		},
	}, nil
}

func (*mockNodeRegistrationServiceSuccess) AdmitNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	return nil
}

func (*mockNodeRegistrationServiceSuccess) CommitCacheTransaction() error {
	return nil
}

func (*mockNodeRegistrationServiceSuccess) SelectNodesToBeExpelled() ([]*model.NodeRegistration, error) {
	return []*model.NodeRegistration{
		{
			AccountAddress: []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126,
				202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
		},
	}, nil
}

func (*mockNodeRegistrationServiceSuccess) BeginCacheTransaction() error {
	return nil
}

func (*mockNodeRegistrationServiceSuccess) RollbackCacheTransaction() error {
	return nil
}

func (*mockNodeRegistryCacheAlwaysSuccess) CommitCacheTransaction() error {
	return nil
}

func (*mockNodeRegistrationServiceSuccess) GetNextNodeAdmissionTimestamp() (*model.NodeAdmissionTimestamp, error) {
	return &model.NodeAdmissionTimestamp{
		Timestamp: mockBlockPushBlock.Timestamp + 1,
	}, nil
}

func (*mockNodeRegistrationServiceSuccess) UpdateNextNodeAdmissionCache(
	newNextNodeAdmission *model.NodeAdmissionTimestamp,
) error {
	return nil
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
			AccountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139,
				255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
		},
	}, nil
}
func (*mockNodeRegistrationServiceFail) ExpelNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	return nil
}

func (*mockNodeRegistrationServiceFail) GetNextNodeAdmissionTimestamp() (*model.NodeAdmissionTimestamp, error) {
	return &model.NodeAdmissionTimestamp{
		Timestamp: mockBlockPushBlock.Timestamp + 1,
	}, nil
}
func (*mockNodeRegistrationServiceFail) UpdateNextNodeAdmissionCache(
	newNextNodeAdmission *model.NodeAdmissionTimestamp,
) error {
	return nil
}

func (*mockNodeRegistrationServiceFail) BackupCache() error {
	return nil
}

func (*mockNodeRegistrationServiceFail) RestoreCacheTransaction() {
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

var (
	bcsAddress1 = []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
		72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169}
	bcsAddress2 = []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126,
		202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14}
	bcsAddress3 = []byte{0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250, 58, 2, 83, 75, 213, 137, 66, 236, 188, 43,
		59, 241, 146, 243, 147, 58, 161, 35, 229, 54}
	bcsNodePubKey1 = []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
	bcsNodePubKey2 = []byte{1, 2, 3, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
	bcsNodePubKey3 = []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
	mockSendMoneyTxBody = &transaction.SendMoney{
		Body: &model.SendMoneyTransactionBody{
			Amount: 10,
		},
	}
	mockSendMoneyTxBodyBytes, _ = mockSendMoneyTxBody.GetBodyBytes()
	mockTransaction             = &model.Transaction{
		ID:      1,
		BlockID: 1,
		Height:  0,
		SenderAccountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
			72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
		RecipientAccountAddress: []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126,
			202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
		TransactionType:       1,
		Fee:                   10,
		Timestamp:             1000,
		TransactionHash:       []byte{},
		TransactionBodyLength: 8,
		TransactionBodyBytes:  mockSendMoneyTxBodyBytes,
		Signature:             []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Version:               1,
		TransactionIndex:      1,
	}
	mockTransactionExpired = &model.Transaction{
		ID:      1,
		BlockID: 1,
		Height:  12,
		SenderAccountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
			72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
		RecipientAccountAddress: []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126,
			202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
		TransactionType:       1,
		Fee:                   10,
		Timestamp:             1000,
		TransactionHash:       []byte{},
		TransactionBodyLength: 8,
		TransactionBodyBytes:  mockSendMoneyTxBodyBytes,
		Signature:             []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Version:               1,
		TransactionIndex:      1,
	}

	mockAccountBalance = &model.AccountBalance{
		AccountAddress:   mockTransaction.GetSenderAccountAddress(),
		BlockHeight:      1,
		SpendableBalance: 10000000000,
		Balance:          10000000000,
		PopRevenue:       0,
		Latest:           true,
	}
)

// mockTypeAction
func (*mockTypeAction) ApplyConfirmed(int64) error {
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
	db, _, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT blocksmith_public_key, pop_change, block_height, blocksmith_index FROM skipped_blocksmith WHERE block_height = 1":
		return nil, errors.New("MockedError")
	default:
		return nil, errors.New("mockQueryExecutorNotFound - InvalidQuery")
	}
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

func (*mockQueryExecutorSuccess) ExecuteTransaction(string, ...interface{}) error {
	return nil
}
func (*mockQueryExecutorSuccess) ExecuteTransactions([][]interface{}) error {
	return nil
}
func (*mockQueryExecutorSuccess) CommitTx() error { return nil }

func (*mockQueryExecutorSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	switch qStr {
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "NodePublicKey", "AccountAddress", "RegistrationHeight", "LockedBalance", "RegistrationStatus",
			"Latest", "Height",
		}).AddRow(1, bcsNodePubKey1, bcsAddress1, 10, "10.10.10.1", 100000000, uint32(model.NodeRegistrationState_NodeQueued), true, 100))
	case "SELECT id, block_height, tree, timestamp FROM merkle_tree ORDER BY timestamp DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "BlockHeight", "Tree", "Timestamp",
		}))
	}
	row := db.QueryRow(qStr)
	return row, nil
}

func (*mockQueryExecutorSuccess) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	transactionUtil := &transaction.Util{}
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE id = ? AND latest=1":
		for idx, arg := range args {
			if idx == 0 {
				nodeID := fmt.Sprintf("%d", arg)
				switch nodeID {
				case "1":
					mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "locked_balance", "registration_status", "latest", "height",
					}).AddRow(1, bcsNodePubKey1, bcsAddress1, 10, 100000000, uint32(model.NodeRegistrationState_NodeRegistered), true, 100))
				case "2":
					mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "locked_balance", "registration_status", "latest", "height",
					}).AddRow(2, bcsNodePubKey2, bcsAddress2, 20, 100000000, uint32(model.NodeRegistrationState_NodeRegistered), true, 200))
				case "3":
					mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "locked_balance", "registration_status", "latest", "height",
					}).AddRow(3, bcsNodePubKey3, bcsAddress3, 30, 100000000, uint32(model.NodeRegistrationState_NodeRegistered), true, 300))
				case "4":
					mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "locked_balance", "registration_status", "latest", "height",
					}).AddRow(3, mockGoodBlock.BlocksmithPublicKey, bcsAddress3, 30, 100000000,
						uint32(model.NodeRegistrationState_NodeRegistered), true, 300))
				}
			}
		}
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND height <= ? " +
		"ORDER BY height DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
			"account_address", "registration_height", "locked_balance", "registration_status", "latest", "height",
		}).AddRow(1, bcsNodePubKey1, bcsAddress1, 10, 100000000, uint32(model.NodeRegistrationState_NodeQueued), true, 100))
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version, merkle_root, " +
		"merkle_tree, reference_block_height FROM main_block WHERE height = 0":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "BlockHash", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
			"Version"},
		).AddRow(1, []byte{}, []byte{}, 1, 10000, []byte{}, []byte{}, "", 2, []byte{}, bcsNodePubKey1, 0, 0, 0, 1))
	case fmt.Sprintf("SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, "+
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version, "+
		"merkle_root, merkle_tree, reference_block_height FROM main_block WHERE height = %d", mockBlockData.GetHeight()+1):
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "BlockHash", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
			"Version"},
		))
	case "SELECT A.node_id, A.score, A.latest, A.height FROM participation_score as A INNER JOIN node_registry as B " +
		"ON A.node_id = B.id WHERE B.node_public_key=? AND B.latest=1 AND B.registration_status=0 AND A.latest=1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"node_id",
			"score",
			"latest",
			"height",
		},
		).AddRow(-1, 100000, true, 0))
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version, merkle_root, " +
		"merkle_tree, reference_block_height FROM main_block ORDER BY height DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).
			WillReturnRows(sqlmock.NewRows(
				query.NewBlockQuery(&chaintype.MainChain{}).Fields,
			).AddRow(
				mockBlockData.GetHeight(),
				mockBlockData.GetID(),
				mockBlockData.GetBlockHash(),
				mockBlockData.GetPreviousBlockHash(),
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
				mockBlockData.GetMerkleRoot(),
				mockBlockData.GetMerkleTree(),
				mockBlockData.GetReferenceBlockHeight(),
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
			mockPublishedReceipt[0].Receipt.SenderPublicKey,
			mockPublishedReceipt[0].Receipt.RecipientPublicKey,
			mockPublishedReceipt[0].Receipt.DatumType,
			mockPublishedReceipt[0].Receipt.DatumHash,
			mockPublishedReceipt[0].Receipt.ReferenceBlockHeight,
			mockPublishedReceipt[0].Receipt.ReferenceBlockHash,
			mockPublishedReceipt[0].Receipt.RMRLinked,
			mockPublishedReceipt[0].Receipt.RecipientSignature,
			mockPublishedReceipt[0].IntermediateHashes,
			mockPublishedReceipt[0].BlockHeight,
			mockPublishedReceipt[0].ReceiptIndex,
			mockPublishedReceipt[0].PublishedIndex,
		))
	case "SELECT id, node_public_key, account_address, registration_height, " +
		"locked_balance, registration_status, latest, height " +
		"FROM node_registry where registration_status = 0 AND (id,height) in " +
		"(SELECT id,MAX(height) FROM node_registry WHERE height <= 0 GROUP BY id) " +
		"ORDER BY height DESC":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id", "node_public_key", "account_address", "registration_height", "locked_balance",
			"registration_status", "latest", "height",
		}))
	case "SELECT id, node_public_key, account_address, registration_height, " +
		"locked_balance, registration_status, latest, height " +
		"FROM node_registry where registration_status = 0 AND (id,height) in " +
		"(SELECT id,MAX(height) FROM node_registry WHERE height <= 1 GROUP BY id) " +
		"ORDER BY height DESC":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id", "node_public_key", "account_address", "registration_height", "locked_balance",
			"registration_status", "latest", "height",
		}))
	case "SELECT account_address,block_height,spendable_balance,balance,pop_revenue,latest " +
		"FROM account_balance WHERE account_address = ? AND latest = 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"account_address", "block_height", "spendable_balance", "balance", "pop_revenue", "latest",
		}).AddRow(
			mockAccountBalance.GetAccountAddress(),
			mockAccountBalance.GetBlockHeight(),
			mockAccountBalance.GetSpendableBalance(),
			mockAccountBalance.GetBalance(),
			mockAccountBalance.GetPopRevenue(),
			mockAccountBalance.GetLatest(),
		))
	case "SELECT id, block_height, fee_per_byte, arrival_timestamp, transaction_bytes, " +
		"sender_account_address, recipient_account_address FROM mempool WHERE id IN (?)  ":
		txBytes, _ := transactionUtil.GetTransactionBytes(mockTransaction, true)
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id", "block_height", "fee_per_byte", "arrival_timestamp", "transaction_bytes",
			"sender_account_address", "recipient_account_address",
		}).AddRow(
			mockTransaction.GetID(),
			mockTransaction.GetHeight(),
			mockTransaction.GetFee(),
			1,
			txBytes,
			mockTransaction.GetSenderAccountAddress(),
			mockTransaction.GetRecipientAccountAddress(),
		))
	case "SELECT sender_address, transaction_hash, transaction_bytes, status, block_height, latest " +
		"FROM pending_transaction WHERE block_height = ? AND status = ? AND latest = ?":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(mock.NewRows(query.NewPendingTransactionQuery().Fields))
	case "SELECT id, sender_address, recipient_address, amount, applied_time, complete_minutes, status," +
		" block_height, latest FROM liquid_payment_transaction WHERE applied_time+(complete_minutes*60) <= ? AND status = ? AND latest = ?":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(mock.NewRows(query.NewLiquidPaymentTransactionQuery().Fields))
	// which is escrow expiration process
	default:
		mockRows := sqlmock.NewRows(query.NewEscrowTransactionQuery().Fields)
		mockRows.AddRow(
			int64(1),
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			"BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
			"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			int64(10),
			int64(1),
			uint64(120),
			model.EscrowStatus_Approved,
			uint32(0),
			true,
			"",
		)
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(mockRows)
	}
	rows, _ := db.Query(qe)
	return rows, nil
}

var mockPublishedReceipt = []*model.PublishedReceipt{
	{
		Receipt: &model.Receipt{
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
			PayloadHash: []byte{167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160, 97, 214, 98, 245, 128,
				255, 77, 228, 59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74},
			PayloadLength:  0,
			BlockSignature: []byte{},
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
		Chaintype                   chaintype.ChainType
		QueryExecutor               query.ExecutorInterface
		BlockQuery                  query.BlockQueryInterface
		MempoolQuery                query.MempoolQueryInterface
		TransactionQuery            query.TransactionQueryInterface
		PublishedReceiptQuery       query.PublishedReceiptQueryInterface
		SkippedBlocksmithQuery      query.SkippedBlocksmithQueryInterface
		Signature                   crypto.SignatureInterface
		MempoolService              MempoolServiceInterface
		ReceiptService              ReceiptServiceInterface
		NodeRegistrationService     NodeRegistrationServiceInterface
		NodeAddressInfoService      NodeAddressInfoServiceInterface
		BlocksmithService           BlocksmithServiceInterface
		FeeScaleService             fee.FeeScaleServiceInterface
		ActionTypeSwitcher          transaction.TypeActionSwitcher
		AccountBalanceQuery         query.AccountBalanceQueryInterface
		ParticipationScoreQuery     query.ParticipationScoreQueryInterface
		NodeRegistrationQuery       query.NodeRegistrationQueryInterface
		AccountLedgerQuery          query.AccountLedgerQueryInterface
		FeeVoteRevealVoteQuery      query.FeeVoteRevealVoteQueryInterface
		BlocksmithStrategy          strategy.BlocksmithStrategyInterface
		BlockIncompleteQueueService BlockIncompleteQueueServiceInterface
		BlockPoolService            BlockPoolServiceInterface
		Observer                    *observer.Observer
		Logger                      *log.Logger
		TransactionUtil             transaction.UtilInterface
		ReceiptUtil                 coreUtil.ReceiptUtilInterface
		PublishedReceiptUtil        coreUtil.PublishedReceiptUtilInterface
		TransactionCoreService      TransactionCoreServiceInterface
		CoinbaseService             CoinbaseServiceInterface
		ParticipationScoreService   ParticipationScoreServiceInterface
		PublishedReceiptService     PublishedReceiptServiceInterface
		PruneQuery                  []query.PruneQuery
		BlockStateStorage           storage.CacheStorageInterface
		BlockchainStatusService     BlockchainStatusServiceInterface
		ScrambleNodeService         ScrambleNodeServiceInterface
	}
	type args struct {
		version              uint32
		previousBlockHash    []byte
		blockSeed            []byte
		blockSmithPublicKey  []byte
		merkleRoot           []byte
		merkleTree           []byte
		previousBlockHeight  uint32
		referenceBlockHeight uint32
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
				version:              1,
				previousBlockHash:    []byte{},
				blockSeed:            []byte{},
				blockSmithPublicKey:  bcsNodePubKey1,
				merkleRoot:           []byte{},
				merkleTree:           []byte{},
				previousBlockHeight:  0,
				referenceBlockHeight: 0,
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
				MerkleRoot:           []byte{},
				MerkleTree:           []byte{},
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
				BlockHash: []byte{222, 81, 44, 228, 147, 156, 145, 104, 1, 97, 62, 138, 253, 90, 55, 41,
					29, 150, 230, 196, 68, 216, 14, 244, 224, 161, 132, 157, 229, 68, 33, 147},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:                   tt.fields.Chaintype,
				QueryExecutor:               tt.fields.QueryExecutor,
				BlockQuery:                  tt.fields.BlockQuery,
				MempoolQuery:                tt.fields.MempoolQuery,
				TransactionQuery:            tt.fields.TransactionQuery,
				PublishedReceiptQuery:       tt.fields.PublishedReceiptQuery,
				SkippedBlocksmithQuery:      tt.fields.SkippedBlocksmithQuery,
				Signature:                   tt.fields.Signature,
				MempoolService:              tt.fields.MempoolService,
				ReceiptService:              tt.fields.ReceiptService,
				NodeRegistrationService:     tt.fields.NodeRegistrationService,
				NodeAddressInfoService:      tt.fields.NodeAddressInfoService,
				BlocksmithService:           tt.fields.BlocksmithService,
				FeeScaleService:             tt.fields.FeeScaleService,
				ActionTypeSwitcher:          tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:         tt.fields.AccountBalanceQuery,
				ParticipationScoreQuery:     tt.fields.ParticipationScoreQuery,
				NodeRegistrationQuery:       tt.fields.NodeRegistrationQuery,
				AccountLedgerQuery:          tt.fields.AccountLedgerQuery,
				FeeVoteRevealVoteQuery:      tt.fields.FeeVoteRevealVoteQuery,
				BlocksmithStrategy:          tt.fields.BlocksmithStrategy,
				BlockIncompleteQueueService: tt.fields.BlockIncompleteQueueService,
				BlockPoolService:            tt.fields.BlockPoolService,
				Observer:                    tt.fields.Observer,
				Logger:                      tt.fields.Logger,
				TransactionUtil:             tt.fields.TransactionUtil,
				ReceiptUtil:                 tt.fields.ReceiptUtil,
				PublishedReceiptUtil:        tt.fields.PublishedReceiptUtil,
				TransactionCoreService:      tt.fields.TransactionCoreService,
				CoinbaseService:             tt.fields.CoinbaseService,
				ParticipationScoreService:   tt.fields.ParticipationScoreService,
				PublishedReceiptService:     tt.fields.PublishedReceiptService,
				PruneQuery:                  tt.fields.PruneQuery,
				BlockStateStorage:           tt.fields.BlockStateStorage,
				BlockchainStatusService:     tt.fields.BlockchainStatusService,
				ScrambleNodeService:         tt.fields.ScrambleNodeService,
			}
			got, err := bs.NewGenesisBlock(
				tt.args.version,
				tt.args.previousBlockHash,
				tt.args.blockSeed,
				tt.args.blockSmithPublicKey,
				tt.args.merkleRoot,
				tt.args.merkleTree,
				tt.args.previousBlockHeight,
				tt.args.referenceBlockHeight,
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
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.NewGenesisBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.NewGenesisBlock() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

var (
	mockBlocksmiths = []*model.Blocksmith{
		{
			NodePublicKey: bcsNodePubKey1,
			NodeID:        2,
			Score:         new(big.Int).SetInt64(1000),
		},
		{
			NodePublicKey: bcsNodePubKey2,
			NodeID:        3,
			Score:         new(big.Int).SetInt64(2000),
		},
		{
			NodePublicKey: mockBlockData.BlocksmithPublicKey,
			NodeID:        4,
			Score:         new(big.Int).SetInt64(3000),
		},
	}
	mockBlockPushBlock = model.Block{
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
func (*mockBlocksmithServicePushBlock) SortBlocksmiths(block *model.Block, withLock bool) {
}

func (*mockBlocksmithServicePushBlock) IsBlockTimestampValid(blocksmithIndex, numberOfBlocksmiths int64, previousBlock,
	currentBlock *model.Block) error {
	return nil
}

type (
	mockBlockPoolServiceNoDuplicate struct {
		BlockPoolService
	}
	mockBlockPoolServiceDuplicate struct {
		BlockPoolService
	}
	mockBlockPoolServiceDuplicateCorrectBlockHash struct {
		BlockPoolService
	}
)

func (*mockBlockPoolServiceDuplicate) GetBlock(index int64) *model.Block {
	return &model.Block{}
}

func (*mockBlockPoolServiceDuplicateCorrectBlockHash) GetBlock(index int64) *model.Block {
	return mockGoodIncomingBlock
}

func (*mockBlockPoolServiceNoDuplicate) GetBlock(index int64) *model.Block {
	return nil
}

type (
	mockPushBlockCoinbaseLotteryWinnersSuccess struct {
		CoinbaseService
	}
	mockPushBlockBlocksmithServiceSuccess struct {
		BlocksmithService
	}
	mockPushBlockPublishedReceiptServiceSuccess struct {
		PublishedReceiptService
	}
	mockBlockchainStatusService struct {
		BlockchainStatusService
	}
	mockPushBlockNodeAddressInfoServiceSuccess struct {
		NodeAddressInfoServiceInterface
	}
)

func (*mockBlockchainStatusService) SetLastBlock(block *model.Block, ct chaintype.ChainType) {}

func (*mockPushBlockCoinbaseLotteryWinnersSuccess) CoinbaseLotteryWinners(
	activeNodeRegistries []storage.NodeRegistry,
	scoreSum, blockTimestamp int64,
	previousBlock *model.Block,
) ([][]byte, error) {
	return make([][]byte, 0), nil
}

func (*mockPushBlockBlocksmithServiceSuccess) RewardBlocksmithAccountAddresses([][]byte, int64, int64, uint32) error {
	return nil
}

func (*mockPushBlockPublishedReceiptServiceSuccess) ProcessPublishedReceipts(block *model.Block) (int, error) {
	return 0, nil
}

func (*mockPushBlockNodeAddressInfoServiceSuccess) BeginCacheTransaction() error {
	return nil
}
func (*mockPushBlockNodeAddressInfoServiceSuccess) RollbackCacheTransaction() error {
	return nil
}
func (*mockPushBlockNodeAddressInfoServiceSuccess) CommitCacheTransaction() error {
	return nil
}

type (
	mockPushBlockFeeScaleServiceNoAdjust struct {
		fee.FeeScaleServiceInterface
	}
)

func (*mockPushBlockFeeScaleServiceNoAdjust) GetCurrentPhase(
	blockTimestamp int64,
	isPostTransaction bool,
) (phase model.FeeVotePhase, canAdjust bool, err error) {
	return model.FeeVotePhase_FeeVotePhaseCommmit, false, nil
}

type (
	mockMempoolServiceRemoveTransactionsSuccess struct {
		MempoolService
	}
)

func (*mockMempoolServiceRemoveTransactionsSuccess) RemoveMempoolTransactions(mempoolTxs []*model.Transaction) error {
	return nil
}

func (*mockMempoolServiceRemoveTransactionsSuccess) GetMempoolTransactions() (storage.MempoolMap, error) {
	return make(storage.MempoolMap), nil
}

type (
	mockScrambleNodeServicePushBlockBuildScrambleNodeSuccess struct {
		ScrambleNodeService
	}
	mockScrambleNodeServicePushBlockBuildScrambleNodeFail struct {
		ScrambleNodeService
	}
)

func (*mockScrambleNodeServicePushBlockBuildScrambleNodeSuccess) BuildScrambledNodes(block *model.Block) error {
	return nil
}

func (*mockScrambleNodeServicePushBlockBuildScrambleNodeSuccess) GetBlockHeightToBuildScrambleNodes(lastBlockHeight uint32) uint32 {
	return 1
}

func (*mockScrambleNodeServicePushBlockBuildScrambleNodeFail) BuildScrambledNodes(block *model.Block) error {
	return errors.New("mockedError")
}

func (*mockScrambleNodeServicePushBlockBuildScrambleNodeFail) GetBlockHeightToBuildScrambleNodes(lastBlockHeight uint32) uint32 {
	return 1
}

type (
	mockPendingTransactionServiceExpiringSuccess struct {
		PendingTransactionServiceInterface
	}
)

func (*mockPendingTransactionServiceExpiringSuccess) ExpiringPendingTransactions(
	blockHeight uint32, useTX bool,
) error {
	return nil
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
			mockBlockData.GetHeight(),
			mockBlockData.GetID(),
			mockBlockData.GetBlockHash(),
			mockBlockData.GetPreviousBlockHash(),
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
			mockBlockData.GetMerkleRoot(),
			mockBlockData.GetMerkleTree(),
			mockBlockData.GetReferenceBlockHeight(),
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
		mockBlockData.GetHeight(),
		mockBlockData.GetID(),
		mockBlockData.GetBlockHash(),
		mockBlockData.GetPreviousBlockHash(),
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
		mockBlockData.GetMerkleRoot(),
		mockBlockData.GetMerkleTree(),
		mockBlockData.GetReferenceBlockHeight(),
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

type (
	mockMempoolCacheStorageRemoveMempoolTransactionsSuccess struct {
		storage.CacheStorageInterface
	}
)

func (*mockMempoolCacheStorageRemoveMempoolTransactionsSuccess) RemoveItem(item interface{}) error {
	return nil
}

func TestMempoolService_RemoveMempoolTransactions(t *testing.T) {
	type fields struct {
		Chaintype           chaintype.ChainType
		QueryExecutor       query.ExecutorInterface
		MempoolQuery        query.MempoolQueryInterface
		Signature           crypto.SignatureInterface
		MempoolCacheStorage storage.CacheStorageInterface
		Logger              *log.Logger
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
				Chaintype:           &chaintype.MainChain{},
				MempoolQuery:        query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor:       &mockQueryExecutorSuccess{},
				MempoolCacheStorage: &mockMempoolCacheStorageRemoveMempoolTransactionsSuccess{},
				Logger:              log.New(),
			},
			args: args{
				transactions: []*model.Transaction{
					transaction.GetFixturesForTransaction(
						1562893303,
						[]byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
							72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
						[]byte{0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250, 58, 2, 83, 75,
							213, 137, 66, 236, 188, 43, 59, 241, 146, 243, 147, 58, 161, 35, 229, 54},
						false,
					),
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
					transaction.GetFixturesForTransaction(
						1562893303,
						[]byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
							72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
						[]byte{0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250, 58, 2, 83, 75,
							213, 137, 66, 236, 188, 43, 59, 241, 146, 243, 147, 58, 161, 35, 229, 54},
						false,
					),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &MempoolService{
				Chaintype:           tt.fields.Chaintype,
				QueryExecutor:       tt.fields.QueryExecutor,
				MempoolQuery:        tt.fields.MempoolQuery,
				Signature:           tt.fields.Signature,
				Logger:              tt.fields.Logger,
				MempoolCacheStorage: tt.fields.MempoolCacheStorage,
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

	mockGenerateBlockCoinbaseServiceSuccess struct {
		CoinbaseService
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
		transaction.GetFixturesForSignedMempoolTransaction(
			1,
			1562893305,
			[]byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
				72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
			[]byte{0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250, 58, 2, 83, 75,
				213, 137, 66, 236, 188, 43, 59, 241, 146, 243, 147, 58, 161, 35, 229, 54},
			false,
		).TransactionBytes),
	)
	return db.Query("")
}

// mockMempoolServiceSelectSuccess
func (*mockMempoolServiceSelectSuccess) SelectTransactionFromMempool() ([]*model.MempoolTransaction, error) {
	return []*model.MempoolTransaction{
		{
			FeePerByte: 1,
			TransactionBytes: transaction.GetFixturesForSignedMempoolTransaction(
				1,
				1562893305,
				[]byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
					72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				[]byte{0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250, 58, 2, 83, 75,
					213, 137, 66, 236, 188, 43, 59, 241, 146, 243, 147, 58, 161, 35, 229, 54},
				false,
			).TransactionBytes,
		},
	}, nil
}

// mockMempoolServiceSelectSuccess
func (*mockMempoolServiceSelectSuccess) SelectTransactionsFromMempool(blockTimestamp int64, blockHeight uint32) ([]*model.Transaction, error) {
	txByte := transaction.GetFixturesForSignedMempoolTransaction(
		1,
		1562893305,
		[]byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
			72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
		[]byte{0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250, 58, 2, 83, 75,
			213, 137, 66, 236, 188, 43, 59, 241, 146, 243, 147, 58, 161, 35, 229, 54},
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

// mockMempoolServiceSelectFail
func (*mockMempoolServiceSelectFail) SelectTransactionsFromMempool(blockTimestamp int64, blockHeight uint32) ([]*model.Transaction, error) {
	return nil, errors.New("want error on select")
}

// mockMempoolServiceSelectFail
func (*mockMempoolServiceSelectFail) GetMempoolTransactions() (storage.MempoolMap, error) {
	return make(storage.MempoolMap), nil
}

// mockMempoolServiceSelectSuccess
func (*mockMempoolServiceSelectWrongTransactionBytes) SelectTransactionsFromMempool(
	blockTimestamp int64,
	blockHeight uint32,
) ([]*model.Transaction, error) {
	return []*model.Transaction{
		{
			ID: 1,
		},
	}, nil
}

type (
	mockNodeRegistrationServiceGenerateBlockSuccess struct {
		NodeRegistrationServiceInterface
	}
)

func (*mockNodeRegistrationServiceGenerateBlockSuccess) GetActiveRegisteredNodes() ([]*model.NodeRegistration, error) {
	return make([]*model.NodeRegistration, 10), nil
}
func (*mockGenerateBlockCoinbaseServiceSuccess) GetCoinbase(
	blockTimesatamp, previousBlockTimesatamp int64,
) int64 {
	return 50 * constant.OneZBC
}

func TestBlockService_GenerateBlock(t *testing.T) {
	type fields struct {
		Chaintype               chaintype.ChainType
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		ReceiptService          ReceiptServiceInterface
		BlocksmithStrategy      strategy.BlocksmithStrategyInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		CoinbaseService         CoinbaseServiceInterface
		NodeRegistrationService NodeRegistrationServiceInterface
	}
	type args struct {
		previousBlock            *model.Block
		secretPhrase             string
		timestamp                int64
		blockSmithAccountAddress string
		empty                    bool
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
				Chaintype:       &chaintype.MainChain{},
				Signature:       &mockSignature{},
				MempoolQuery:    query.NewMempoolQuery(&chaintype.MainChain{}),
				MempoolService:  &mockMempoolServiceSelectFail{},
				CoinbaseService: &mockGenerateBlockCoinbaseServiceSuccess{},
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
				empty:                    false,
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
				BlocksmithStrategy:      &mockBlocksmithServicePushBlock{},
				ReceiptService:          &mockReceiptServiceReturnEmpty{},
				ActionTypeSwitcher:      &mockTypeActionSuccess{},
				CoinbaseService:         &mockGenerateBlockCoinbaseServiceSuccess{},
				NodeRegistrationService: &mockNodeRegistrationServiceGenerateBlockSuccess{},
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
				empty:        false,
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
				ReceiptService:          tt.fields.ReceiptService,
				BlocksmithStrategy:      tt.fields.BlocksmithStrategy,
				ActionTypeSwitcher:      tt.fields.ActionTypeSwitcher,
				ReceiptUtil:             &coreUtil.ReceiptUtil{},
				CoinbaseService:         tt.fields.CoinbaseService,
				NodeRegistrationService: tt.fields.NodeRegistrationService,
			}
			_, err := bs.GenerateBlock(
				tt.args.previousBlock,
				tt.args.secretPhrase,
				tt.args.timestamp,
				tt.args.empty,
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
	mockAddGenesisPublishedReceiptServiceSuccess struct {
		PublishedReceiptService
	}
	mockAddGenesisFeeScaleServiceCache struct {
		fee.FeeScaleServiceInterface
	}
	mockAddGenesisNodeAddressInfoServiceSuccess struct {
		NodeAddressInfoServiceInterface
	}
)

func (*mockAddGenesisNodeAddressInfoServiceSuccess) BeginCacheTransaction() error {
	return nil
}
func (*mockAddGenesisNodeAddressInfoServiceSuccess) RollbackCacheTransaction() error {
	return nil
}
func (*mockAddGenesisNodeAddressInfoServiceSuccess) CommitCacheTransaction() error {
	return nil
}

func (*mockAddGenesisFeeScaleServiceCache) GetCurrentPhase(
	blockTimestamp int64,
	isPostTransaction bool,
) (phase model.FeeVotePhase, canAdjust bool, err error) {
	return model.FeeVotePhase_FeeVotePhaseCommmit, false, nil
}

func (*mockAddGenesisFeeScaleServiceCache) GetLatestFeeScale(feeScale *model.FeeScale) error {
	*feeScale = model.FeeScale{
		FeeScale:    constant.OneZBC,
		BlockHeight: 0,
		Latest:      true,
	}
	return nil
}
func (*mockAddGenesisFeeScaleServiceCache) InsertFeeScale(feeScale *model.FeeScale) error {
	return nil
}
func (*mockBlocksmithServiceAddGenesisSuccess) SortBlocksmiths(block *model.Block, withLock bool) {

}

func (*mockAddGenesisPublishedReceiptServiceSuccess) ProcessPublishedReceipts(block *model.Block) (int, error) {
	return 0, nil
}

type (
	mockScrambleServiceAddGenesisSuccess struct {
		ScrambleNodeService
	}
)

func (*mockScrambleServiceAddGenesisSuccess) BuildScrambledNodes(block *model.Block) error {
	return nil
}

func (*mockScrambleServiceAddGenesisSuccess) GetBlockHeightToBuildScrambleNodes(lastBlockHeight uint32) uint32 {
	return 1
}

func TestBlockService_AddGenesis(t *testing.T) {
	type fields struct {
		Chaintype                 chaintype.ChainType
		QueryExecutor             query.ExecutorInterface
		BlockQuery                query.BlockQueryInterface
		MempoolQuery              query.MempoolQueryInterface
		TransactionQuery          query.TransactionQueryInterface
		AccountBalanceQuery       query.AccountBalanceQueryInterface
		Signature                 crypto.SignatureInterface
		MempoolService            MempoolServiceInterface
		ActionTypeSwitcher        transaction.TypeActionSwitcher
		Observer                  *observer.Observer
		NodeRegistrationService   NodeRegistrationServiceInterface
		NodeAddressInfoService    NodeAddressInfoServiceInterface
		BlocksmithStrategy        strategy.BlocksmithStrategyInterface
		BlockPoolService          BlockPoolServiceInterface
		Logger                    *log.Logger
		TransactionCoreService    TransactionCoreServiceInterface
		PublishedReceiptService   PublishedReceiptServiceInterface
		BlockStateStorage         storage.CacheStorageInterface
		BlocksStorage             storage.CacheStackStorageInterface
		BlockchainStatusService   BlockchainStatusServiceInterface
		ScrambleNodeService       ScrambleNodeServiceInterface
		PendingTransactionService PendingTransactionServiceInterface
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
				NodeAddressInfoService:  &mockAddGenesisNodeAddressInfoServiceSuccess{},
				BlocksmithStrategy:      &mockBlocksmithServiceAddGenesisSuccess{},
				BlockPoolService:        &mockBlockPoolServiceNoDuplicate{},
				Logger:                  log.New(),
				TransactionCoreService: NewTransactionCoreService(
					log.New(),
					&mockQueryExecutorSuccess{},
					&transaction.TypeSwitcher{
						Executor:                   &mockQueryExecutorSuccess{},
						ActiveNodeRegistryStorage:  &mockNodeRegistryCacheAlwaysSuccess{},
						PendingNodeRegistryStorage: &mockNodeRegistryCacheAlwaysSuccess{},
					},
					&transaction.Util{},
					query.NewTransactionQuery(&chaintype.MainChain{}),
					query.NewEscrowTransactionQuery(),
					query.NewLiquidPaymentTransactionQuery(),
				),
				PublishedReceiptService:   &mockAddGenesisPublishedReceiptServiceSuccess{},
				BlockStateStorage:         storage.NewBlockStateStorage(),
				BlocksStorage:             storage.NewBlocksStorage(),
				BlockchainStatusService:   &mockBlockchainStatusService{},
				ScrambleNodeService:       &mockScrambleServiceAddGenesisSuccess{},
				PendingTransactionService: &mockPendingTransactionServiceExpiringSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:                 tt.fields.Chaintype,
				QueryExecutor:             tt.fields.QueryExecutor,
				BlockQuery:                tt.fields.BlockQuery,
				MempoolQuery:              tt.fields.MempoolQuery,
				AccountBalanceQuery:       tt.fields.AccountBalanceQuery,
				TransactionQuery:          tt.fields.TransactionQuery,
				Signature:                 tt.fields.Signature,
				MempoolService:            tt.fields.MempoolService,
				ActionTypeSwitcher:        tt.fields.ActionTypeSwitcher,
				Observer:                  tt.fields.Observer,
				NodeRegistrationService:   tt.fields.NodeRegistrationService,
				NodeAddressInfoService:    tt.fields.NodeAddressInfoService,
				BlocksmithStrategy:        tt.fields.BlocksmithStrategy,
				BlockPoolService:          tt.fields.BlockPoolService,
				Logger:                    tt.fields.Logger,
				TransactionCoreService:    tt.fields.TransactionCoreService,
				PublishedReceiptService:   tt.fields.PublishedReceiptService,
				FeeScaleService:           &mockAddGenesisFeeScaleServiceCache{},
				BlockStateStorage:         tt.fields.BlockStateStorage,
				BlocksStorage:             tt.fields.BlocksStorage,
				BlockchainStatusService:   tt.fields.BlockchainStatusService,
				ScrambleNodeService:       tt.fields.ScrambleNodeService,
				PendingTransactionService: tt.fields.PendingTransactionService,
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
		mockBlockData.GetHeight(),
		mockBlockData.GetID(),
		mockBlockData.GetBlockHash(),
		mockBlockData.GetPreviousBlockHash(),
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
		mockBlockData.GetMerkleRoot(),
		mockBlockData.GetMerkleTree(),
		mockBlockData.GetReferenceBlockHeight(),
	))
	return db.Query("")
}

func (*mockQueryExecutorCheckGenesisTrue) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.MainChain{}).Fields,
		).AddRow(
			mockBlockData.GetHeight(),
			mockBlockData.GetID(),
			mockBlockData.GetBlockHash(),
			mockBlockData.GetPreviousBlockHash(),
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
			mockBlockData.GetMerkleRoot(),
			mockBlockData.GetMerkleTree(),
			mockBlockData.GetReferenceBlockHeight(),
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
			if got, _ := bs.CheckGenesis(); got != tt.want {
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

func (*mockQueryExecutorGetBlockByHeightSuccess) ExecuteSelectRow(qStr string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
		query.NewBlockQuery(&chaintype.MainChain{}).Fields).AddRow(
		mockBlockData.GetHeight(),
		mockBlockData.GetID(),
		mockBlockData.GetBlockHash(),
		mockBlockData.GetPreviousBlockHash(),
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
		mockBlockData.GetMerkleRoot(),
		mockBlockData.GetMerkleTree(),
		mockBlockData.GetReferenceBlockHeight(),
	))
	return db.QueryRow(qStr), nil
}
func (*mockQueryExecutorGetBlockByHeightSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	switch qStr {
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height FROM main_block WHERE height = 0":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.MainChain{}).Fields).AddRow(
			mockBlockData.GetHeight(),
			mockBlockData.GetID(),
			mockBlockData.GetBlockHash(),
			mockBlockData.GetPreviousBlockHash(),
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
			mockBlockData.GetMerkleRoot(),
			mockBlockData.GetMerkleTree(),
			mockBlockData.GetReferenceBlockHeight(),
		))
	case "SELECT id, block_id, block_height, sender_account_address, recipient_account_address, transaction_type, " +
		"fee, timestamp, transaction_hash, transaction_body_length, transaction_body_bytes, " +
		"signature, version, transaction_index FROM \"transaction\" WHERE block_id = ? ORDER BY transaction_index ASC":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
			query.NewTransactionQuery(&chaintype.MainChain{}).Fields))
	}
	return db.Query(qStr)
}

func (*mockQueryExecutorGetBlockByHeightFail) ExecuteSelectRow(qStr string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnRows(mock.NewRows(query.NewBlockQuery(&chaintype.MainChain{}).Fields))
	return db.QueryRow(qStr), nil
}

func (*mockQueryExecutorGetBlockByHeightFail) ExecuteSelect(string, bool, ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}

type (
	// GetBlockByHeight mocks
	mockGetBlockByHeightTransactionCoreServiceSuccess struct {
		TransactionCoreService
	}
	// GetBlockByHeight mocks
	mockGetBlockByHeightPublishedReceiptUtilSuccess struct {
		coreUtil.PublishedReceiptUtilInterface
	}
)

var (
	// GetBlockByHeight mocks
	mockGetBlockByHeightResult = model.Block{
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
		Transactions:  make([]*model.Transaction, 0),
	}
	// GetBlockByHeight mocks
)

func (*mockGetBlockByHeightTransactionCoreServiceSuccess) GetTransactionsByBlockID(blockID int64) ([]*model.Transaction, error) {
	return make([]*model.Transaction, 0), nil
}

func (*mockGetBlockByHeightPublishedReceiptUtilSuccess) GetPublishedReceiptsByBlockHeight(blockHeight uint32) ([]*model.PublishedReceipt, error) {
	return nil, nil
}

func TestBlockService_GetBlockByHeight(t *testing.T) {
	type fields struct {
		Chaintype              chaintype.ChainType
		QueryExecutor          query.ExecutorInterface
		BlockQuery             query.BlockQueryInterface
		MempoolQuery           query.MempoolQueryInterface
		TransactionQuery       query.TransactionQueryInterface
		Signature              crypto.SignatureInterface
		MempoolService         MempoolServiceInterface
		ActionTypeSwitcher     transaction.TypeActionSwitcher
		AccountBalanceQuery    query.AccountBalanceQueryInterface
		TransactionCoreService TransactionCoreServiceInterface
		PublishedReceiptUtil   coreUtil.PublishedReceiptUtilInterface
		Observer               *observer.Observer
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
				Chaintype:              &chaintype.MainChain{},
				QueryExecutor:          &mockQueryExecutorGetBlockByHeightSuccess{},
				BlockQuery:             query.NewBlockQuery(&chaintype.MainChain{}),
				TransactionQuery:       query.NewTransactionQuery(&chaintype.MainChain{}),
				TransactionCoreService: &mockGetBlockByHeightTransactionCoreServiceSuccess{},
				PublishedReceiptUtil:   &mockGetBlockByHeightPublishedReceiptUtilSuccess{},
			},
			want:    &mockGetBlockByHeightResult,
			wantErr: false,
		},
		{
			name: "GetBlockByHeight:FailNoEntryFound", // All is good
			fields: fields{
				Chaintype:              &chaintype.MainChain{},
				QueryExecutor:          &mockQueryExecutorGetBlockByHeightFail{},
				BlockQuery:             query.NewBlockQuery(&chaintype.MainChain{}),
				TransactionQuery:       query.NewTransactionQuery(&chaintype.MainChain{}),
				TransactionCoreService: &mockGetBlockByHeightTransactionCoreServiceSuccess{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:              tt.fields.Chaintype,
				QueryExecutor:          tt.fields.QueryExecutor,
				BlockQuery:             tt.fields.BlockQuery,
				MempoolQuery:           tt.fields.MempoolQuery,
				TransactionQuery:       tt.fields.TransactionQuery,
				Signature:              tt.fields.Signature,
				MempoolService:         tt.fields.MempoolService,
				ActionTypeSwitcher:     tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:    tt.fields.AccountBalanceQuery,
				Observer:               tt.fields.Observer,
				TransactionCoreService: tt.fields.TransactionCoreService,
				PublishedReceiptUtil:   tt.fields.PublishedReceiptUtil,
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
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, " +
		"version, merkle_root, merkle_tree, reference_block_height FROM main_block WHERE id = 1":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.MainChain{}).Fields).AddRow(
			mockBlockData.GetHeight(),
			mockBlockData.GetID(),
			mockBlockData.GetBlockHash(),
			mockBlockData.GetPreviousBlockHash(),
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
			mockBlockData.GetMerkleRoot(),
			mockBlockData.GetMerkleTree(),
			mockBlockData.GetReferenceBlockHeight(),
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
			mockBlockData.GetHeight(),
			mockBlockData.GetID(),
			mockBlockData.GetBlockHash(),
			mockBlockData.GetPreviousBlockHash(),
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
			mockBlockData.GetMerkleRoot(),
			mockBlockData.GetMerkleTree(),
			mockBlockData.GetReferenceBlockHeight(),
		))
	return db.QueryRow(qStr), nil
}

func (*mockQueryExecutorGetBlockByIDFail) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, errors.New("MockedError")
}

type (
	// GetBlockByID mocks
	mockGetBlockByIDTransactionCoreServiceSuccess struct {
		TransactionCoreService
	}
	// GetBlockByID mocks
	mockGetBlockByIDPublishedReceiptUtilSuccess struct {
		coreUtil.PublishedReceiptUtilInterface
	}
)

var (
	// GetBlockByID mocks
	mockGetBlockByIDResult = model.Block{
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
		Transactions:  make([]*model.Transaction, 0),
	}
	// GetBlockByID mocks
)

func (*mockGetBlockByIDTransactionCoreServiceSuccess) GetTransactionsByBlockID(blockID int64) ([]*model.Transaction, error) {
	return make([]*model.Transaction, 0), nil
}

func (*mockGetBlockByIDPublishedReceiptUtilSuccess) GetPublishedReceiptsByBlockHeight(blockHeight uint32) ([]*model.PublishedReceipt, error) {
	return nil, nil
}

func TestBlockService_GetBlockByID(t *testing.T) {
	type fields struct {
		Chaintype              chaintype.ChainType
		QueryExecutor          query.ExecutorInterface
		BlockQuery             query.BlockQueryInterface
		MempoolQuery           query.MempoolQueryInterface
		TransactionQuery       query.TransactionQueryInterface
		Signature              crypto.SignatureInterface
		MempoolService         MempoolServiceInterface
		ActionTypeSwitcher     transaction.TypeActionSwitcher
		AccountBalanceQuery    query.AccountBalanceQueryInterface
		Observer               *observer.Observer
		TransactionCoreService TransactionCoreServiceInterface
		PublishedReceiptUtil   coreUtil.PublishedReceiptUtilInterface
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
				Chaintype:              &chaintype.MainChain{},
				QueryExecutor:          &mockQueryExecutorGetBlockByIDSuccess{},
				BlockQuery:             query.NewBlockQuery(&chaintype.MainChain{}),
				TransactionQuery:       query.NewTransactionQuery(&chaintype.MainChain{}),
				TransactionCoreService: &mockGetBlockByIDTransactionCoreServiceSuccess{},
				PublishedReceiptUtil:   &mockGetBlockByIDPublishedReceiptUtilSuccess{},
			},
			args: args{
				ID:               int64(1),
				withAttachedData: true,
			},
			want:    &mockGetBlockByIDResult,
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
				Chaintype:              tt.fields.Chaintype,
				QueryExecutor:          tt.fields.QueryExecutor,
				BlockQuery:             tt.fields.BlockQuery,
				MempoolQuery:           tt.fields.MempoolQuery,
				TransactionQuery:       tt.fields.TransactionQuery,
				Signature:              tt.fields.Signature,
				MempoolService:         tt.fields.MempoolService,
				ActionTypeSwitcher:     tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:    tt.fields.AccountBalanceQuery,
				Observer:               tt.fields.Observer,
				TransactionCoreService: tt.fields.TransactionCoreService,
				PublishedReceiptUtil:   tt.fields.PublishedReceiptUtil,
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
		mockBlockData.GetHeight(),
		mockBlockData.GetID(),
		mockBlockData.GetBlockHash(),
		mockBlockData.GetPreviousBlockHash(),
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
		mockBlockData.GetMerkleRoot(),
		mockBlockData.GetMerkleTree(),
		mockBlockData.GetReferenceBlockHeight(),
	).AddRow(
		mockBlockData.GetHeight(),
		mockBlockData.GetID(),
		mockBlockData.GetBlockHash(),
		mockBlockData.GetPreviousBlockHash(),
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
		mockBlockData.GetMerkleRoot(),
		mockBlockData.GetMerkleTree(),
		mockBlockData.GetReferenceBlockHeight(),
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
		withAttachedData   bool
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
				startHeight:      0,
				limit:            2,
				withAttachedData: false,
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
			got, err := bs.GetBlocksFromHeight(tt.args.startHeight, tt.args.limit, tt.args.withAttachedData)
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

type (
	mockReceiptServiceSuccess struct {
		ReceiptService
		WantDuplicated bool
	}
	mockReceiptServiceFail struct {
		ReceiptService
	}
	mockBlocksmithServiceReceiveBlock struct {
		strategy.BlocksmithStrategyMain
	}

	mockBlockIncompleteQueueServiceReceiveBlock struct {
		BlockIncompleteQueueService
	}

	mockQueryExecutorReceiveBlockFail struct {
		query.Executor
	}
	mockBlockStateStorageReceiveBlockFail struct {
		storage.CacheStorageInterface
	}
	mockBlockStateStorageReceiveBlockSuccess struct {
		storage.CacheStorageInterface
	}
)

func (*mockReceiptServiceSuccess) GenerateReceiptWithReminder(
	chaintype.ChainType, []byte, *storage.BlockCacheObject, []byte, string, uint32,
) (*model.Receipt, error) {
	return nil, nil
}

func (mrs *mockReceiptServiceSuccess) CheckDuplication([]byte, []byte) (err error) {
	if mrs.WantDuplicated {
		return blocker.NewBlocker(
			blocker.DuplicateReceiptErr,
			err.Error(),
		)
	}
	return nil
}

func (*mockReceiptServiceFail) GenerateBatchReceiptWithReminder(
	chaintype.ChainType, []byte, *model.Block, []byte, string, uint32,
) (*model.BatchReceipt, error) {
	return nil, errors.New("mockedErr")
}

func (*mockBlockStateStorageReceiveBlockFail) GetItem(lastChange, item interface{}) error {
	return errors.New("mockedErr")
}
func (*mockBlockStateStorageReceiveBlockSuccess) GetItem(lastChange, item interface{}) error {
	return nil
}

// mocks for ReceiveBlock tests
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
		BlockHash: []byte{
			167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160, 97, 214, 98, 245, 128, 255, 77,
			228, 59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74,
		},
	}

	mockBlockIDProcessQueueReceiveBlockAlreadyQueued int64 = 1
)

func (*mockBlocksmithServiceReceiveBlock) IsBlockTimestampValid(blocksmithIndex, numberOfBlocksmiths int64, previousBlock,
	currentBlock *model.Block) error {
	return nil
}

func (*mockBlockIncompleteQueueServiceReceiveBlock) GetBlockQueue(blockID int64) *model.Block {
	switch blockID {
	case mockBlockIDProcessQueueReceiveBlockAlreadyQueued:
		return &model.Block{
			ID: constant.MainchainGenesisBlockID,
		}

	default:
		return nil
	}
}

func (*mockBlockIncompleteQueueServiceReceiveBlock) AddBlockQueue(block *model.Block) {
}

func (*mockBlockIncompleteQueueServiceReceiveBlock) SetTransactionsRequired(blockIDs int64, requiredTxIDs TransactionIDsMap) {
}

func (*mockQueryExecutorReceiveBlockFail) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("Mock Error")
}
func (*mockQueryExecutorReceiveBlockFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mockRows := mock.NewRows([]string{"fake"})
	mockRows.AddRow("1")
	mock.ExpectQuery(qStr).WillReturnRows(mockRows)
	return db.QueryRow(qStr), nil
}

func (bss *mockBlocksmithServiceReceiveBlock) IsValidSmithTime(
	blocksmithIndex,
	numberOfBlocksmiths int64,
	previousBlock *model.Block,
) error {
	return nil
}

type (
	mockMempoolServiceGetMempoolTransactionSuccess struct {
		MempoolService
	}
)

func (*mockMempoolServiceGetMempoolTransactionSuccess) GetMempoolTransactions() (storage.MempoolMap, error) {
	return make(storage.MempoolMap), nil
}

type (
	mockNodeRegistryCacheAlwaysSuccess struct {
		storage.NodeRegistryCacheStorage
	}
)

func (*mockNodeRegistryCacheAlwaysSuccess) Begin() error {
	return nil
}

func (*mockNodeRegistryCacheAlwaysSuccess) Commit() error {
	return nil
}

func (*mockNodeRegistryCacheAlwaysSuccess) Rollback() error {
	return nil
}

func (*mockNodeRegistryCacheAlwaysSuccess) TxSetItem(idx, item interface{}) error {
	return nil
}

func (*mockNodeRegistryCacheAlwaysSuccess) TxSetItems(items interface{}) error {
	return nil
}

func (*mockNodeRegistryCacheAlwaysSuccess) TxRemoveItem(idx interface{}) error {
	return nil
}

func (*mockNodeRegistryCacheAlwaysSuccess) RemoveItem(idx interface{}) error {
	return nil
}

func (*mockNodeRegistryCacheAlwaysSuccess) GetItem(idx, item interface{}) error {
	return nil
}

func (*mockNodeRegistryCacheAlwaysSuccess) GetAllItems(items interface{}) error {
	return nil
}

func (*mockNodeRegistryCacheAlwaysSuccess) SetItem(idx, item interface{}) error {
	return nil
}

func (*mockNodeRegistryCacheAlwaysSuccess) SetItems(items interface{}) error {
	return nil
}

func TestBlockService_GenerateGenesisBlock(t *testing.T) {
	type fields struct {
		Chaintype               chaintype.ChainType
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
				Chaintype:        &chaintype.MainChain{},
				QueryExecutor:    nil,
				BlockQuery:       nil,
				MempoolQuery:     nil,
				TransactionQuery: nil,
				MerkleTreeQuery:  nil,
				Signature:        nil,
				MempoolService:   nil,
				ActionTypeSwitcher: &transaction.TypeSwitcher{
					ActiveNodeRegistryStorage:  &mockNodeRegistryCacheAlwaysSuccess{},
					PendingNodeRegistryStorage: &mockNodeRegistryCacheAlwaysSuccess{},
				},
				AccountBalanceQuery:     nil,
				ParticipationScoreQuery: nil,
				NodeRegistrationQuery:   nil,
				Observer:                nil,
			},
			args: args{
				genesisEntries: []constant.GenesisConfigEntry{
					{
						AccountAddress: "ZBC_TE5DFSAH_HVWOLTBQ_Y6IRKY35_JMYS25TB_3NIPF5DE_Q2IPMJMQ_2WD2R5BJ",
						AccountBalance: 0,
						NodePublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118,
							97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
						LockedBalance:      10000000000000,
						ParticipationScore: 1000000000,
					},
					{
						AccountAddress: "ZBC_AAHANWVK_GY6DEASC_QJ36F236_ZMCQZGGC_VKJCWP7A_MV77CPUY_XP7THV2L",
						AccountBalance: 0,
						NodePublicKey: []byte{0, 14, 6, 218, 170, 54, 60, 50, 2, 66, 130, 119, 226, 235, 126, 203, 5, 12, 152,
							194, 170, 146, 43, 63, 224, 101, 127, 241, 62, 152, 187, 255},
						LockedBalance:      0,
						ParticipationScore: 1000000000,
					},
					{
						AccountAddress: "ZBC_RRZSGM47_C3VMAJTI_MAMFARSW_2N5UQNG5_MF4TXF46_LKTRC3X5_PKPA44KK",
						AccountBalance: 0,
						NodePublicKey: []byte{140, 115, 35, 51, 159, 22, 234, 192, 38, 104, 96, 24, 80, 70, 86, 211, 123, 72, 52,
							221, 97, 121, 59, 151, 158, 90, 167, 17, 110, 253, 122, 158},
						LockedBalance:      0,
						ParticipationScore: 1000000000,
					},
					{
						AccountAddress: "ZBC_FHV3RVSG_C6MVS2BJ_7L4DGB2F_LHVLKZFD_FVCZQRRU_ZGJUMBXS_GTOFKGPU",
						AccountBalance: 100000000000,
						NodePublicKey: []byte{41, 235, 184, 214, 70, 23, 153, 89, 104, 41, 250, 248, 51, 7, 69, 89, 234, 181, 100,
							163, 45, 69, 152, 70, 52, 201, 147, 70, 6, 242, 52, 220},
						LockedBalance:      0,
						ParticipationScore: 1000000000,
					},
				},
			},
			wantErr: false,
			want:    -1404528444615386701,
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
			mockBlockData.GetHeight(),
			mockBlockData.GetID(),
			mockBlockData.GetBlockHash(),
			mockBlockData.GetPreviousBlockHash(),
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
			mockBlockData.GetMerkleRoot(),
			mockBlockData.GetMerkleTree(),
			mockBlockData.GetReferenceBlockHeight(),
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

func (*mockBlocksmithServiceValidateBlockSuccess) IsBlockTimestampValid(blocksmithIndex, numberOfBlocksmiths int64, previousBlock,
	currentBlock *model.Block) error {
	return nil
}

func (*mockBlocksmithServiceValidateBlockSuccess) IsValidSmithTime(blocksmithIndex, numberOfBlocksmiths int64,
	previousBlock *model.Block) error {
	return nil
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
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height FROM main_block WHERE id = 0":
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
	case "SELECT MAX(height), id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height FROM main_block":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields[:len(blockQ.Fields)-1]).AddRow(
				mockGoodBlock.GetHeight(),
				mockGoodBlock.GetID(),
				mockGoodBlock.GetBlockHash(),
				mockGoodBlock.GetPreviousBlockHash(),
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
				mockGoodBlock.GetTotalCoinBase(),
				mockGoodBlock.GetVersion(),
				mockGoodBlock.GetMerkleRoot(),
				mockGoodBlock.GetMerkleTree(),
				mockGoodBlock.GetReferenceBlockHeight(),
			),
		)
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height FROM main_block WHERE id = 1":
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
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height FROM main_block WHERE id = 0":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields))
	default:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields[:len(blockQ.Fields)-1]).AddRow(
				mockGoodBlock.GetHeight(),
				mockGoodBlock.GetID(),
				mockGoodBlock.GetBlockHash(),
				mockGoodBlock.GetPreviousBlockHash(),
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
				mockGoodBlock.GetMerkleRoot(),
				mockGoodBlock.GetMerkleTree(),
				mockGoodBlock.GetReferenceBlockHeight(),
			),
		)
	}

	return db.QueryRow(qStr), nil
}

func (*mockNodeRegistrationServiceBlockPopSuccess) ResetScrambledNodes() {

}

func (*mockNodeRegistrationServiceBlockPopSuccess) UpdateNextNodeAdmissionCache(
	newNextNodeAdmission *model.NodeAdmissionTimestamp,
) error {
	return nil
}

func (*mockNodeRegistrationServiceBlockPopSuccess) InitializeCache() error {
	return nil
}

func (*mockMempoolServiceBlockPopSuccess) GetMempoolTransactionsWantToBackup(height uint32) ([]*model.Transaction, error) {
	return make([]*model.Transaction, 0), nil
}

func (*mockMempoolServiceBlockPopSuccess) BackupMempools(commonBlock *model.Block) error {
	return nil
}

func (*mockMempoolServiceBlockPopFail) GetMempoolTransactionsWantToBackup(height uint32) ([]*model.Transaction, error) {
	return nil, errors.New("mockedError")
}

func (*mockMempoolServiceBlockPopFail) BackupMempools(commonBlock *model.Block) error {
	return errors.New("error BackupMempools")
}

func (*mockReceiptSuccess) GetPublishedReceiptsByHeight(blockHeight uint32) ([]*model.PublishedReceipt, error) {
	return make([]*model.PublishedReceipt, 0), nil
}
func (*mockReceiptSuccess) ClearCache() {

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
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height " +
		"FROM main_block WHERE height = 999":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields).AddRow(
				mockGoodCommonBlock.GetHeight(),
				mockGoodCommonBlock.GetID(),
				mockGoodCommonBlock.GetBlockHash(),
				mockGoodCommonBlock.GetPreviousBlockHash(),
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
				mockGoodCommonBlock.GetMerkleRoot(),
				mockGoodCommonBlock.GetMerkleTree(),
				mockGoodCommonBlock.GetReferenceBlockHeight(),
			),
		)
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height " +
		"FROM main_block WHERE id = 0":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields).AddRow(
				mockGoodCommonBlock.GetHeight(),
				mockGoodCommonBlock.GetID(),
				mockGoodCommonBlock.GetBlockHash(),
				mockGoodCommonBlock.GetPreviousBlockHash(),
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
				mockGoodCommonBlock.GetMerkleRoot(),
				mockGoodCommonBlock.GetMerkleTree(),
				mockGoodCommonBlock.GetReferenceBlockHeight(),
			),
		)
	case "SELECT id, block_id, block_height, sender_account_address, recipient_account_address, transaction_type, fee, " +
		"timestamp, transaction_hash, transaction_body_length, transaction_body_bytes, signature, version, " +
		"transaction_index FROM \"transaction\" WHERE block_id = ? ORDER BY transaction_index ASC":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(transactionQ.Fields))
	case "SELECT sender_address, transaction_hash, transaction_bytes, status, block_height, latest FROM pending_transaction " +
		"WHERE (block_height+?) = ? AND status = ? AND latest = ?":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(mock.NewRows(query.NewPendingTransactionQuery().Fields))
	}

	return db.Query(qStr)
}

func (*mockExecutorBlockPopSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	blockQ := query.NewBlockQuery(&chaintype.MainChain{})

	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
		sqlmock.NewRows(blockQ.Fields).AddRow(
			mockGoodBlock.GetHeight(),
			mockGoodBlock.GetID(),
			mockGoodBlock.GetBlockHash(),
			mockGoodBlock.GetPreviousBlockHash(),
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
			mockGoodBlock.GetMerkleRoot(),
			mockGoodBlock.GetMerkleTree(),
			mockGoodBlock.GetReferenceBlockHeight(),
		),
	)
	return db.QueryRow(qStr), nil
}

type (
	// PopOffToBlock mocks
	mockBlockPoolServicePopOffToBlockSuccess struct {
		BlockPoolService
	}
	mockPopOffToBlockTransactionCoreService struct {
		TransactionCoreService
	}
	// PopOffToBlock mocks
	mockPublishedReceiptUtilSuccess struct {
		coreUtil.PublishedReceiptUtil
	}
	mockPopOffToBlockNodeRegistrationServiceSucess struct {
		NodeRegistrationServiceInterface
	}
	mockPopOffToBlockBlockStateStorageFail struct {
		storage.CacheStorageInterface
	}
	mockPopOffToBlockBlockStateStorageSuccess struct {
		storage.CacheStorageInterface
	}
	mockPopOffToBlockBlocksStorageSuccess struct {
		storage.CacheStackStorageInterface
	}
	mockPopOffToBlockBlockStateStorageHardForkSuccess struct {
		storage.CacheStorageInterface
	}
)

func (*mockPopOffToBlockTransactionCoreService) GetTransactionsByBlockID(blockID int64) ([]*model.Transaction, error) {
	return make([]*model.Transaction, 0), nil
}

func (*mockBlockPoolServicePopOffToBlockSuccess) ClearBlockPool() {}

func (*mockPublishedReceiptUtilSuccess) GetPublishedReceiptsByBlockHeight(blockHeight uint32) ([]*model.PublishedReceipt, error) {
	return make([]*model.PublishedReceipt, 0), nil
}

func (*mockPopOffToBlockBlocksStorageSuccess) PopTo(uint32) error {
	return nil
}

type (
	mockedExecutorPopOffToBlockSuccessPopping struct {
		query.Executor
	}
)

func (*mockedExecutorPopOffToBlockSuccessPopping) BeginTx() error {
	return nil
}

func (*mockedExecutorPopOffToBlockSuccessPopping) CommitTx() error {
	return nil
}
func (*mockedExecutorPopOffToBlockSuccessPopping) ExecuteTransactions([][]interface{}) error {
	return nil
}
func (*mockedExecutorPopOffToBlockSuccessPopping) RollbackTx() error {
	return nil
}

func (*mockedExecutorPopOffToBlockSuccessPopping) ExecuteSelectRow(qStr string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	blockQ := query.NewBlockQuery(&chaintype.MainChain{})
	mockedRows := mock.NewRows(blockQ.Fields)
	switch {
	case strings.Contains(qStr, "WHERE height ="):
		mockedRows.AddRow(
			mockGoodBlock.GetHeight(),
			mockGoodBlock.GetID(),
			mockGoodBlock.GetBlockHash(),
			mockGoodBlock.GetPreviousBlockHash(),
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
			mockGoodBlock.GetMerkleRoot(),
			mockGoodBlock.GetMerkleTree(),
			mockGoodBlock.GetReferenceBlockHeight(),
		)
	default:
		mockedRows.AddRow(
			mockGoodBlock.GetHeight(),
			int64(100),
			mockGoodBlock.GetBlockHash(),
			mockGoodBlock.GetPreviousBlockHash(),
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
			mockGoodBlock.GetMerkleRoot(),
			mockGoodBlock.GetMerkleTree(),
			mockGoodBlock.GetReferenceBlockHeight(),
		)

	}
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(mockedRows)
	return db.QueryRow(qStr), nil
}

func (*mockPopOffToBlockNodeRegistrationServiceSucess) UpdateNextNodeAdmissionCache(
	newNextNodeAdmission *model.NodeAdmissionTimestamp) error {
	return nil
}

func (*mockPopOffToBlockBlockStateStorageFail) GetItem(lastChange, item interface{}) error {
	return errors.New("mockedErr")
}
func (*mockPopOffToBlockBlockStateStorageSuccess) GetItem(lastChange, item interface{}) error {
	var blockCopy, _ = item.(*model.Block)
	*blockCopy = *mockGoodCommonBlock
	return nil
}

func (*mockPopOffToBlockBlockStateStorageSuccess) SetItem(lastChange, item interface{}) error {
	return nil
}

func (*mockPopOffToBlockBlockStateStorageHardForkSuccess) GetItem(lastChange, item interface{}) error {
	var blockCopy, _ = item.(*model.Block)
	*blockCopy = *mockGoodCommonBlock
	return nil
}

func (*mockPopOffToBlockBlockStateStorageHardForkSuccess) SetItem(lastChange, item interface{}) error {
	return nil
}

type (
	mockScrambleServicePopOffToBlockSuccess struct {
		ScrambleNodeService
	}
)

func (*mockScrambleServicePopOffToBlockSuccess) PopOffScrambleToHeight(height uint32) error {
	return nil
}

func TestBlockService_PopOffToBlock(t *testing.T) {
	var mockPopedBlock = mockGoodBlock
	mockPopedBlock.ID = 100
	mockPopedBlock.Height = 100
	type fields struct {
		Chaintype                   chaintype.ChainType
		QueryExecutor               query.ExecutorInterface
		BlockQuery                  query.BlockQueryInterface
		MempoolQuery                query.MempoolQueryInterface
		TransactionQuery            query.TransactionQueryInterface
		PublishedReceiptQuery       query.PublishedReceiptQueryInterface
		SkippedBlocksmithQuery      query.SkippedBlocksmithQueryInterface
		Signature                   crypto.SignatureInterface
		MempoolService              MempoolServiceInterface
		ReceiptService              ReceiptServiceInterface
		NodeRegistrationService     NodeRegistrationServiceInterface
		BlocksmithService           BlocksmithServiceInterface
		ActionTypeSwitcher          transaction.TypeActionSwitcher
		AccountBalanceQuery         query.AccountBalanceQueryInterface
		ParticipationScoreQuery     query.ParticipationScoreQueryInterface
		NodeRegistrationQuery       query.NodeRegistrationQueryInterface
		AccountLedgerQuery          query.AccountLedgerQueryInterface
		BlocksmithStrategy          strategy.BlocksmithStrategyInterface
		BlockIncompleteQueueService BlockIncompleteQueueServiceInterface
		BlockPoolService            BlockPoolServiceInterface
		Observer                    *observer.Observer
		Logger                      *log.Logger
		TransactionUtil             transaction.UtilInterface
		ReceiptUtil                 coreUtil.ReceiptUtilInterface
		PublishedReceiptUtil        coreUtil.PublishedReceiptUtilInterface
		TransactionCoreService      TransactionCoreServiceInterface
		CoinbaseService             CoinbaseServiceInterface
		ParticipationScoreService   ParticipationScoreServiceInterface
		PublishedReceiptService     PublishedReceiptServiceInterface
		BlockStateStorage           storage.CacheStorageInterface
		BlocksStorage               storage.CacheStackStorageInterface
		ScrambleNodeService         ScrambleNodeServiceInterface
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
			name: "Fail-GetLastBlock",
			fields: fields{
				Chaintype:               &chaintype.MainChain{},
				QueryExecutor:           &mockExecutorBlockPopGetLastBlockFail{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.MainChain{}),
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
				BlockPoolService:        &mockBlockPoolServicePopOffToBlockSuccess{},
				Observer:                nil,
				Logger:                  log.New(),
				BlockStateStorage:       &mockPopOffToBlockBlockStateStorageFail{},
				ScrambleNodeService:     &mockScrambleServicePopOffToBlockSuccess{},
			},
			args: args{
				commonBlock: mockGoodCommonBlock,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Fail-HardFork",
			fields: fields{
				Chaintype:               &chaintype.MainChain{},
				QueryExecutor:           &mockExecutorBlockPopSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.MainChain{}),
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
				BlockPoolService:        &mockBlockPoolServicePopOffToBlockSuccess{},
				TransactionCoreService:  &mockPopOffToBlockTransactionCoreService{},
				Logger:                  log.New(),
				BlockStateStorage:       &mockPopOffToBlockBlockStateStorageHardForkSuccess{},
				ScrambleNodeService:     &mockScrambleServicePopOffToBlockSuccess{},
			},
			args: args{
				commonBlock: mockBadCommonBlockHardFork,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Fail-CommonBlockNotFound",
			fields: fields{
				Chaintype:               &chaintype.MainChain{},
				QueryExecutor:           &mockExecutorBlockPopFailCommonNotFound{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.MainChain{}),
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
				BlockPoolService:        &mockBlockPoolServicePopOffToBlockSuccess{},
				TransactionCoreService:  &mockPopOffToBlockTransactionCoreService{},
				Logger:                  log.New(),
				BlockStateStorage:       &mockPopOffToBlockBlockStateStorageSuccess{},
				ScrambleNodeService:     &mockScrambleServicePopOffToBlockSuccess{},
			},
			args: args{
				commonBlock: mockGoodCommonBlock,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Fail-GetPublishedReceiptError",
			fields: fields{
				Chaintype:               &chaintype.MainChain{},
				QueryExecutor:           &mockExecutorBlockPopSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.MainChain{}),
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
				BlockPoolService:        &mockBlockPoolServicePopOffToBlockSuccess{},
				TransactionCoreService:  &mockPopOffToBlockTransactionCoreService{},
				Logger:                  log.New(),
				PublishedReceiptUtil:    &mockPublishedReceiptUtilSuccess{},
				BlockStateStorage:       &mockPopOffToBlockBlockStateStorageSuccess{},
				BlocksStorage:           &mockPopOffToBlockBlocksStorageSuccess{},
				ScrambleNodeService:     &mockScrambleServicePopOffToBlockSuccess{},
			},
			args: args{
				commonBlock: mockGoodCommonBlock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Fail-GetMempoolToBackupFail",
			fields: fields{
				Chaintype:               &chaintype.MainChain{},
				QueryExecutor:           &mockExecutorBlockPopSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.MainChain{}),
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
				BlockPoolService:        &mockBlockPoolServicePopOffToBlockSuccess{},
				TransactionCoreService:  &mockPopOffToBlockTransactionCoreService{},
				Logger:                  log.New(),
				PublishedReceiptUtil:    &mockPublishedReceiptUtilSuccess{},
				BlockStateStorage:       &mockPopOffToBlockBlockStateStorageSuccess{},
				BlocksStorage:           &mockPopOffToBlockBlocksStorageSuccess{},
				ScrambleNodeService:     &mockScrambleServicePopOffToBlockSuccess{},
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
				Chaintype:               &chaintype.MainChain{},
				QueryExecutor:           &mockExecutorBlockPopSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.MainChain{}),
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
				BlockPoolService:        &mockBlockPoolServicePopOffToBlockSuccess{},
				TransactionCoreService:  &mockPopOffToBlockTransactionCoreService{},
				Logger:                  log.New(),
				PublishedReceiptUtil:    &mockPublishedReceiptUtilSuccess{},
				BlockStateStorage:       &mockPopOffToBlockBlockStateStorageSuccess{},
				BlocksStorage:           &mockPopOffToBlockBlocksStorageSuccess{},
				ScrambleNodeService:     &mockScrambleServicePopOffToBlockSuccess{},
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
				Chaintype:                   tt.fields.Chaintype,
				QueryExecutor:               tt.fields.QueryExecutor,
				BlockQuery:                  tt.fields.BlockQuery,
				MempoolQuery:                tt.fields.MempoolQuery,
				TransactionQuery:            tt.fields.TransactionQuery,
				PublishedReceiptQuery:       tt.fields.PublishedReceiptQuery,
				SkippedBlocksmithQuery:      tt.fields.SkippedBlocksmithQuery,
				Signature:                   tt.fields.Signature,
				MempoolService:              tt.fields.MempoolService,
				ReceiptService:              tt.fields.ReceiptService,
				NodeRegistrationService:     tt.fields.NodeRegistrationService,
				BlocksmithService:           tt.fields.BlocksmithService,
				ActionTypeSwitcher:          tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:         tt.fields.AccountBalanceQuery,
				ParticipationScoreQuery:     tt.fields.ParticipationScoreQuery,
				NodeRegistrationQuery:       tt.fields.NodeRegistrationQuery,
				AccountLedgerQuery:          tt.fields.AccountLedgerQuery,
				BlocksmithStrategy:          tt.fields.BlocksmithStrategy,
				BlockIncompleteQueueService: tt.fields.BlockIncompleteQueueService,
				BlockPoolService:            tt.fields.BlockPoolService,
				Observer:                    tt.fields.Observer,
				Logger:                      tt.fields.Logger,
				TransactionUtil:             tt.fields.TransactionUtil,
				ReceiptUtil:                 tt.fields.ReceiptUtil,
				PublishedReceiptUtil:        tt.fields.PublishedReceiptUtil,
				TransactionCoreService:      tt.fields.TransactionCoreService,
				CoinbaseService:             tt.fields.CoinbaseService,
				ParticipationScoreService:   tt.fields.ParticipationScoreService,
				PublishedReceiptService:     tt.fields.PublishedReceiptService,
				BlockStateStorage:           tt.fields.BlockStateStorage,
				BlocksStorage:               tt.fields.BlocksStorage,
				ScrambleNodeService:         tt.fields.ScrambleNodeService,
			}
			got, err := bs.PopOffToBlock(tt.args.commonBlock)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.PopOffToBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.PopOffToBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockMempoolServiceProcessQueuedBlockSuccess struct {
		MempoolService
	}
	mockBlocksmithServiceProcessQueued struct {
		strategy.BlocksmithStrategyMain
	}

	mockBlockIncompleteQueueServiceAlreadyExist struct {
		BlockIncompleteQueueService
	}
)

func (*mockMempoolServiceProcessQueuedBlockSuccess) GetMempoolTransactions() (storage.MempoolMap, error) {
	return make(storage.MempoolMap), nil
}

func (*mockBlocksmithServiceProcessQueued) GetSortedBlocksmiths(*model.Block) []*model.Blocksmith {
	return mockBlocksmiths
}
func (*mockBlocksmithServiceProcessQueued) GetSortedBlocksmithsMap(block *model.Block) map[string]*int64 {
	var result = make(map[string]*int64)
	for index, mock := range mockBlocksmiths {
		mockIndex := int64(index)
		result[string(mock.NodePublicKey)] = &mockIndex
	}
	return result
}
func (*mockBlocksmithServiceProcessQueued) SortBlocksmiths(block *model.Block, withLock bool) {
}

func (*mockBlocksmithServiceProcessQueued) IsBlockTimestampValid(blocksmithIndex, numberOfBlocksmiths int64, previousBlock,
	currentBlock *model.Block) error {
	return nil
}

func (*mockBlockIncompleteQueueServiceAlreadyExist) GetBlockQueue(blockID int64) *model.Block {
	return &model.Block{
		ID: constant.MainchainGenesisBlockID,
	}
}

type (
	mockTransactionCoreServiceProcessQueueBlockDuplicateFound struct {
		TransactionCoreService
	}
	mockTransactionCoreServiceProcessQueueBlockGetTransactionsError struct {
		TransactionCoreService
	}
	mockTransactionCoreServiceProcessQueueBlockNoDuplicateFound struct {
		TransactionCoreService
	}
)

func (*mockTransactionCoreServiceProcessQueueBlockDuplicateFound) GetTransactionsByIds(
	transactionIds []int64,
) ([]*model.Transaction, error) {
	return make([]*model.Transaction, 2), nil
}

func (*mockTransactionCoreServiceProcessQueueBlockGetTransactionsError) GetTransactionsByIds(
	transactionIds []int64,
) ([]*model.Transaction, error) {
	return nil, errors.New("mockedError")
}

func (*mockTransactionCoreServiceProcessQueueBlockNoDuplicateFound) GetTransactionsByIds(
	transactionIds []int64,
) ([]*model.Transaction, error) {
	return make([]*model.Transaction, 0), nil
}

func TestBlockService_ProcessQueueBlock(t *testing.T) {
	mockBlockData.Transactions = []*model.Transaction{
		mockTransaction,
	}
	var (
		previousBlockHash, _        = util.GetBlockHash(&mockBlockData, &chaintype.MainChain{})
		mockBlockWithTransactionIDs = model.Block{
			ID:                   constant.MainchainGenesisBlockID,
			BlockHash:            nil,
			PreviousBlockHash:    previousBlockHash,
			Height:               0,
			Timestamp:            mockBlockData.GetTimestamp() + 1,
			BlockSeed:            nil,
			BlockSignature:       []byte{144, 246, 37, 144, 213, 135},
			CumulativeDifficulty: "1000",
			PayloadLength:        1,
			PayloadHash:          []byte{},
			BlocksmithPublicKey:  mockBlockData.GetBlocksmithPublicKey(),
			TotalAmount:          1,
			TotalFee:             1,
			TotalCoinBase:        1,
			Version:              1,
			TransactionIDs: []int64{
				mockTransaction.GetID(),
			},
			Transactions:      nil,
			PublishedReceipts: nil,
		}
		mockBlockHash, _    = util.GetBlockHash(&mockBlockWithTransactionIDs, &chaintype.MainChain{})
		mockWaitingTxBlocks = make(map[int64]*BlockWithMetaData)
	)
	mockBlockWithTransactionIDs.BlockHash = mockBlockHash

	// add mock block into mock waited tx block
	mockWaitingTxBlocks[mockBlockWithTransactionIDs.GetID()] = &BlockWithMetaData{
		Block: &mockBlockWithTransactionIDs,
	}

	mockPeer := &model.Peer{}

	type fields struct {
		Chaintype                   chaintype.ChainType
		QueryExecutor               query.ExecutorInterface
		BlockQuery                  query.BlockQueryInterface
		MempoolQuery                query.MempoolQueryInterface
		TransactionQuery            query.TransactionQueryInterface
		MerkleTreeQuery             query.MerkleTreeQueryInterface
		PublishedReceiptQuery       query.PublishedReceiptQueryInterface
		SkippedBlocksmithQuery      query.SkippedBlocksmithQueryInterface
		SpinePublicKeyQuery         query.SpinePublicKeyQueryInterface
		Signature                   crypto.SignatureInterface
		MempoolService              MempoolServiceInterface
		ReceiptService              ReceiptServiceInterface
		NodeRegistrationService     NodeRegistrationServiceInterface
		ActionTypeSwitcher          transaction.TypeActionSwitcher
		AccountBalanceQuery         query.AccountBalanceQueryInterface
		ParticipationScoreQuery     query.ParticipationScoreQueryInterface
		NodeRegistrationQuery       query.NodeRegistrationQueryInterface
		AccountLedgerQuery          query.AccountLedgerQueryInterface
		BlocksmithStrategy          strategy.BlocksmithStrategyInterface
		BlockIncompleteQueueService BlockIncompleteQueueServiceInterface
		Observer                    *observer.Observer
		Logger                      *log.Logger
		TransactionCoreService      TransactionCoreServiceInterface
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantIsQueued bool
		wantErr      bool
	}{
		{
			name:   "wantFail:NoTransactions",
			fields: fields{},
			args: args{
				block: &mockBlockData,
			},
			wantIsQueued: false,
			wantErr:      false,
		},
		{
			name: "wantSuccess:BlockAlreadyQueued",
			args: args{
				block: &mockBlockWithTransactionIDs,
			},
			fields: fields{
				BlockIncompleteQueueService: &mockBlockIncompleteQueueServiceAlreadyExist{},
			},
			wantIsQueued: true,
			wantErr:      false,
		},
		{
			name: "wantErr:DuplicateTxFound",
			args: args{
				block: &mockBlockWithTransactionIDs,
			},
			fields: fields{
				Chaintype:        &chaintype.MainChain{},
				QueryExecutor:    &mockQueryExecutorSuccess{},
				BlockQuery:       query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:     query.NewMempoolQuery(&chaintype.MainChain{}),
				MerkleTreeQuery:  query.NewMerkleTreeQuery(),
				TransactionQuery: query.NewTransactionQuery(&chaintype.MainChain{}),
				Signature:        &mockSignature{},
				MempoolService:   &mockMempoolServiceProcessQueuedBlockSuccess{},
				ActionTypeSwitcher: &transaction.TypeSwitcher{
					Executor: &mockQueryExecutorSuccess{},
				},
				AccountBalanceQuery:         query.NewAccountBalanceQuery(),
				AccountLedgerQuery:          query.NewAccountLedgerQuery(),
				SkippedBlocksmithQuery:      query.NewSkippedBlocksmithQuery(),
				NodeRegistrationQuery:       query.NewNodeRegistrationQuery(),
				Observer:                    observer.NewObserver(),
				NodeRegistrationService:     &mockNodeRegistrationServiceSuccess{},
				BlocksmithStrategy:          &mockBlocksmithServiceProcessQueued{},
				BlockIncompleteQueueService: NewBlockIncompleteQueueService(&chaintype.MainChain{}, observer.NewObserver()),
				Logger:                      log.New(),
				TransactionCoreService:      &mockTransactionCoreServiceProcessQueueBlockDuplicateFound{},
			},
			wantIsQueued: false,
			wantErr:      true,
		},
		{
			name: "wantErr:GetTransactionIDsFail",
			args: args{
				block: &mockBlockWithTransactionIDs,
			},
			fields: fields{
				Chaintype:        &chaintype.MainChain{},
				QueryExecutor:    &mockQueryExecutorSuccess{},
				BlockQuery:       query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:     query.NewMempoolQuery(&chaintype.MainChain{}),
				MerkleTreeQuery:  query.NewMerkleTreeQuery(),
				TransactionQuery: query.NewTransactionQuery(&chaintype.MainChain{}),
				Signature:        &mockSignature{},
				MempoolService:   &mockMempoolServiceProcessQueuedBlockSuccess{},
				ActionTypeSwitcher: &transaction.TypeSwitcher{
					Executor: &mockQueryExecutorSuccess{},
				},
				AccountBalanceQuery:         query.NewAccountBalanceQuery(),
				AccountLedgerQuery:          query.NewAccountLedgerQuery(),
				SkippedBlocksmithQuery:      query.NewSkippedBlocksmithQuery(),
				NodeRegistrationQuery:       query.NewNodeRegistrationQuery(),
				Observer:                    observer.NewObserver(),
				NodeRegistrationService:     &mockNodeRegistrationServiceSuccess{},
				BlocksmithStrategy:          &mockBlocksmithServiceProcessQueued{},
				BlockIncompleteQueueService: NewBlockIncompleteQueueService(&chaintype.MainChain{}, observer.NewObserver()),
				Logger:                      log.New(),
				TransactionCoreService:      &mockTransactionCoreServiceProcessQueueBlockGetTransactionsError{},
			},
			wantIsQueued: false,
			wantErr:      true,
		},
		{
			name: "wantSuccess:AllTxInCached",
			args: args{
				block: &mockBlockWithTransactionIDs,
			},
			fields: fields{
				Chaintype:        &chaintype.MainChain{},
				QueryExecutor:    &mockQueryExecutorSuccess{},
				BlockQuery:       query.NewBlockQuery(&chaintype.MainChain{}),
				MempoolQuery:     query.NewMempoolQuery(&chaintype.MainChain{}),
				MerkleTreeQuery:  query.NewMerkleTreeQuery(),
				TransactionQuery: query.NewTransactionQuery(&chaintype.MainChain{}),
				Signature:        &mockSignature{},
				MempoolService:   &mockMempoolServiceProcessQueuedBlockSuccess{},
				ActionTypeSwitcher: &transaction.TypeSwitcher{
					Executor: &mockQueryExecutorSuccess{},
				},
				AccountBalanceQuery:         query.NewAccountBalanceQuery(),
				AccountLedgerQuery:          query.NewAccountLedgerQuery(),
				SkippedBlocksmithQuery:      query.NewSkippedBlocksmithQuery(),
				NodeRegistrationQuery:       query.NewNodeRegistrationQuery(),
				Observer:                    observer.NewObserver(),
				NodeRegistrationService:     &mockNodeRegistrationServiceSuccess{},
				BlocksmithStrategy:          &mockBlocksmithServiceProcessQueued{},
				BlockIncompleteQueueService: NewBlockIncompleteQueueService(&chaintype.MainChain{}, observer.NewObserver()),
				Logger:                      log.New(),
				TransactionCoreService:      &mockTransactionCoreServiceProcessQueueBlockNoDuplicateFound{},
			},
			wantIsQueued: true,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:                   tt.fields.Chaintype,
				QueryExecutor:               tt.fields.QueryExecutor,
				BlockQuery:                  tt.fields.BlockQuery,
				MempoolQuery:                tt.fields.MempoolQuery,
				TransactionQuery:            tt.fields.TransactionQuery,
				PublishedReceiptQuery:       tt.fields.PublishedReceiptQuery,
				SkippedBlocksmithQuery:      tt.fields.SkippedBlocksmithQuery,
				Signature:                   tt.fields.Signature,
				MempoolService:              tt.fields.MempoolService,
				ReceiptService:              tt.fields.ReceiptService,
				NodeRegistrationService:     tt.fields.NodeRegistrationService,
				ActionTypeSwitcher:          tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:         tt.fields.AccountBalanceQuery,
				ParticipationScoreQuery:     tt.fields.ParticipationScoreQuery,
				NodeRegistrationQuery:       tt.fields.NodeRegistrationQuery,
				AccountLedgerQuery:          tt.fields.AccountLedgerQuery,
				BlocksmithStrategy:          tt.fields.BlocksmithStrategy,
				BlockIncompleteQueueService: tt.fields.BlockIncompleteQueueService,
				Observer:                    tt.fields.Observer,
				Logger:                      tt.fields.Logger,
				TransactionUtil:             &transaction.Util{},
				TransactionCoreService:      tt.fields.TransactionCoreService,
			}
			gotIsQueued, err := bs.ProcessQueueBlock(tt.args.block, mockPeer)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.ProcessQueueBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotIsQueued != tt.wantIsQueued {
				t.Errorf("BlockService.ProcessQueueBlock() = %v, want %v", gotIsQueued, tt.wantIsQueued)
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
				mockPublishedReceipt[0].Receipt.SenderPublicKey,
				mockPublishedReceipt[0].Receipt.RecipientPublicKey,
				mockPublishedReceipt[0].Receipt.DatumType,
				mockPublishedReceipt[0].Receipt.DatumHash,
				mockPublishedReceipt[0].Receipt.ReferenceBlockHeight,
				mockPublishedReceipt[0].Receipt.ReferenceBlockHash,
				mockPublishedReceipt[0].Receipt.RMRLinked,
				mockPublishedReceipt[0].Receipt.RecipientSignature,
				mockPublishedReceipt[0].IntermediateHashes,
				mockPublishedReceipt[0].BlockHeight,
				mockPublishedReceipt[0].ReceiptIndex,
				mockPublishedReceipt[0].PublishedIndex,
			))

	}
	rows, _ := db.Query(qStr)
	return rows, nil
}

type (
	// PopulateBlockData mocks
	mockPopulateBlockDataTransactionCoreServiceSuccess struct {
		TransactionCoreService
	}
	mockPopulateBlockDataPublishedReceiptUtilSuccess struct {
		coreUtil.PublishedReceiptUtil
	}

	mockPopulateBlockDataPublishedReceiptUtilFail struct {
		coreUtil.PublishedReceiptUtil
	}
	// PopulateBlockData mocks

)

func (*mockPopulateBlockDataTransactionCoreServiceSuccess) GetTransactionsByBlockID(blockID int64) ([]*model.Transaction, error) {
	return []*model.Transaction{
		mockTransaction,
	}, nil
}

func (*mockPopulateBlockDataPublishedReceiptUtilSuccess) GetPublishedReceiptsByBlockHeight(
	blockHeight uint32,
) ([]*model.PublishedReceipt, error) {
	return []*model.PublishedReceipt{
		mockPublishedReceipt[0],
	}, nil
}

func (*mockPopulateBlockDataPublishedReceiptUtilFail) GetPublishedReceiptsByBlockHeight(
	blockHeight uint32,
) ([]*model.PublishedReceipt, error) {
	return nil, errors.New("mockedError")
}

func TestBlockMainService_PopulateBlockData(t *testing.T) {
	type fields struct {
		Chaintype               chaintype.ChainType
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
		TransactionCoreService  TransactionCoreServiceInterface
		PublishedReceiptUtil    coreUtil.PublishedReceiptUtilInterface
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
				Chaintype:              &chaintype.SpineChain{},
				QueryExecutor:          &mockMainExecutorPopulateBlockDataFail{},
				TransactionQuery:       query.NewTransactionQuery(&chaintype.MainChain{}),
				PublishedReceiptQuery:  query.NewPublishedReceiptQuery(),
				TransactionCoreService: &mockPopulateBlockDataTransactionCoreServiceSuccess{},
				PublishedReceiptUtil:   &mockPopulateBlockDataPublishedReceiptUtilFail{},
				Logger:                 log.New(),
			},
			args: args{
				block: &model.Block{},
			},
			wantErr: true,
		},
		{
			name: "PopulateBlockData:success",
			fields: fields{
				Chaintype:              &chaintype.SpineChain{},
				QueryExecutor:          &mockMainExecutorPopulateBlockDataSuccess{},
				TransactionQuery:       query.NewTransactionQuery(&chaintype.MainChain{}),
				PublishedReceiptQuery:  query.NewPublishedReceiptQuery(),
				TransactionCoreService: &mockPopulateBlockDataTransactionCoreServiceSuccess{},
				PublishedReceiptUtil:   &mockPopulateBlockDataPublishedReceiptUtilSuccess{},
				Logger:                 log.New(),
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
				QueryExecutor:           tt.fields.QueryExecutor,
				BlockQuery:              tt.fields.BlockQuery,
				MempoolQuery:            tt.fields.MempoolQuery,
				TransactionQuery:        tt.fields.TransactionQuery,
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
				TransactionCoreService:  tt.fields.TransactionCoreService,
				Observer:                tt.fields.Observer,
				PublishedReceiptUtil:    tt.fields.PublishedReceiptUtil,
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

type (
	mockReceiptUtil struct {
		coreUtil.ReceiptUtil
		resSignetBytes []byte
	}
)

func (mRu *mockReceiptUtil) GetSignedReceiptBytes(receipt *model.Receipt) []byte {
	if mRu.resSignetBytes != nil {
		return mRu.resSignetBytes
	}
	return []byte{}
}

func TestBlockService_ValidatePayloadHash(t *testing.T) {
	mockBlock := &model.Block{
		PayloadHash: []byte{102, 253, 86, 32, 28, 24, 212, 55, 129, 77, 244, 149, 6, 198, 243, 4, 86, 251, 61, 45, 48, 99, 191,
			108, 13, 232, 254, 123, 170, 190, 3, 141},
		PayloadLength: uint32(13),
		Transactions: []*model.Transaction{
			mockTransaction,
		},
		PublishedReceipts: mockPublishedReceipt,
	}
	mockInvalidBlock := &model.Block{
		PayloadHash: []byte{102, 253, 86, 32, 28, 24, 212, 55, 129, 77, 244, 149, 6, 198, 243, 4, 86, 251, 61, 45, 48, 99, 191,
			108, 13, 232, 254, 123, 170, 190, 3, 0},
		PayloadLength: uint32(13),
		Transactions: []*model.Transaction{
			mockTransaction,
		},
		PublishedReceipts: mockPublishedReceipt,
	}
	type fields struct {
		Chaintype                   chaintype.ChainType
		QueryExecutor               query.ExecutorInterface
		BlockQuery                  query.BlockQueryInterface
		MempoolQuery                query.MempoolQueryInterface
		TransactionQuery            query.TransactionQueryInterface
		PublishedReceiptQuery       query.PublishedReceiptQueryInterface
		SkippedBlocksmithQuery      query.SkippedBlocksmithQueryInterface
		Signature                   crypto.SignatureInterface
		MempoolService              MempoolServiceInterface
		ReceiptService              ReceiptServiceInterface
		NodeRegistrationService     NodeRegistrationServiceInterface
		BlocksmithService           BlocksmithServiceInterface
		ActionTypeSwitcher          transaction.TypeActionSwitcher
		AccountBalanceQuery         query.AccountBalanceQueryInterface
		ParticipationScoreQuery     query.ParticipationScoreQueryInterface
		NodeRegistrationQuery       query.NodeRegistrationQueryInterface
		AccountLedgerQuery          query.AccountLedgerQueryInterface
		BlocksmithStrategy          strategy.BlocksmithStrategyInterface
		BlockIncompleteQueueService BlockIncompleteQueueServiceInterface
		BlockPoolService            BlockPoolServiceInterface
		Observer                    *observer.Observer
		Logger                      *log.Logger
		TransactionUtil             transaction.UtilInterface
		ReceiptUtil                 coreUtil.ReceiptUtilInterface
		PublishedReceiptUtil        coreUtil.PublishedReceiptUtilInterface
		TransactionCoreService      TransactionCoreServiceInterface
		CoinbaseService             CoinbaseServiceInterface
		ParticipationScoreService   ParticipationScoreServiceInterface
		PublishedReceiptService     PublishedReceiptServiceInterface
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ValidatePayloadHash:success",
			fields: fields{
				BlockQuery:         query.NewBlockQuery(&chaintype.MainChain{}),
				ActionTypeSwitcher: &mockTypeActionSuccess{},
				ReceiptUtil: &mockReceiptUtil{
					resSignetBytes: []byte{1, 1, 1, 1, 1},
				},
			},
			args: args{
				block: mockBlock,
			},
		},
		{
			name: "ValidatePayloadHash:fail",
			fields: fields{
				BlockQuery:         query.NewBlockQuery(&chaintype.MainChain{}),
				ActionTypeSwitcher: &mockTypeActionSuccess{},
				ReceiptUtil: &mockReceiptUtil{
					resSignetBytes: []byte{1, 1, 1, 1, 1},
				},
			},
			args: args{
				block: mockInvalidBlock,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Chaintype:                   tt.fields.Chaintype,
				QueryExecutor:               tt.fields.QueryExecutor,
				BlockQuery:                  tt.fields.BlockQuery,
				MempoolQuery:                tt.fields.MempoolQuery,
				TransactionQuery:            tt.fields.TransactionQuery,
				PublishedReceiptQuery:       tt.fields.PublishedReceiptQuery,
				SkippedBlocksmithQuery:      tt.fields.SkippedBlocksmithQuery,
				Signature:                   tt.fields.Signature,
				MempoolService:              tt.fields.MempoolService,
				ReceiptService:              tt.fields.ReceiptService,
				NodeRegistrationService:     tt.fields.NodeRegistrationService,
				BlocksmithService:           tt.fields.BlocksmithService,
				ActionTypeSwitcher:          tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:         tt.fields.AccountBalanceQuery,
				ParticipationScoreQuery:     tt.fields.ParticipationScoreQuery,
				NodeRegistrationQuery:       tt.fields.NodeRegistrationQuery,
				AccountLedgerQuery:          tt.fields.AccountLedgerQuery,
				BlocksmithStrategy:          tt.fields.BlocksmithStrategy,
				BlockIncompleteQueueService: tt.fields.BlockIncompleteQueueService,
				BlockPoolService:            tt.fields.BlockPoolService,
				Observer:                    tt.fields.Observer,
				Logger:                      tt.fields.Logger,
				TransactionUtil:             tt.fields.TransactionUtil,
				ReceiptUtil:                 tt.fields.ReceiptUtil,
				PublishedReceiptUtil:        tt.fields.PublishedReceiptUtil,
				TransactionCoreService:      tt.fields.TransactionCoreService,
				CoinbaseService:             tt.fields.CoinbaseService,
				ParticipationScoreService:   tt.fields.ParticipationScoreService,
				PublishedReceiptService:     tt.fields.PublishedReceiptService,
			}
			if err := bs.ValidatePayloadHash(tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("BlockService.ValidatePayloadHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
