package transaction

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockAuthPoown struct {
		success bool
		auth.ProofOfOwnershipValidation
	}
	// validate mock
	mockExecutorValidateFailExecuteSelectFail struct {
		query.Executor
	}
	mockExecutorValidateFailBalanceNotEnough struct {
		query.Executor
	}
	mockExecutorValidateFailExecuteSelectNodeFail struct {
		query.Executor
	}
	mockExecutorValidateFailExecuteSelectAccountAlreadyOnwer struct {
		query.Executor
	}
	mockExecutorValidateFailExecuteSelectNodeExist struct {
		query.Executor
	}
	mockExecutorValidateFailExecuteSelectNodeExistButDeleted struct {
		query.Executor
	}
	mockExecutorValidateSuccess struct {
		query.Executor
	}

	// undo unconfirmed mock
	mockExecutorUndoUnconfirmedExecuteTransactionsFail struct {
		query.Executor
	}

	mockExecutorUndoUnconfirmedSuccess struct {
		query.Executor
	}

	// apply unconfirmed mock
	mockExecutorApplyUnconfirmedExecuteTransactionFail struct {
		mockExecutorValidateSuccess
	}
	mockExecutorApplyUnconfirmedSuccess struct {
		mockExecutorValidateSuccess
	}

	// apply confirmed mock
	mockApplyConfirmedUndoUnconfirmedFail struct {
		mockExecutorValidateSuccess
	}
	mockApplyConfirmedExecuteTransactionsFail struct {
		mockExecutorValidateSuccess
	}
	mockApplyConfirmedFailNodeAlreadyInRegistry struct {
		mockExecutorValidateSuccess
	}
	mockApplyConfirmedFailNodeAlreadyInRegistrySuccess struct {
		mockExecutorValidateSuccess
	}
	mockApplyConfirmedSuccess struct {
		mockExecutorValidateSuccess
	}
	mockApplyConfirmedSuccessWithExDeleted struct {
		query.Executor
	}
)

func (mk *mockAuthPoown) ValidateProofOfOwnership(
	poown *model.ProofOfOwnership,
	nodePublicKey []byte,
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
) error {
	if mk.success {
		return nil
	}
	return errors.New("MockedError")
}

func (*mockExecutorValidateFailExecuteSelectFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError:selectFail")
}

func (*mockExecutorValidateFailBalanceNotEnough) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
		"AccountID",
		"BlockHeight",
		"SpendableBalance",
		"Balance",
		"PopRevenue",
		"Latest",
	}).AddRow(
		[]byte{1},
		1,
		10,
		10,
		0,
		true,
	),
	)
	return db.Query("")
}

func (*mockExecutorValidateFailExecuteSelectNodeFail) ExecuteSelect(qe string, tx bool,
	args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	if qe == "SELECT account_id,block_height,spendable_balance,balance,pop_revenue,latest FROM account_balance WHERE "+
		"account_id = ? AND latest = 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"AccountID",
			"BlockHeight",
			"SpendableBalance",
			"Balance",
			"PopRevenue",
			"Latest",
		}).AddRow(
			[]byte{1},
			1,
			1000000,
			1000000,
			0,
			true,
		))
		return db.Query("A")
	}
	return nil, errors.New("mockError:nodeFail")
}

func (*mockExecutorValidateFailExecuteSelectNodeExist) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	if qe == "SELECT account_address,block_height,spendable_balance,balance,pop_revenue,latest "+
		"FROM account_balance WHERE account_address = ? AND latest = 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"AccountID",
			"BlockHeight",
			"SpendableBalance",
			"Balance",
			"PopRevenue",
			"Latest",
		}).AddRow(
			[]byte{1},
			1,
			1000000,
			1000000,
			0,
			true,
		))
		return db.Query("A")
	}
	mock.ExpectQuery("B").WillReturnRows(sqlmock.NewRows([]string{
		"NodeID",
		"NodePublicKey",
		"AccountId",
		"RegistrationHeight",
		"NodeAddress",
		"LockedBalance",
		"RegistrationStatus",
		"Latest",
		"Height",
	}).AddRow(
		1,
		[]byte{1},
		[]byte{2},
		1,
		"127.0.0.1",
		1000000,
		uint32(model.NodeRegistrationState_NodeRegistered),
		true,
		1,
	))
	return db.Query("B")
}

func (*mockExecutorValidateFailExecuteSelectAccountAlreadyOnwer) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, "+
		"registration_status, latest, height FROM node_registry WHERE account_address = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"NodeID",
			"NodePublicKey",
			"AccountId",
			"RegistrationHeight",
			"NodeAddress",
			"LockedBalance",
			"RegistrationStatus",
			"Latest",
			"Height",
		}).AddRow(
			1,
			[]byte{1},
			[]byte{2},
			1,
			"127.0.0.1",
			1000000,
			uint32(model.NodeRegistrationState_NodeRegistered),
			true,
			1,
		))
		return db.Query("A")
	}

	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, "+
		"locked_balance, registration_status, latest, height FROM node_registry "+
		"WHERE account_address = ? AND latest=1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"NodeID",
			"NodePublicKey",
			"AccountId",
			"RegistrationHeight",
			"NodeAddress",
			"LockedBalance",
			"RegistrationStatus",
			"Latest",
			"Height",
		}).AddRow(
			1,
			[]byte{1},
			[]byte{2},
			1,
			"127.0.0.1",
			1000000,
			uint32(model.NodeRegistrationState_NodeQueued),
			true,
			1,
		))
		return db.Query("A")
	}

	if qe == "SELECT account_address,block_height,spendable_balance,balance,pop_revenue,latest FROM account_balance WHERE "+
		"account_address = ? AND latest = 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"AccountAddress",
			"BlockHeight",
			"SpendableBalance",
			"Balance",
			"PopRevenue",
			"Latest",
		}).AddRow(
			"BCZ",
			1,
			1000000,
			1000000,
			0,
			true,
		))
		return db.Query("A")
	}
	if qe == "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty,"+
		" payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version"+
		" FROM main_block ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"previous_block_hash",
			"height",
			"timestamp",
			"block_seed",
			"block_signature",
			"cumulative_difficulty",
			"payload_length",
			"payload_hash",
			"blocksmith_public_key",
			"total_amount",
			"total_fee",
			"total_coinbase",
			"version",
		}).AddRow(
			0,
			[]byte{},
			1,
			1562806389280,
			[]byte{},
			[]byte{},
			100000000,
			0,
			[]byte{},
			senderAddress1,
			100000000,
			10000000,
			1,
			0,
		))
		return db.Query("A")
	}
	if qe == "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty,"+
		" payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version"+
		" FROM main_block WHERE height = 0" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"previous_block_hash",
			"height",
			"timestamp",
			"block_seed",
			"block_signature",
			"cumulative_difficulty",
			"payload_length",
			"payload_hash",
			"blocksmith_public_key",
			"total_amount",
			"total_fee",
			"total_coinbase",
			"version",
		}).AddRow(
			0,
			[]byte{},
			1,
			1562806389280,
			[]byte{},
			[]byte{},
			100000000,
			0,
			[]byte{},
			senderAddress1,
			100000000,
			10000000,
			1,
			0,
		))
		return db.Query("A")
	}
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance,"+
		" registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}))
		return db.Query("A")
	}
	return nil, nil
}

func (*mockExecutorValidateFailExecuteSelectNodeExistButDeleted) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	if qe == "SELECT account_address,block_height,spendable_balance,balance,pop_revenue,latest "+
		"FROM account_balance WHERE account_address = ? AND latest = 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"AccountID",
			"BlockHeight",
			"SpendableBalance",
			"Balance",
			"PopRevenue",
			"Latest",
		}).AddRow(
			[]byte{1},
			1,
			1000000,
			1000000,
			0,
			true,
		))
		return db.Query("A")
	}
	mock.ExpectQuery("B").WillReturnRows(sqlmock.NewRows([]string{
		"NodeID",
		"NodePublicKey",
		"AccountId",
		"RegistrationHeight",
		"NodeAddress",
		"LockedBalance",
		"RegistrationStatus",
		"Latest",
		"Height",
	}).AddRow(
		1,
		[]byte{1},
		[]byte{2},
		1,
		"127.0.0.1",
		1000000,
		uint32(model.NodeRegistrationState_NodeDeleted),
		true,
		1,
	))
	return db.Query("B")
}

func (*mockExecutorValidateSuccess) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT account_address,block_height,spendable_balance,balance,pop_revenue,latest FROM account_balance WHERE "+
		"account_address = ? AND latest = 1" {
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"AccountAddress",
			"BlockHeight",
			"SpendableBalance",
			"Balance",
			"PopRevenue",
			"Latest",
		}).AddRow(
			"BCZ",
			1,
			1000000,
			1000000,
			0,
			true,
		))
		return db.QueryRow(qe), nil
	}
	return nil, nil
}

func (*mockExecutorValidateSuccess) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT account_address,block_height,spendable_balance,balance,pop_revenue,latest FROM account_balance WHERE "+
		"account_address = ? AND latest = 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"AccountAddress",
			"BlockHeight",
			"SpendableBalance",
			"Balance",
			"PopRevenue",
			"Latest",
		}).AddRow(
			"BCZ",
			1,
			1000000,
			1000000,
			0,
			true,
		))
		return db.Query("A")
	}
	if qe == "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty,"+
		" payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version"+
		" FROM main_block ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"previous_block_hash",
			"height",
			"timestamp",
			"block_seed",
			"block_signature",
			"cumulative_difficulty",
			"payload_length",
			"payload_hash",
			"blocksmith_public_key",
			"total_amount",
			"total_fee",
			"total_coinbase",
			"version",
		}).AddRow(
			0,
			[]byte{},
			1,
			1562806389280,
			[]byte{},
			[]byte{},
			100000000,
			0,
			[]byte{},
			senderAddress1,
			100000000,
			10000000,
			1,
			0,
		))
		return db.Query("A")
	}
	if qe == "SELECT id, previous_block_hash, height, timestamp, block_seed, block_signature, cumulative_difficulty,"+
		" payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version"+
		" FROM main_block WHERE height = 0" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"previous_block_hash",
			"height",
			"timestamp",
			"block_seed",
			"block_signature",
			"cumulative_difficulty",
			"payload_length",
			"payload_hash",
			"blocksmith_public_key",
			"total_amount",
			"total_fee",
			"total_coinbase",
			"version",
		}).AddRow(
			0,
			[]byte{},
			1,
			1562806389280,
			[]byte{},
			[]byte{},
			100000000,
			0,
			[]byte{},
			senderAddress1,
			100000000,
			10000000,
			1,
			0,
		))
		return db.Query("A")
	}
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance,"+
		" registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}))
		return db.Query("A")
	}
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, "+
		"registration_status, latest, height FROM node_registry WHERE account_address = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}))
		return db.Query("A")
	}
	return nil, nil
}

func (*mockApplyConfirmedFailNodeAlreadyInRegistry) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, "+
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}).AddRow(
			1,
			nodePubKey1,
			"OnEYzI-EMV6UTfoUEzpQUjkSlnqB82-SyRN7469lJTWH",
			100,
			"10.10.10.1",
			10000000000,
			uint32(model.NodeRegistrationState_NodeQueued),
			true,
			110,
		))
		return db.QueryRow("A"), nil
	}

	if qe == "UPDATE node_registry SET latest = 0 WHERE ID = ?" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}))
		return db.QueryRow("A"), nil
	}
	return nil, nil
}

func (*mockApplyConfirmedFailNodeAlreadyInRegistrySuccess) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, "+
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}).AddRow(
			1,
			nodePubKey1,
			"OnEYzI-EMV6UTfoUEzpQUjkSlnqB82-SyRN7469lJTWH",
			100,
			"10.10.10.1",
			10000000000,
			uint32(model.NodeRegistrationState_NodeDeleted),
			true,
			110,
		))
		return db.QueryRow("A"), nil
	}
	return nil, nil
}

func (*mockApplyConfirmedFailNodeAlreadyInRegistrySuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func (*mockExecutorUndoUnconfirmedExecuteTransactionsFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("mockError:undoFail")
}

func (*mockExecutorUndoUnconfirmedExecuteTransactionsFail) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(qe).WillReturnRows(sqlmock.NewRows([]string{
		"id",
		"node_public_key",
		"account_address",
		"registration_height",
		"node_address",
		"locked_balance",
		"registration_status",
		"latest",
		"height",
	}).AddRow(
		int64(10000),
		nodePubKey1,
		senderAddress1,
		uint32(1),
		"10.10.10.10",
		int64(1000),
		model.NodeRegistrationState_NodeRegistered,
		true,
		uint32(1),
	))
	return db.Query("")
}

func (*mockExecutorUndoUnconfirmedExecuteTransactionsFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}).AddRow(
			int64(10000),
			nodePubKey1,
			senderAddress1,
			uint32(1),
			"10.10.10.10",
			int64(1000),
			model.NodeRegistrationState_NodeRegistered,
			true,
			uint32(1),
		))
	return db.QueryRow(qStr), nil
}

func (*mockExecutorUndoUnconfirmedSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*mockExecutorUndoUnconfirmedSuccess) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"id",
		"node_public_key",
		"account_address",
		"registration_height",
		"node_address",
		"locked_balance",
		"registration_status",
		"latest",
		"height",
	}).AddRow(
		int64(10000),
		nodePubKey1,
		senderAddress1,
		uint32(1),
		"10.10.10.10",
		int64(1000),
		model.NodeRegistrationState_NodeRegistered,
		true,
		uint32(1),
	))
	return db.Query(qe)
}

func (*mockExecutorUndoUnconfirmedSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}).AddRow(
			int64(10000),
			nodePubKey1,
			senderAddress1,
			uint32(1),
			"10.10.10.10",
			int64(1000),
			model.NodeRegistrationState_NodeRegistered,
			true,
			uint32(1),
		))
	return db.QueryRow(qStr), nil
}

func (*mockExecutorApplyUnconfirmedExecuteTransactionFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("mockError:ApplyUnconfirmedFail")
}

func (*mockExecutorApplyUnconfirmedSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*mockApplyConfirmedUndoUnconfirmedFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("mockUndoUnconfirmedFail")
}

func (*mockApplyConfirmedExecuteTransactionsFail) ExecuteTransactions(queries [][]interface{}) error {
	return errors.New("mockError:ExecuteTransactionsFail")
}

func (*mockApplyConfirmedExecuteTransactionsFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qStr == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, "+
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}))
		return db.QueryRow("A"), nil
	}
	if qStr == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, "+
		"registration_status, latest, height FROM node_registry WHERE account_address = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}))
		return db.QueryRow("A"), nil
	}
	return nil, nil
}

func (*mockApplyConfirmedSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func (*mockApplyConfirmedSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qStr == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, "+
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}))
		return db.QueryRow("A"), nil
	}
	if qStr == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, "+
		"registration_status, latest, height FROM node_registry WHERE account_address = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}))
		return db.QueryRow("A"), nil
	}
	return nil, nil
}

func (*mockApplyConfirmedSuccessWithExDeleted) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func (*mockApplyConfirmedSuccessWithExDeleted) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, "+
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}))
		return db.QueryRow("A"), nil
	}
	if qe == "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, "+
		"registration_status, latest, height FROM node_registry WHERE account_address = ? AND latest=1 ORDER BY height DESC LIMIT 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}).AddRow(
			1,
			nodePubKey1,
			"OnEYzI-EMV6UTfoUEzpQUjkSlnqB82-SyRN7469lJTWH",
			100,
			"10.10.10.1",
			10000000000,
			uint32(model.NodeRegistrationState_NodeDeleted),
			true,
			110,
		))
		return db.QueryRow("A"), nil
	}

	return nil, nil
}

func (*mockApplyConfirmedSuccessWithExDeleted) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if qe == "UPDATE node_registry SET latest = 0 WHERE ID = ?" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		}))
		return db.Query("A")
	}

	return nil, nil
}

func TestNodeRegistration_ApplyConfirmed(t *testing.T) {
	type fields struct {
		Body                    *model.NodeRegistrationTransactionBody
		Fee                     int64
		SenderAddress           string
		Height                  uint32
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		BlockQuery              query.BlockQueryInterface
		QueryExecutor           query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "ApplyConfirmed:fail-{executeTransactionsFail}",
			wantErr: true,
			fields: fields{
				Height:                  0,
				SenderAddress:           senderAddress1,
				QueryExecutor:           &mockApplyConfirmedExecuteTransactionsFail{},
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Fee:                     1,
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10000,
				},
			},
		},
		{
			name:    "ApplyConfirmed:success",
			wantErr: false,
			fields: fields{
				Height:                  0,
				SenderAddress:           senderAddress1,
				QueryExecutor:           &mockApplyConfirmedSuccess{},
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Fee:                     1,
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10000,
				},
			},
		},
		{
			name:    "ApplyConfirmed:success-{withExDeletedNode}",
			wantErr: false,
			fields: fields{
				Height:                  0,
				SenderAddress:           senderAddress1,
				QueryExecutor:           &mockApplyConfirmedSuccessWithExDeleted{},
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Fee:                     1,
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10000,
				},
			},
		},
		{
			name:    "ApplyConfirmed:fail-{NodeAreadyInRegistry}",
			wantErr: true,
			fields: fields{
				Height:                  0,
				SenderAddress:           senderAddress1,
				QueryExecutor:           &mockApplyConfirmedFailNodeAlreadyInRegistry{},
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Fee:                     1,
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10000,
				},
			},
		},
		{
			name: "ApplyConfirmed:success-{withExDeletedNode_2}",
			fields: fields{
				Height:                  0,
				SenderAddress:           senderAddress1,
				QueryExecutor:           &mockApplyConfirmedFailNodeAlreadyInRegistrySuccess{},
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Fee:                     1,
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10000,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &NodeRegistration{
				Body:                    tt.fields.Body,
				Fee:                     tt.fields.Fee,
				SenderAddress:           tt.fields.SenderAddress,
				Height:                  tt.fields.Height,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				QueryExecutor:           tt.fields.QueryExecutor,
			}
			if err := tx.ApplyConfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistration.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistration_ApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                  *model.NodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "ApplyUnconfirmed:fail-{ExecuteTransactionFail}",
			wantErr: true,
			fields: fields{
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorApplyUnconfirmedExecuteTransactionFail{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10,
					NodePublicKey: []byte{1},
				},
				Fee: 1,
			},
		},
		{
			name:    "ApplyUnconfirmed:success",
			wantErr: false,
			fields: fields{
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorApplyUnconfirmedSuccess{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10,
					NodePublicKey: []byte{1},
				},
				Fee: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &NodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistration.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistration_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                  *model.NodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "UndoApplyUnconfirmed:fail-{executeTransactionsFail}",
			fields: fields{
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorUndoUnconfirmedExecuteTransactionsFail{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				Fee:                   1,
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10000,
				},
			},
			wantErr: true,
		},
		{
			name: "UndoApplyUnconfirmed:success",
			fields: fields{
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorUndoUnconfirmedSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				Fee:                   1,
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10000,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &NodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistration.UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistration_Validate(t *testing.T) {
	_, poown, _, _ := GetFixturesForNoderegistration(query.NewNodeRegistrationQuery())
	bodyWithPoown := &model.NodeRegistrationTransactionBody{
		Poown:         poown,
		NodePublicKey: nodePubKey1,
		NodeAddress: &model.NodeAddress{
			Address: "10.10.10.1",
		},
	}
	bodyWithNullNodeAddress := &model.NodeRegistrationTransactionBody{
		Poown:         poown,
		NodePublicKey: nodePubKey1,
	}
	txBody := &model.NodeRegistrationTransactionBody{
		Poown:         poown,
		NodePublicKey: nodePubKey1,
		NodeAddress: &model.NodeAddress{
			Address: "10.10.10.1",
		},
	}
	bodyWithoutPoown := &model.NodeRegistrationTransactionBody{}
	type fields struct {
		Body                  *model.NodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
		AuthPoown             auth.ProofOfOwnershipValidationInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Validate:success-{GenesisHeight}",
			fields: fields{
				Height:              0,
				Body:                bodyWithoutPoown,
				SenderAddress:       constant.MainchainGenesisAccountAddress,
				QueryExecutor:       &mockExecutorValidateFailExecuteSelectFail{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				BlockQuery:          query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:           &mockAuthPoown{success: false},
			},
			wantErr: false,
		},
		{
			name: "Validate:fail-{PoownRequired}",
			fields: fields{
				Height:              1,
				Body:                bodyWithoutPoown,
				SenderAddress:       senderAddress1,
				QueryExecutor:       &mockExecutorValidateFailExecuteSelectFail{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				BlockQuery:          query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:           &mockAuthPoown{success: false},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{PoownAuth}",
			fields: fields{
				Height:              1,
				Body:                bodyWithPoown,
				SenderAddress:       senderAddress1,
				QueryExecutor:       &mockExecutorValidateFailExecuteSelectFail{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				BlockQuery:          query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:           &mockAuthPoown{success: false},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{executeSelectFail}",
			fields: fields{
				Height:              1,
				Body:                bodyWithPoown,
				SenderAddress:       senderAddress1,
				QueryExecutor:       &mockExecutorValidateFailExecuteSelectFail{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				BlockQuery:          query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:           &mockAuthPoown{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{balanceNotEnough}",
			fields: fields{
				Height: 1,
				Body: &model.NodeRegistrationTransactionBody{
					Poown:         poown,
					NodePublicKey: nodePubKey1,
					LockedBalance: 10000,
				},
				SenderAddress:       senderAddress1,
				QueryExecutor:       &mockExecutorValidateFailBalanceNotEnough{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				BlockQuery:          query.NewBlockQuery(&chaintype.MainChain{}),
				Fee:                 1,
				AuthPoown:           &mockAuthPoown{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{failGetNode}",
			fields: fields{
				Height:                1,
				Body:                  bodyWithPoown,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateFailExecuteSelectNodeFail{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				Fee:                   1,
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{nodeExist}",
			fields: fields{
				Height:                1,
				Body:                  bodyWithPoown,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateFailExecuteSelectNodeExist{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				Fee:                   1,
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{nodeExistButDeleted}",
			fields: fields{
				Height:                1,
				Body:                  bodyWithPoown,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateFailExecuteSelectNodeExistButDeleted{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				Fee:                   1,
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: false,
		},
		{
			name: "Validate:fail-{AccountAlreadyNodeOwner}",
			fields: fields{
				Height:                1,
				Body:                  bodyWithPoown,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateFailExecuteSelectAccountAlreadyOnwer{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				Fee:                   1,
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{NullNodeAddress}",
			fields: fields{
				Height:                1,
				Body:                  bodyWithNullNodeAddress,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				Fee:                   1,
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:success",
			fields: fields{
				Height:                1,
				Body:                  txBody,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &mockExecutorValidateSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				Fee:                   1,
				AuthPoown:             &mockAuthPoown{success: true},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &NodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
				AuthPoown:             tt.fields.AuthPoown,
			}
			if err := tx.Validate(false); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistration.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistration_GetAmount(t *testing.T) {
	type fields struct {
		Body                  *model.NodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		BlockQuery            query.BlockQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "GetAmount:success",
			fields: fields{
				BlockQuery: query.NewBlockQuery(&chaintype.MainChain{}),
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 1000,
				},
			},
			want: 1000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &NodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				BlockQuery:            tt.fields.BlockQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if got := tx.GetAmount(); got != tt.want {
				t.Errorf("NodeRegistration.GetAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistration_GetSize(t *testing.T) {
	type fields struct {
		Body                  *model.NodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		{
			name: "GetSize:success",
			fields: fields{
				Body: &model.NodeRegistrationTransactionBody{
					NodeAddress: &model.NodeAddress{
						Address: "127.0.0.1",
					},
				},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			want: 245,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if got := n.GetSize(); got != tt.want {
				t.Errorf("NodeRegistration.GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistration_ParseBodyBytes(t *testing.T) {

	mockNodeRegistrationQ := query.NewNodeRegistrationQuery()

	_, _, body, bodyBytes := GetFixturesForNoderegistration(mockNodeRegistrationQ)
	// bodyBytes :=
	type fields struct {
		Body                  *model.NodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	type args struct {
		txBodyBytes []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    model.TransactionBodyInterface
		wantErr bool
	}{
		{
			name: "NodeRegistration:error - empty body bytes",
			fields: fields{
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
				AccountBalanceQuery:   nil,
				NodeRegistrationQuery: nil,
				QueryExecutor:         nil,
			},
			args:    args{txBodyBytes: []byte{}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "NodeRegistration:error - wrong public key length",
			fields: fields{
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
				AccountBalanceQuery:   nil,
				NodeRegistrationQuery: nil,
				QueryExecutor:         nil,
			},
			args:    args{txBodyBytes: bodyBytes[:10]},
			want:    nil,
			wantErr: true,
		},
		{
			name: "NodeRegistration:error - no account address length",
			fields: fields{
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
				AccountBalanceQuery:   nil,
				NodeRegistrationQuery: nil,
				QueryExecutor:         nil,
			},
			args:    args{txBodyBytes: bodyBytes[:(len(body.NodePublicKey))]},
			want:    nil,
			wantErr: true,
		},
		{
			name: "NodeRegistration:error - no account address",
			fields: fields{
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
				AccountBalanceQuery:   nil,
				NodeRegistrationQuery: nil,
				QueryExecutor:         nil,
			},
			args:    args{txBodyBytes: bodyBytes[:(len(body.NodePublicKey) + 4)]},
			want:    nil,
			wantErr: true,
		},
		{
			name: "NodeRegistration:error - no node address length",
			fields: fields{
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
				AccountBalanceQuery:   nil,
				NodeRegistrationQuery: nil,
				QueryExecutor:         nil,
			},
			args:    args{txBodyBytes: bodyBytes[:(len(body.NodePublicKey) + 4 + len([]byte(body.AccountAddress)))]},
			want:    nil,
			wantErr: true,
		},
		{
			name: "NodeRegistration:error - no node address",
			fields: fields{
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
				AccountBalanceQuery:   nil,
				NodeRegistrationQuery: nil,
				QueryExecutor:         nil,
			},
			args:    args{txBodyBytes: bodyBytes[:(len(body.NodePublicKey) + 4 + len([]byte(body.AccountAddress)) + 4)]},
			want:    nil,
			wantErr: true,
		},
		{
			name: "NodeRegistration:error - no locked balance",
			fields: fields{
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
				AccountBalanceQuery:   nil,
				NodeRegistrationQuery: nil,
				QueryExecutor:         nil,
			},
			args: args{
				txBodyBytes: bodyBytes[:(len(body.NodePublicKey) + 4 + len([]byte(body.AccountAddress)) + 4 +
					len([]byte(mockNodeRegistrationQ.ExtractNodeAddress(body.GetNodeAddress()))))],
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "NodeRegistration:error - no poown",
			fields: fields{
				Body:                  nil,
				Fee:                   0,
				SenderAddress:         "",
				Height:                0,
				AccountBalanceQuery:   nil,
				NodeRegistrationQuery: nil,
				QueryExecutor:         nil,
			},
			args: args{
				txBodyBytes: bodyBytes[:(len(body.NodePublicKey) + 4 + len([]byte(body.AccountAddress)) + 4 +
					len([]byte(mockNodeRegistrationQ.ExtractNodeAddress(body.GetNodeAddress()))) + int(constant.Balance))],
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "NodeRegistration:ParseBodyBytes - success",
			fields: fields{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			args: args{
				txBodyBytes: bodyBytes,
			},
			want:    body,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			got, err := n.ParseBodyBytes(tt.args.txBodyBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBodyBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistration.ParseBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistration_GetBodyBytes(t *testing.T) {
	_, _, body, bodyBytes := GetFixturesForNoderegistration(query.NewNodeRegistrationQuery())
	type fields struct {
		Body                  *model.NodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	type args struct {
		txBody *model.NodeRegistrationTransactionBody
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		{
			name: "GetBodyBytes:success",
			fields: fields{
				Body:                  body,
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			args: args{
				txBody: body,
			},
			want: bodyBytes,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if got := n.GetBodyBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistration.GetBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistration_GetTransactionBody(t *testing.T) {
	_, _, mockTxBody, _ := GetFixturesForNoderegistration(query.NewNodeRegistrationQuery())
	type fields struct {
		ID                      int64
		Body                    *model.NodeRegistrationTransactionBody
		Fee                     int64
		SenderAddress           string
		Height                  uint32
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		QueryExecutor           query.ExecutorInterface
		AuthPoown               auth.ProofOfOwnershipValidationInterface
	}
	type args struct {
		transaction *model.Transaction
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Body: mockTxBody,
			},
			args: args{
				transaction: &model.Transaction{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &NodeRegistration{
				ID:                      tt.fields.ID,
				Body:                    tt.fields.Body,
				Fee:                     tt.fields.Fee,
				SenderAddress:           tt.fields.SenderAddress,
				Height:                  tt.fields.Height,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				QueryExecutor:           tt.fields.QueryExecutor,
				AuthPoown:               tt.fields.AuthPoown,
			}
			tx.GetTransactionBody(tt.args.transaction)
		})
	}
}

func TestNodeRegistration_SkipMempoolTransaction(t *testing.T) {
	type fields struct {
		ID                      int64
		Body                    *model.NodeRegistrationTransactionBody
		Fee                     int64
		SenderAddress           string
		Height                  uint32
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		QueryExecutor           query.ExecutorInterface
		AuthPoown               auth.ProofOfOwnershipValidationInterface
	}
	type args struct {
		selectedTransactions []*model.Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "SkipMempoolTransaction:success-{Filtered}",
			fields: fields{
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			},
			args: args{
				selectedTransactions: []*model.Transaction{
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
						TransactionType:      uint32(model.TransactionType_NodeRegistrationTransaction),
					},
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
						TransactionType:      uint32(model.TransactionType_EmptyTransaction),
					},
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
						TransactionType:      uint32(model.TransactionType_ClaimNodeRegistrationTransaction),
					},
				},
			},
			want: true,
		},
		{
			name: "SkipMempoolTransaction:success-{UnFiltered_DifferentSenders}",
			fields: fields{
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			},
			args: args{
				selectedTransactions: []*model.Transaction{
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tAAAA",
						TransactionType:      uint32(model.TransactionType_NodeRegistrationTransaction),
					},
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
						TransactionType:      uint32(model.TransactionType_EmptyTransaction),
					},
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tAAAA",
						TransactionType:      uint32(model.TransactionType_ClaimNodeRegistrationTransaction),
					},
				},
			},
		},
		{
			name: "SkipMempoolTransaction:success-{UnFiltered_NoOtherRecordsFound}",
			fields: fields{
				SenderAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			},
			args: args{
				selectedTransactions: []*model.Transaction{
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
						TransactionType:      uint32(model.TransactionType_SetupAccountDatasetTransaction),
					},
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
						TransactionType:      uint32(model.TransactionType_EmptyTransaction),
					},
					{
						SenderAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
						TransactionType:      uint32(model.TransactionType_SendMoneyTransaction),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &NodeRegistration{
				ID:                      tt.fields.ID,
				Body:                    tt.fields.Body,
				Fee:                     tt.fields.Fee,
				SenderAddress:           tt.fields.SenderAddress,
				Height:                  tt.fields.Height,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				QueryExecutor:           tt.fields.QueryExecutor,
				AuthPoown:               tt.fields.AuthPoown,
			}
			got, err := tx.SkipMempoolTransaction(tt.args.selectedTransactions)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistration.SkipMempoolTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NodeRegistration.SkipMempoolTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}
