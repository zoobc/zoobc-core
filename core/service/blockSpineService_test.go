// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
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
		TotalAmount:         0,
		TotalFee:            0,
		TotalCoinBase:       0,
		Version:             0,
		PayloadLength:       1,
		PayloadHash:         []byte{},
		SpineBlockManifests: make([]*model.SpineBlockManifest, 0),
	}
	mockSpineBlockData1 = model.Block{
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
		TotalAmount:         0,
		TotalFee:            0,
		TotalCoinBase:       0,
		Version:             0,
		PayloadLength:       1,
		PayloadHash:         []byte{},
		SpineBlockManifests: make([]*model.SpineBlockManifest, 0),
		SpinePublicKeys:     []*model.SpinePublicKey{mockSpinePublicKey},
	}
	mockSpinePublicKey = &model.SpinePublicKey{
		NodeID:          1,
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

	mockSpineNodeRegistrationServiceSuccess struct {
		NodeRegistrationService
	}

	mockSpineNodeRegistrationServiceFail struct {
		NodeRegistrationService
	}

	mockBlockSpinePublicKeyService struct {
		BlockSpinePublicKeyService
	}
)

func (*mockBlockSpinePublicKeyService) GetSpinePublicKeysByBlockHeight(height uint32) (spinePublicKeys []*model.SpinePublicKey, err error) {
	return nil, nil
}

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
			AccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
				45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		},
	}, nil
}

func (*mockSpineNodeRegistrationServiceSuccess) AdmitNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	return nil
}

func (*mockSpineNodeRegistrationServiceSuccess) SelectNodesToBeExpelled() ([]*model.NodeRegistration, error) {
	return []*model.NodeRegistration{
		{
			AccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
				45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
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
			AccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
				45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
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
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
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
func (*mockSpineQueryExecutorFail) BeginTx(bool, int) error { return nil }

func (*mockSpineQueryExecutorFail) RollbackTx(bool) error { return nil }

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
func (*mockSpineQueryExecutorFail) CommitTx(bool) error {
	return errors.New("mockSpineError:commitFail")
}

// mockSpineQueryExecutorSuccess
func (*mockSpineQueryExecutorSuccess) BeginTx(bool, int) error { return nil }

func (*mockSpineQueryExecutorSuccess) RollbackTx(bool) error { return nil }

func (*mockSpineQueryExecutorSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}
func (*mockSpineQueryExecutorSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}
func (*mockSpineQueryExecutorSuccess) CommitTx(bool) error { return nil }

func (*mockSpineQueryExecutorSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()

	switch qStr {
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND latest=1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "NodePublicKey", "AccountAddress", "RegistrationHeight", "LockedBalance", "RegistrationStatus",
			"Latest", "Height",
		}).AddRow(1, bcsNodePubKey1, bcsAddress1, 10, 100000000, uint32(model.NodeRegistrationState_NodeQueued), true, 100))
	case "SELECT id, block_height, tree, timestamp FROM merkle_tree ORDER BY timestamp DESC LIMIT 1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "BlockHeight", "Tree", "Timestamp",
		}))
	case "SELECT MAX(height), id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version, " +
		"merkle_root, merkle_tree, reference_block_height FROM spine_block":
		mockSpineRows := mockSpine.NewRows(query.NewBlockQuery(&chaintype.SpineChain{}).Fields)
		mockSpineRows.AddRow(
			mockSpineBlockData.GetHeight(),
			mockSpineBlockData.GetID(),
			mockSpineBlockData.GetBlockHash(),
			mockSpineBlockData.GetPreviousBlockHash(),
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
			mockSpineBlockData.GetMerkleRoot(),
			mockSpineBlockData.GetMerkleTree(),
			mockSpineBlockData.GetReferenceBlockHeight(),
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
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE id = ? AND latest=1":
		for idx, arg := range args {
			if idx == 0 {
				nodeID := fmt.Sprintf("%d", arg)
				switch nodeID {
				case "1":
					mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "locked_balance", "registration_status", "latest", "height",
					}).AddRow(1, bcsNodePubKey1, bcsAddress1, 10, 100000000, uint32(model.NodeRegistrationState_NodeRegistered), true, 100))
				case "2":
					mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "locked_balance", "registration_status", "latest", "height",
					}).AddRow(2, bcsNodePubKey2, bcsAddress2, 20, 100000000, uint32(model.NodeRegistrationState_NodeRegistered), true, 200))
				case "3":
					mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "locked_balance", "registration_status", "latest", "height",
					}).AddRow(3, bcsNodePubKey3, bcsAddress3, 30, 100000000, uint32(model.NodeRegistrationState_NodeRegistered), true, 300))
				case "4":
					mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
						"account_address", "registration_height", "locked_balance", "registration_status", "latest", "height",
					}).AddRow(3, mockSpineGoodBlock.BlocksmithPublicKey, bcsAddress3, 30, 100000000,
						uint32(model.NodeRegistrationState_NodeRegistered), true, 300))
				}
			}
		}
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND height <= ? " +
		"ORDER BY height DESC LIMIT 1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{"id", "node_public_key",
			"account_address", "registration_height", "locked_balance", "registration_status", "latest", "height",
		}).AddRow(1, bcsNodePubKey1, bcsAddress1, 10, 100000000, uint32(model.NodeRegistrationState_NodeQueued), true, 100))
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version, merkle_root, " +
		"merkle_tree, reference_block_height FROM spine_block WHERE height = 0":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.SpineChain{}).Fields,
		).AddRow(1, 1, []byte{}, []byte{}, 10000, []byte{}, []byte{}, "", 2, []byte{}, bcsNodePubKey1, 0, 0, 0, 1, []byte{}, []byte{}, 0))
	case "SELECT A.node_id, A.score, A.latest, A.height FROM participation_score as A INNER JOIN node_registry as B " +
		"ON A.node_id = B.id WHERE B.node_public_key=? AND B.latest=1 AND B.registration_status=0 AND A.latest=1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"node_id",
			"score",
			"latest",
			"height",
		},
		).AddRow(-1, 100000, true, 0))
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version, merkle_root, " +
		"merkle_tree, reference_block_height FROM spine_block ORDER BY " +
		"height DESC LIMIT 1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).
			WillReturnRows(sqlmock.NewRows(
				query.NewBlockQuery(&chaintype.SpineChain{}).Fields,
			).AddRow(
				mockSpineBlockData.GetHeight(),
				mockSpineBlockData.GetID(),
				mockSpineBlockData.GetBlockHash(),
				mockSpineBlockData.GetPreviousBlockHash(),
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
				mockSpineBlockData.GetMerkleRoot(),
				mockSpineBlockData.GetMerkleTree(),
				mockSpineBlockData.GetReferenceBlockHeight(),
			))
	case "SELECT node_public_key, node_id, public_key_action, main_block_height, latest, height FROM spine_public_key WHERE height = 1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).
			WillReturnRows(sqlmock.NewRows(
				query.NewSpinePublicKeyQuery().Fields,
			).AddRow(
				mockSpinePublicKey.NodePublicKey,
				mockSpinePublicKey.NodeID,
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
			mockSpinePublishedReceipt[0].Receipt.SenderPublicKey,
			mockSpinePublishedReceipt[0].Receipt.RecipientPublicKey,
			mockSpinePublishedReceipt[0].Receipt.DatumType,
			mockSpinePublishedReceipt[0].Receipt.DatumHash,
			mockSpinePublishedReceipt[0].Receipt.ReferenceBlockHeight,
			mockSpinePublishedReceipt[0].Receipt.ReferenceBlockHash,
			mockSpinePublishedReceipt[0].Receipt.RMRLinked,
			mockSpinePublishedReceipt[0].Receipt.RecipientSignature,
			mockSpinePublishedReceipt[0].IntermediateHashes,
			mockSpinePublishedReceipt[0].BlockHeight,
			mockSpinePublishedReceipt[0].ReceiptIndex,
			mockSpinePublishedReceipt[0].PublishedIndex,
		))
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
		"registration_status, latest, height, max(height) AS max_height FROM node_registry where height <= 0 AND " +
		"registration_status = 0 GROUP BY id ORDER BY height DESC":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id", "node_public_key", "account_address", "registration_height", "locked_balance",
			"registration_status", "latest", "height",
		}))
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
		"registration_status, latest, height, max(height) AS max_height FROM node_registry where height <= 1 " +
		"AND registration_status = 0 GROUP BY id ORDER BY height DESC":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id", "node_public_key", "account_address", "registration_height", "locked_balance",
			"registration_status", "latest", "height",
		}))
	}
	rows, _ := db.Query(qe)
	return rows, nil
}

var mockSpinePublishedReceipt = []*model.PublishedReceipt{
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
			PayloadHash: []byte{167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160, 97, 214, 98, 245, 128,
				255, 77, 228, 59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74},
			PayloadLength:  0,
			BlockSignature: []byte{},
		}
		mockSpineBlockHash, _ = util.GetBlockHash(mockSpineBlock, &chaintype.SpineChain{})
	)
	mockSpineBlock.BlockHash = mockSpineBlockHash

	type fields struct {
		Chaintype                 chaintype.ChainType
		QueryExecutor             query.ExecutorInterface
		BlockQuery                query.BlockQueryInterface
		Signature                 crypto.SignatureInterface
		BlocksmithStrategy        strategy.BlocksmithStrategyInterface
		Observer                  *observer.Observer
		Logger                    *log.Logger
		SpinePublicKeyService     BlockSpinePublicKeyServiceInterface
		SpineBlockManifestService SpineBlockManifestServiceInterface
		BlocksmithService         BlocksmithServiceInterface
		SnapshotMainBlockService  SnapshotBlockServiceInterface
		BlockStateStorage         storage.CacheStorageInterface
		BlockchainStatusService   BlockchainStatusServiceInterface
		MainBlockService          BlockServiceInterface
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
		secretPhrase         string
		spinePublicKeys      []*model.SpinePublicKey
		spineBlockManifests  []*model.SpineBlockManifest
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
				secretPhrase:        "secretphrase",
			},
			want: mockSpineBlock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:                 tt.fields.Chaintype,
				QueryExecutor:             tt.fields.QueryExecutor,
				BlockQuery:                tt.fields.BlockQuery,
				Signature:                 tt.fields.Signature,
				BlocksmithStrategy:        tt.fields.BlocksmithStrategy,
				Observer:                  tt.fields.Observer,
				Logger:                    tt.fields.Logger,
				SpinePublicKeyService:     tt.fields.SpinePublicKeyService,
				SpineBlockManifestService: tt.fields.SpineBlockManifestService,
				BlocksmithService:         tt.fields.BlocksmithService,
				SnapshotMainBlockService:  tt.fields.SnapshotMainBlockService,
				BlockStateStorage:         tt.fields.BlockStateStorage,
				BlockchainStatusService:   tt.fields.BlockchainStatusService,
				MainBlockService:          tt.fields.MainBlockService,
			}
			got, err := bs.NewSpineBlock(
				tt.args.version,
				tt.args.previousBlockHash,
				tt.args.blockSeed,
				tt.args.blockSmithPublicKey,
				tt.args.merkleRoot,
				tt.args.merkleTree,
				tt.args.previousBlockHeight,
				tt.args.referenceBlockHeight,
				tt.args.timestamp,
				tt.args.secretPhrase,
				tt.args.spinePublicKeys,
				tt.args.spineBlockManifests,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineService.NewSpineBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockSpineService.NewSpineBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockSpineService_NewGenesisBlock(t *testing.T) {
	type fields struct {
		Chaintype                 chaintype.ChainType
		QueryExecutor             query.ExecutorInterface
		BlockQuery                query.BlockQueryInterface
		Signature                 crypto.SignatureInterface
		BlocksmithStrategy        strategy.BlocksmithStrategyInterface
		Observer                  *observer.Observer
		Logger                    *log.Logger
		SpinePublicKeyService     BlockSpinePublicKeyServiceInterface
		SpineBlockManifestService SpineBlockManifestServiceInterface
		BlocksmithService         BlocksmithServiceInterface
		SnapshotMainBlockService  SnapshotBlockServiceInterface
		BlockStateStorage         storage.CacheStorageInterface
		BlockchainStatusService   BlockchainStatusServiceInterface
		MainBlockService          BlockServiceInterface
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:                 tt.fields.Chaintype,
				QueryExecutor:             tt.fields.QueryExecutor,
				BlockQuery:                tt.fields.BlockQuery,
				Signature:                 tt.fields.Signature,
				BlocksmithStrategy:        tt.fields.BlocksmithStrategy,
				Observer:                  tt.fields.Observer,
				Logger:                    tt.fields.Logger,
				SpinePublicKeyService:     tt.fields.SpinePublicKeyService,
				SpineBlockManifestService: tt.fields.SpineBlockManifestService,
				BlocksmithService:         tt.fields.BlocksmithService,
				SnapshotMainBlockService:  tt.fields.SnapshotMainBlockService,
				BlockStateStorage:         tt.fields.BlockStateStorage,
				BlockchainStatusService:   tt.fields.BlockchainStatusService,
				MainBlockService:          tt.fields.MainBlockService,
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
				t.Errorf("BlockSpineService.NewGenesisBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
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
			Score:         new(big.Int).SetInt64(1000),
		},
		{
			NodePublicKey: bcsNodePubKey2,
			NodeID:        3,
			Score:         new(big.Int).SetInt64(2000),
		},
		{
			NodePublicKey: mockSpineBlockData.BlocksmithPublicKey,
			NodeID:        4,
			Score:         new(big.Int).SetInt64(3000),
		},
	}
)

type (
	mockSpineBlocksmithServicePushBlock struct {
		strategy.BlocksmithStrategyMain
	}
)

func (*mockSpineBlocksmithServicePushBlock) IsBlockTimestampValid(blocksmithIndex, numberOfBlocksmiths int64, previousBlock,
	currentBlock *model.Block) error {
	return nil
}

type (
	mockSpineBlockStateStorageSuccess struct {
		storage.CacheStorageInterface
	}
	mockSpineBlockStateStorageFail struct {
		storage.CacheStorageInterface
	}
)

func (*mockSpineBlockStateStorageSuccess) GetItem(lastChange, item interface{}) error {
	var blockCopy, _ = item.(*model.Block)
	*blockCopy = mockSpineBlockData
	return nil
}

func (*mockSpineBlockStateStorageFail) GetItem(lastChange, item interface{}) error {
	return errors.New("MockedError")
}
func TestBlockSpineService_GetLastBlock(t *testing.T) {
	type fields struct {
		Chaintype                 chaintype.ChainType
		QueryExecutor             query.ExecutorInterface
		BlockQuery                query.BlockQueryInterface
		MempoolQuery              query.MempoolQueryInterface
		TransactionQuery          query.TransactionQueryInterface
		SpinePublicKeyQuery       query.SpinePublicKeyQueryInterface
		Signature                 crypto.SignatureInterface
		ActionTypeSwitcher        transaction.TypeActionSwitcher
		SpinePublicKeyService     BlockSpinePublicKeyServiceInterface
		SpineBlockManifestService SpineBlockManifestServiceInterface
		BlockStateStorage         storage.CacheStorageInterface
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
				SpinePublicKeyService: &BlockSpinePublicKeyService{
					Logger:                log.New(),
					NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
					QueryExecutor:         &mockSpineQueryExecutorSuccess{},
					Signature:             nil,
					SpinePublicKeyQuery:   query.NewSpinePublicKeyQuery(),
				},
				SpineBlockManifestService: &mockSpineBlockManifestService{},
				BlockStateStorage:         &mockSpineBlockStateStorageSuccess{},
			},
			want:    &mockSpineBlockData1,
			wantErr: false,
		},
		{
			name: "GetLastBlock:SelectFail",
			fields: fields{
				Chaintype:                 &chaintype.SpineChain{},
				QueryExecutor:             &mockSpineQueryExecutorFail{},
				BlockQuery:                query.NewBlockQuery(&chaintype.SpineChain{}),
				SpinePublicKeyQuery:       query.NewSpinePublicKeyQuery(),
				SpineBlockManifestService: &mockSpineBlockManifestService{},
				BlockStateStorage:         &mockSpineBlockStateStorageFail{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:                 tt.fields.Chaintype,
				QueryExecutor:             tt.fields.QueryExecutor,
				BlockQuery:                tt.fields.BlockQuery,
				Signature:                 tt.fields.Signature,
				SpinePublicKeyService:     tt.fields.SpinePublicKeyService,
				SpineBlockManifestService: tt.fields.SpineBlockManifestService,
				BlockStateStorage:         tt.fields.BlockStateStorage,
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
			mockSpineBlockData.GetHeight(),
			mockSpineBlockData.GetID(),
			mockSpineBlockData.GetBlockHash(),
			mockSpineBlockData.GetPreviousBlockHash(),
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
			mockSpineBlockData.GetMerkleRoot(),
			mockSpineBlockData.GetMerkleTree(),
			mockSpineBlockData.GetReferenceBlockHeight(),
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
		mockSpineBlockData.GetHeight(),
		mockSpineBlockData.GetID(),
		mockSpineBlockData.GetBlockHash(),
		mockSpineBlockData.GetPreviousBlockHash(),
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
		mockSpineBlockData.GetMerkleRoot(),
		mockSpineBlockData.GetMerkleTree(),
		mockSpineBlockData.GetReferenceBlockHeight(),
	))
	return db.Query(qStr)
}

func (*mockSpineQueryExecutorGetBlocksFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}

func TestBlockSpineService_GetBlocks(t *testing.T) {
	mockSpineBlockData.SpineBlockManifests = nil
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
	mockSpineBlockManifestService struct {
		SpineBlockManifestService
		ResSpineBlockManifests     []*model.SpineBlockManifest
		ResError                   error
		ResSpineBlockManifestBytes []byte
	}
)

func (ss *mockSpineBlockManifestService) GetSpineBlockManifestsForSpineBlock(
	spineHeight uint32,
	spineTimestamp int64,
) ([]*model.SpineBlockManifest, error) {
	var (
		spineBlockManifests = make([]*model.SpineBlockManifest, 0)
		err                 error
	)
	if ss.ResSpineBlockManifests != nil {
		spineBlockManifests = ss.ResSpineBlockManifests
	}
	if ss.ResError != nil {
		err = ss.ResError
	}
	return spineBlockManifests, err
}

func (ss *mockSpineBlockManifestService) GetSpineBlockManifestBySpineBlockHeight(
	spineHeight uint32,
) ([]*model.SpineBlockManifest, error) {
	var (
		spineBlockManifests = make([]*model.SpineBlockManifest, 0)
		err                 error
	)
	if ss.ResSpineBlockManifests != nil {
		spineBlockManifests = ss.ResSpineBlockManifests
	}
	if ss.ResError != nil {
		err = ss.ResError
	}
	return spineBlockManifests, err
}

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
			[]byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
				45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
			[]byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
				81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
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
				[]byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
					45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
				[]byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
					81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				false,
			).TransactionBytes,
		},
	}, nil
}

// mockSpineMempoolServiceSelectSuccess
func (*mockSpineMempoolServiceSelectSuccess) SelectTransactionsFromMempool(int64, uint32) ([]*model.Transaction, error) {
	txByte := transaction.GetFixturesForSignedMempoolTransaction(
		1,
		1562893305,
		[]byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		[]byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
			81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
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
func (*mockSpineMempoolServiceSelectFail) SelectTransactionsFromMempool(int64, uint32) ([]*model.Transaction, error) {
	return nil, errors.New("want error on select")
}

// mockSpineMempoolServiceSelectSuccess
func (*mockSpineMempoolServiceSelectWrongTransactionBytes) SelectTransactionsFromMempool(int64, uint32) ([]*model.Transaction, error) {
	return []*model.Transaction{
		{
			ID: 1,
		},
	}, nil
}

type (
	mockSpineGenerateblockMainBlockServiceSuccess struct {
		BlockServiceInterface
	}
	mockSpineGenerateBlockSpinePublicKeyServiceSuccess struct {
		BlockSpinePublicKeyServiceInterface
	}
)

var (
	mockGenerateBlockMainBlock = model.Block{
		Height: 1 + constant.SpineReferenceBlockHeightOffset,
	}
)

func (*mockSpineGenerateblockMainBlockServiceSuccess) GetLastBlockCacheFormat() (*storage.BlockCacheObject, error) {
	return &storage.BlockCacheObject{
		ID:        mockGenerateBlockMainBlock.ID,
		Height:    mockGenerateBlockMainBlock.Height,
		BlockHash: mockGenerateBlockMainBlock.BlockHash,
	}, nil
}

func (*mockSpineGenerateblockMainBlockServiceSuccess) GetBlocksFromHeight(startHeight, limit uint32, withAttachedData bool) ([]*model.Block, error) {
	return []*model.Block{
		&mockGenerateBlockMainBlock,
	}, nil
}

func (*mockSpineGenerateBlockSpinePublicKeyServiceSuccess) BuildSpinePublicKeysFromNodeRegistry(
	mainFromHeight,
	mainToHeight,
	spineHeight uint32,
) (spinePublicKeys []*model.SpinePublicKey, err error) {
	return []*model.SpinePublicKey{}, nil
}

func TestBlockSpineService_GenerateBlock(t *testing.T) {
	type fields struct {
		Chaintype                 chaintype.ChainType
		QueryExecutor             query.ExecutorInterface
		BlockQuery                query.BlockQueryInterface
		MempoolQuery              query.MempoolQueryInterface
		TransactionQuery          query.TransactionQueryInterface
		NodeRegistrationQuery     query.NodeRegistrationQueryInterface
		Signature                 crypto.SignatureInterface
		MempoolService            MempoolServiceInterface
		ReceiptService            ReceiptServiceInterface
		BlocksmithStrategy        strategy.BlocksmithStrategyInterface
		ActionTypeSwitcher        transaction.TypeActionSwitcher
		SpinePublicKeyService     BlockSpinePublicKeyServiceInterface
		SpineBlockManifestService SpineBlockManifestServiceInterface
		MainBlockService          BlockServiceInterface
	}
	type args struct {
		previousBlock *model.Block
		secretPhrase  string
		timestamp     int64
		empty         bool
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
				BlocksmithStrategy:    &mockSpineBlocksmithServicePushBlock{},
				ReceiptService:        &mockSpineReceiptServiceReturnEmpty{},
				ActionTypeSwitcher:    &mockSpineTypeActionSuccess{},
				SpinePublicKeyService: &mockSpineGenerateBlockSpinePublicKeyServiceSuccess{},
				SpineBlockManifestService: &mockSpineBlockManifestService{
					ResSpineBlockManifests: []*model.SpineBlockManifest{
						{
							ID:                      1,
							FullFileHash:            make([]byte, 64),
							FileChunkHashes:         make([]byte, 0),
							ManifestReferenceHeight: 720,
							SpineBlockManifestType:  model.SpineBlockManifestType_Snapshot,
							ExpirationTimestamp:     int64(1000),
						},
					},
				},
				MainBlockService: &mockSpineGenerateblockMainBlockServiceSuccess{},
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
				Chaintype:                 tt.fields.Chaintype,
				QueryExecutor:             tt.fields.QueryExecutor,
				BlockQuery:                tt.fields.BlockQuery,
				Signature:                 tt.fields.Signature,
				BlocksmithStrategy:        tt.fields.BlocksmithStrategy,
				SpinePublicKeyService:     tt.fields.SpinePublicKeyService,
				SpineBlockManifestService: tt.fields.SpineBlockManifestService,
				MainBlockService:          tt.fields.MainBlockService,
			}
			_, err := bs.GenerateBlock(
				tt.args.previousBlock,
				tt.args.secretPhrase,
				tt.args.timestamp,
				tt.args.empty,
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

func (*mockSpineAddGenesisExecutor) BeginTx(bool, int) error { return nil }
func (*mockSpineAddGenesisExecutor) RollbackTx(bool) error   { return nil }
func (*mockSpineAddGenesisExecutor) CommitTx(bool) error     { return nil }
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

func (*mockSpineBlocksmithServiceAddGenesisSuccess) SortBlocksmiths(block *model.Block, withLock bool) {

}

type (
	mockAddGenesisBlockMainBlockServiceSuccess struct {
		BlockServiceInterface
	}
)

func (*mockAddGenesisBlockMainBlockServiceSuccess) GenerateGenesisBlock(genesisEntries []constant.GenesisConfigEntry) (*model.Block, error) {
	return &model.Block{}, nil
}

func TestBlockSpineService_AddGenesis(t *testing.T) {
	type fields struct {
		Chaintype                 chaintype.ChainType
		QueryExecutor             query.ExecutorInterface
		BlockQuery                query.BlockQueryInterface
		MempoolQuery              query.MempoolQueryInterface
		TransactionQuery          query.TransactionQueryInterface
		SpinePublicKeyQuery       query.SpinePublicKeyQueryInterface
		AccountBalanceQuery       query.AccountBalanceQueryInterface
		Signature                 crypto.SignatureInterface
		MempoolService            MempoolServiceInterface
		ActionTypeSwitcher        transaction.TypeActionSwitcher
		Observer                  *observer.Observer
		NodeRegistrationService   NodeRegistrationServiceInterface
		BlocksmithStrategy        strategy.BlocksmithStrategyInterface
		Logger                    *log.Logger
		SpinePublicKeyService     BlockSpinePublicKeyServiceInterface
		SpineBlockManifestService SpineBlockManifestServiceInterface
		BlockStateStorage         storage.CacheStorageInterface
		BlocksStorage             storage.CacheStackStorageInterface
		BlockchainStatusService   BlockchainStatusServiceInterface
		MainBlockService          BlockServiceInterface
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
				SpinePublicKeyService: &BlockSpinePublicKeyService{
					Logger:                log.New(),
					NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
					QueryExecutor:         &mockSpineAddGenesisExecutor{},
					Signature:             nil,
					SpinePublicKeyQuery:   query.NewSpinePublicKeyQuery(),
				},
				SpineBlockManifestService: &SpineBlockManifestService{
					QueryExecutor:           &mockSpineAddGenesisExecutor{},
					Logger:                  log.New(),
					SpineBlockManifestQuery: query.NewSpineBlockManifestQuery(),
					SpineBlockQuery:         query.NewBlockQuery(&chaintype.SpineChain{}),
				},
				BlockStateStorage:       storage.NewBlockStateStorage(),
				BlocksStorage:           storage.NewBlocksStorage(monitoring.TypeSpineBlocksCacheStorage),
				BlockchainStatusService: &mockBlockchainStatusService{},
				MainBlockService:        &mockAddGenesisBlockMainBlockServiceSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:                 tt.fields.Chaintype,
				QueryExecutor:             tt.fields.QueryExecutor,
				BlockQuery:                tt.fields.BlockQuery,
				Signature:                 tt.fields.Signature,
				Observer:                  tt.fields.Observer,
				BlocksmithStrategy:        tt.fields.BlocksmithStrategy,
				Logger:                    tt.fields.Logger,
				SpinePublicKeyService:     tt.fields.SpinePublicKeyService,
				SpineBlockManifestService: tt.fields.SpineBlockManifestService,
				BlockStateStorage:         tt.fields.BlockStateStorage,
				BlockchainStatusService:   tt.fields.BlockchainStatusService,
				MainBlockService:          tt.fields.MainBlockService,
				BlocksStorage:             tt.fields.BlocksStorage,
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
		mockSpineBlockData.GetHeight(),
		mockSpineBlockData.GetID(),
		mockSpineBlockData.GetBlockHash(),
		mockSpineBlockData.GetPreviousBlockHash(),
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
		mockSpineBlockData.GetMerkleRoot(),
		mockSpineBlockData.GetMerkleTree(),
		mockSpineBlockData.GetReferenceBlockHeight(),
	))
	return db.Query("")
}

func (*mockSpineQueryExecutorCheckGenesisTrue) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mockSpine, _ := sqlmock.New()
	mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.SpineChain{}).Fields,
		).AddRow(
			mockSpineBlockData.GetHeight(),
			constant.SpinechainGenesisBlockID,
			mockSpineBlockData.GetBlockHash(),
			mockSpineBlockData.GetPreviousBlockHash(),
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
			mockSpineBlockData.GetMerkleRoot(),
			mockSpineBlockData.GetMerkleTree(),
			mockSpineBlockData.GetReferenceBlockHeight(),
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
			if got, _ := bs.CheckGenesis(); got != tt.want {
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

func (*mockSpineQueryExecutorGetBlockByHeightSuccess) ExecuteSelectRow(qStr string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()

	mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
		query.NewBlockQuery(&chaintype.SpineChain{}).Fields).AddRow(
		mockSpineBlockData.GetHeight(),
		mockSpineBlockData.GetID(),
		mockSpineBlockData.GetBlockHash(),
		mockSpineBlockData.GetPreviousBlockHash(),
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
		mockSpineBlockData.GetMerkleRoot(),
		mockSpineBlockData.GetMerkleTree(),
		mockSpineBlockData.GetReferenceBlockHeight(),
	))
	return db.QueryRow(qStr), nil
}

func (*mockSpineQueryExecutorGetBlockByHeightSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()

	switch qStr {

	case "SELECT node_public_key, node_id, public_key_action, latest, height FROM spine_public_key " +
		"WHERE height >= 0 AND height <= 0 AND public_key_action=0 AND latest=1 ORDER BY height":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
			query.NewSpinePublicKeyQuery().Fields))
	case "SELECT node_public_key, node_id, public_key_action, main_block_height, latest, height FROM spine_public_key WHERE height = 1":
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

func (*mockSpineQueryExecutorGetBlockByHeightFail) ExecuteSelectRow(qStr string, _ bool, _ ...interface{}) (*sql.Row, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()

	mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
		query.NewBlockQuery(&chaintype.SpineChain{}).Fields))
	return db.QueryRow(qStr), nil
}

func (*mockSpineQueryExecutorGetBlockByHeightFail) ExecuteSelect(string, bool, ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}

func TestBlockSpineService_GetBlockByHeight(t *testing.T) {
	type fields struct {
		Chaintype                 chaintype.ChainType
		QueryExecutor             query.ExecutorInterface
		BlockQuery                query.BlockQueryInterface
		MempoolQuery              query.MempoolQueryInterface
		TransactionQuery          query.TransactionQueryInterface
		SpinePublicKeyQuery       query.SpinePublicKeyQueryInterface
		Signature                 crypto.SignatureInterface
		MempoolService            MempoolServiceInterface
		ActionTypeSwitcher        transaction.TypeActionSwitcher
		AccountBalanceQuery       query.AccountBalanceQueryInterface
		Observer                  *observer.Observer
		SpinePublicKeyService     BlockSpinePublicKeyServiceInterface
		SpineBlockManifestService SpineBlockManifestServiceInterface
	}
	type args struct {
		height uint32
	}
	mockSpineBlockData.SpineBlockManifests = make([]*model.SpineBlockManifest, 0)
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
				Chaintype:        &chaintype.SpineChain{},
				QueryExecutor:    &mockSpineQueryExecutorGetBlockByHeightSuccess{},
				BlockQuery:       query.NewBlockQuery(&chaintype.SpineChain{}),
				TransactionQuery: query.NewTransactionQuery(&chaintype.SpineChain{}),
				SpinePublicKeyService: &BlockSpinePublicKeyService{
					Logger:                log.New(),
					NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
					QueryExecutor:         &mockSpineQueryExecutorGetBlockByHeightSuccess{},
					Signature:             nil,
					SpinePublicKeyQuery:   query.NewSpinePublicKeyQuery(),
				},
				SpineBlockManifestService: &mockSpineBlockManifestService{},
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
				SpinePublicKeyService: &BlockSpinePublicKeyService{
					Logger:                log.New(),
					NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
					QueryExecutor:         &mockSpineQueryExecutorGetBlockByHeightFail{},
					Signature:             nil,
					SpinePublicKeyQuery:   query.NewSpinePublicKeyQuery(),
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:                 tt.fields.Chaintype,
				QueryExecutor:             tt.fields.QueryExecutor,
				BlockQuery:                tt.fields.BlockQuery,
				Signature:                 tt.fields.Signature,
				Observer:                  tt.fields.Observer,
				SpinePublicKeyService:     tt.fields.SpinePublicKeyService,
				SpineBlockManifestService: tt.fields.SpineBlockManifestService,
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
	case "SELECT node_public_key, node_id, public_key_action, latest, height FROM spine_public_key " +
		"WHERE height >= 0 AND height <= 1 AND public_key_action=0 AND latest=1 ORDER BY height":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(query.NewSpinePublicKeyQuery().Fields))
	case "SELECT node_public_key, node_id, public_key_action, main_block_height, latest, height FROM spine_public_key WHERE height = 1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(query.NewSpinePublicKeyQuery().Fields))
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, " +
		"version, merkle_root, merkle_tree, reference_block_height FROM spine_block WHERE id = 1":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.SpineChain{}).Fields).AddRow(
			mockSpineBlockData.GetHeight(),
			mockSpineBlockData.GetID(),
			mockSpineBlockData.GetBlockHash(),
			mockSpineBlockData.GetPreviousBlockHash(),
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
			mockSpineBlockData.GetMerkleRoot(),
			mockSpineBlockData.GetMerkleTree(),
			mockSpineBlockData.GetReferenceBlockHeight(),
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
			mockSpineBlockData.GetHeight(),
			mockSpineBlockData.GetID(),
			mockSpineBlockData.GetBlockHash(),
			mockSpineBlockData.GetPreviousBlockHash(),
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
			mockSpineBlockData.GetMerkleRoot(),
			mockSpineBlockData.GetMerkleTree(),
			mockSpineBlockData.GetReferenceBlockHeight(),
		))
	return db.QueryRow(qStr), nil
}

func (*mockSpineQueryExecutorGetBlockByIDFail) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, errors.New("MockedError")
}

func TestBlockSpineService_GetBlockByID(t *testing.T) {
	var mockData = mockSpineBlockData
	mockData.SpineBlockManifests = []*model.SpineBlockManifest{
		ssMockSpineBlockManifest,
	}
	type fields struct {
		Chaintype                 chaintype.ChainType
		QueryExecutor             query.ExecutorInterface
		BlockQuery                query.BlockQueryInterface
		MempoolQuery              query.MempoolQueryInterface
		TransactionQuery          query.TransactionQueryInterface
		SpinePublicKeyQuery       query.SpinePublicKeyQueryInterface
		Signature                 crypto.SignatureInterface
		MempoolService            MempoolServiceInterface
		ActionTypeSwitcher        transaction.TypeActionSwitcher
		AccountBalanceQuery       query.AccountBalanceQueryInterface
		Observer                  *observer.Observer
		SpinePublicKeyService     BlockSpinePublicKeyServiceInterface
		SpineBlockManifestService SpineBlockManifestServiceInterface
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
				Chaintype:        &chaintype.SpineChain{},
				QueryExecutor:    &mockSpineQueryExecutorGetBlockByIDSuccess{},
				BlockQuery:       query.NewBlockQuery(&chaintype.SpineChain{}),
				TransactionQuery: query.NewTransactionQuery(&chaintype.SpineChain{}),
				SpinePublicKeyService: &BlockSpinePublicKeyService{
					Logger:                log.New(),
					NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
					QueryExecutor:         &mockSpineQueryExecutorGetBlockByIDSuccess{},
					Signature:             nil,
					SpinePublicKeyQuery:   query.NewSpinePublicKeyQuery(),
				},
				SpineBlockManifestService: &mockSpineBlockManifestService{
					ResSpineBlockManifests: []*model.SpineBlockManifest{
						ssMockSpineBlockManifest,
					},
				},
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
				Chaintype:     &chaintype.SpineChain{},
				QueryExecutor: &mockSpineQueryExecutorGetBlockByIDFail{},
				BlockQuery:    query.NewBlockQuery(&chaintype.SpineChain{}),
				SpinePublicKeyService: &BlockSpinePublicKeyService{
					Logger:                log.New(),
					NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
					QueryExecutor:         &mockSpineQueryExecutorGetBlockByIDFail{},
					Signature:             nil,
					SpinePublicKeyQuery:   query.NewSpinePublicKeyQuery(),
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:                 tt.fields.Chaintype,
				QueryExecutor:             tt.fields.QueryExecutor,
				BlockQuery:                tt.fields.BlockQuery,
				Signature:                 tt.fields.Signature,
				Observer:                  tt.fields.Observer,
				SpinePublicKeyService:     tt.fields.SpinePublicKeyService,
				SpineBlockManifestService: tt.fields.SpineBlockManifestService,
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
		mockSpineBlockData.GetHeight(),
		mockSpineBlockData.GetID(),
		mockSpineBlockData.GetBlockHash(),
		mockSpineBlockData.GetPreviousBlockHash(),
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
		mockSpineBlockData.GetMerkleRoot(),
		mockSpineBlockData.GetMerkleTree(),
		mockSpineBlockData.GetReferenceBlockHeight(),
	).AddRow(
		mockSpineBlockData.GetHeight(),
		mockSpineBlockData.GetID(),
		mockSpineBlockData.GetBlockHash(),
		mockSpineBlockData.GetPreviousBlockHash(),
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
		mockSpineBlockData.GetMerkleRoot(),
		mockSpineBlockData.GetMerkleTree(),
		mockSpineBlockData.GetReferenceBlockHeight(),
	),
	)
	return db.Query(qStr)
}

func (*mockSpineQueryExecutorGetBlocksFromHeightFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}

func TestBlockSpineService_GetBlocksFromHeight(t *testing.T) {
	mockSpineBlockData.SpineBlockManifests = nil
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
				Chaintype:     &chaintype.SpineChain{},
				QueryExecutor: &mockSpineQueryExecutorGetBlocksFromHeightSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.SpineChain{}),
			},
			args: args{
				startHeight:      0,
				limit:            2,
				withAttachedData: false,
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
			got, err := bs.GetBlocksFromHeight(tt.args.startHeight, tt.args.limit, tt.args.withAttachedData)
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

type (
	mockSpineReceiveBlockBlockStateStorageSuccess struct {
		storage.CacheStorageInterface
	}
)

func (*mockSpineReceiveBlockBlockStateStorageSuccess) GetItem(lastChange, item interface{}) error {
	var blockCopy, _ = item.(*model.Block)
	*blockCopy = mockSpineBlockData
	return nil
}

func (*mockSpineReceiveBlockBlockStateStorageSuccess) SetItem(lastChange, item interface{}) error {
	return nil
}

type (
	mockReceiveBlockMainBlockServiceSuccess struct {
		BlockServiceInterface
	}
)

func (*mockReceiveBlockMainBlockServiceSuccess) GetBlockByHeightCacheFormat(uint32) (*storage.BlockCacheObject, error) {
	return &storage.BlockCacheObject{}, nil
}

func (*mockReceiveBlockMainBlockServiceSuccess) GetLastBlockCacheFormat() (*storage.BlockCacheObject, error) {
	return &storage.BlockCacheObject{}, nil
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
			NodePublicKey: []byte{1, 3, 4, 5, 6},
		},
		{
			NodeID: 2,
		},
		{
			NodeID: 3,
		},
	}
}

type (
	mockGenerateGenesisBlockMainBlockServiceSuccess struct {
		BlockServiceInterface
	}
)

func (*mockGenerateGenesisBlockMainBlockServiceSuccess) GenerateGenesisBlock(genesisEntries []constant.GenesisConfigEntry) (*model.Block, error) {
	return &model.Block{}, nil
}
func TestBlockSpineService_GenerateGenesisBlock(t *testing.T) {
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
		MainBlockService        BlockServiceInterface
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
				MainBlockService:        &mockGenerateGenesisBlockMainBlockServiceSuccess{},
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
				Chaintype:        tt.fields.Chaintype,
				QueryExecutor:    tt.fields.QueryExecutor,
				BlockQuery:       tt.fields.BlockQuery,
				Signature:        tt.fields.Signature,
				Observer:         tt.fields.Observer,
				Logger:           tt.fields.Logger,
				MainBlockService: tt.fields.MainBlockService,
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
			mockSpineBlockData.GetHeight(),
			mockSpineBlockData.GetID(),
			mockSpineBlockData.GetBlockHash(),
			mockSpineBlockData.GetPreviousBlockHash(),
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
			mockSpineBlockData.GetMerkleRoot(),
			mockSpineBlockData.GetMerkleTree(),
			mockSpineBlockData.GetReferenceBlockHeight(),
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
)

type (
	mockSpineBlocksmithServiceValidateBlockSuccess struct {
		strategy.BlocksmithStrategyMain
	}
	mockValidateBlockMainBlockServiceSuccess struct {
		BlockServiceInterface
	}
)

func (*mockValidateBlockMainBlockServiceSuccess) GetBlockByHeightCacheFormat(uint32) (*storage.BlockCacheObject, error) {
	return &storage.BlockCacheObject{}, nil
}

func (*mockValidateBlockMainBlockServiceSuccess) GetLastBlockCacheFormat() (*storage.BlockCacheObject, error) {
	return &storage.BlockCacheObject{}, nil
}

func (*mockSpineBlocksmithServiceValidateBlockSuccess) GetSortedBlocksmithsMap(*model.Block) map[string]*int64 {
	firstIndex := int64(0)
	secondIndex := int64(1)
	return map[string]*int64{
		string(mockSpineValidateBadBlockInvalidBlockHash.BlocksmithPublicKey): &firstIndex,
		string(mockSpineBlockData.BlocksmithPublicKey):                        &secondIndex,
	}
}
func (*mockSpineBlocksmithServiceValidateBlockSuccess) IsBlockTimestampValid(blocksmithIndex, numberOfBlocksmiths int64, previousBlock,
	currentBlock *model.Block) error {
	return nil
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

func (*mockSpinePopOffToBlockReturnCommonBlock) BeginTx(bool, int) error {
	return nil
}
func (*mockSpinePopOffToBlockReturnCommonBlock) CommitTx(bool) error {
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
func (*mockSpinePopOffToBlockReturnBeginTxFunc) BeginTx(bool, int) error {
	return errors.New("i want this")
}
func (*mockSpinePopOffToBlockReturnBeginTxFunc) CommitTx(bool) error {
	return nil
}
func (*mockSpinePopOffToBlockReturnWantFailOnCommit) BeginTx(bool, int) error {
	return nil
}
func (*mockSpinePopOffToBlockReturnWantFailOnCommit) CommitTx(bool) error {
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
func (*mockSpinePopOffToBlockReturnWantFailOnExecuteTransactions) BeginTx(bool, int) error {
	return nil
}
func (*mockSpinePopOffToBlockReturnWantFailOnExecuteTransactions) CommitTx(bool) error {
	return nil
}
func (*mockSpinePopOffToBlockReturnWantFailOnExecuteTransactions) ExecuteTransactions(queries [][]interface{}) error {
	return errors.New("i want this")
}
func (*mockSpinePopOffToBlockReturnWantFailOnExecuteTransactions) RollbackTx(bool) error {
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
)

func (*mockSpineExecutorBlockPopFailCommonNotFound) ExecuteSelectRow(
	qStr string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	blockQ := query.NewBlockQuery(&chaintype.SpineChain{})
	switch qStr {
	case "SELECT MAX(height), id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height FROM spine_block":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields))
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height FROM spine_block WHERE id = 1":
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
	case "SELECT MAX(height), id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height FROM spine_block":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields).AddRow(
				mockSpineGoodBlock.GetHeight(),
				mockSpineGoodBlock.GetID(),
				mockSpineGoodBlock.GetBlockHash(),
				mockSpineGoodBlock.GetPreviousBlockHash(),
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
				mockSpineGoodBlock.GetMerkleRoot(),
				mockSpineGoodBlock.GetMerkleTree(),
				mockSpineGoodBlock.GetReferenceBlockHeight(),
			),
		)
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height FROM spine_block WHERE id = 0":
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
	case "SELECT MAX(height), id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version, merkle_root, merkle_tree, " +
		"reference_block_height FROM spine_block":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnError(
			errors.New("MockErr"))
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version FROM main_block WHERE id = 0":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields))
	default:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields[:len(blockQ.Fields)-1]).AddRow(
				mockSpineGoodBlock.GetHeight(),
				mockSpineGoodBlock.GetID(),
				mockSpineGoodBlock.GetBlockHash(),
				mockSpineGoodBlock.GetPreviousBlockHash(),
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

func (*mockSpineMempoolServiceBlockPopSuccess) GetMempoolTransactionsWantToBackup(height uint32) ([]*model.Transaction, error) {
	return make([]*model.Transaction, 0), nil
}

func (*mockSpineMempoolServiceBlockPopFail) GetMempoolTransactionsWantToBackup(height uint32) ([]*model.Transaction, error) {
	return nil, errors.New("mockSpineedError")
}

func (*mockSpineReceiptSuccess) GetPublishedReceiptsByHeight(blockHeight uint32) ([]*model.PublishedReceipt, error) {
	return make([]*model.PublishedReceipt, 0), nil
}

func (*mockSpineReceiptFail) GetPublishedReceiptsByHeight(blockHeight uint32) ([]*model.PublishedReceipt, error) {
	return nil, errors.New("mockSpineError")
}

func (*mockSpineExecutorBlockPopSuccess) BeginTx(bool, int) error {
	return nil
}

func (*mockSpineExecutorBlockPopSuccess) CommitTx(bool) error {
	return nil
}

func (*mockSpineExecutorBlockPopSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}
func (*mockSpineExecutorBlockPopSuccess) RollbackTx(bool) error {
	return nil
}
func (*mockSpineExecutorBlockPopSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()

	transactionQ := query.NewTransactionQuery(&chaintype.SpineChain{})
	blockQ := query.NewBlockQuery(&chaintype.SpineChain{})
	spinePubKeyQ := query.NewSpinePublicKeyQuery()
	switch qStr {
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height FROM spine_block WHERE id = 0":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields).AddRow(
				mockSpineGoodCommonBlock.GetHeight(),
				mockSpineGoodCommonBlock.GetID(),
				mockSpineGoodCommonBlock.GetBlockHash(),
				mockSpineGoodCommonBlock.GetPreviousBlockHash(),
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
				mockSpineGoodCommonBlock.GetMerkleRoot(),
				mockSpineGoodCommonBlock.GetMerkleTree(),
				mockSpineGoodCommonBlock.GetReferenceBlockHeight(),
			),
		)
	case "SELECT node_public_key, node_id, public_key_action, latest, height FROM spine_public_key " +
		"WHERE height >= 0 AND height <= 1000 AND public_key_action=0 AND latest=1 ORDER BY height":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(spinePubKeyQ.Fields))
	case "SELECT node_public_key, node_id, public_key_action, main_block_height, latest, height FROM spine_public_key WHERE height = 1000":
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
			mockSpineGoodBlock.GetHeight(),
			mockSpineGoodBlock.GetID(),
			mockSpineGoodBlock.GetBlockHash(),
			mockSpineGoodBlock.GetPreviousBlockHash(),
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
			mockSpineGoodBlock.GetMerkleRoot(),
			mockSpineGoodBlock.GetMerkleTree(),
			mockSpineGoodBlock.GetReferenceBlockHeight(),
		),
	)
	return db.QueryRow(qStr), nil
}

type (
	// mock PopOffToBlock
	mockSnapshotMainBlockServiceDeleteFail struct {
		SnapshotMainBlockService
	}
	mockSnapshotMainBlockServiceDeleteSuccess struct {
		SnapshotMainBlockService
	}

	mockSpineBlockManifestServiceFailGetManifestFromHeight struct {
		SpineBlockManifestServiceInterface
	}

	mockSpineBlockManifestServiceSuccesGetManifestFromHeight struct {
		SpineBlockManifestServiceInterface
	}
	mockSpineExecutorBlockPopSuccessPoppedBlocks struct {
		query.Executor
	}
	mockSpinePopOffBlockBlockStateStorageSuccess struct {
		storage.CacheStorageInterface
	}
	mockSpinePopOffBlockBlocksStorageSuccess struct {
		storage.CacheStackStorageInterface
	}
	mockSpinePopOffBlockBlockStateStorageFail struct {
		storage.CacheStorageInterface
	}
)

func (*mockSpinePopOffBlockBlockStateStorageSuccess) GetItem(lastChange, item interface{}) error {
	var blockCopy, _ = item.(*model.Block)
	*blockCopy = *mockSpineGoodCommonBlock
	return nil
}

func (*mockSpinePopOffBlockBlockStateStorageSuccess) SetItem(lastChange, item interface{}) error {
	return nil
}

func (*mockSpinePopOffBlockBlockStateStorageFail) GetItem(lastChange, item interface{}) error {
	return errors.New("mockedError")
}

func (mockSpinePopOffBlockBlocksStorageSuccess) Pop() error {
	return nil
}

func (mockSpinePopOffBlockBlocksStorageSuccess) Push(interface{}) error {
	return nil
}
func (mockSpinePopOffBlockBlocksStorageSuccess) PopTo(uint32) error {
	return nil
}
func (mockSpinePopOffBlockBlocksStorageSuccess) GetAll(interface{}) error {
	return nil
}
func (mockSpinePopOffBlockBlocksStorageSuccess) GetAtIndex(uint32, interface{}) error {
	return nil
}
func (mockSpinePopOffBlockBlocksStorageSuccess) GetTop(interface{}) error {
	return nil
}

// Clear clean up the whole stack and reinitialize with new array
func (mockSpinePopOffBlockBlocksStorageSuccess) Clear() error {
	return nil
}

func (*mockSnapshotMainBlockServiceDeleteFail) DeleteFileByChunkHashes([]byte) error {
	return errors.New("mockedError")
}

func (*mockSnapshotMainBlockServiceDeleteSuccess) DeleteFileByChunkHashes([]byte) error {
	return nil
}

func (*mockSpineBlockManifestServiceFailGetManifestFromHeight) GetSpineBlockManifestsFromSpineBlockHeight(
	uint32,
) ([]*model.SpineBlockManifest, error) {
	return []*model.SpineBlockManifest{}, errors.New("mockedError")
}

func (*mockSpineBlockManifestServiceFailGetManifestFromHeight) GetSpineBlockManifestBySpineBlockHeight(uint32) (
	[]*model.SpineBlockManifest, error,
) {
	return make([]*model.SpineBlockManifest, 0), nil
}

func (*mockSpineBlockManifestServiceSuccesGetManifestFromHeight) GetSpineBlockManifestsFromSpineBlockHeight(
	uint32,
) ([]*model.SpineBlockManifest, error) {
	return []*model.SpineBlockManifest{
		ssMockSpineBlockManifest,
	}, nil
}

func (*mockSpineBlockManifestServiceSuccesGetManifestFromHeight) GetSpineBlockManifestBySpineBlockHeight(uint32) (
	[]*model.SpineBlockManifest, error,
) {
	return make([]*model.SpineBlockManifest, 0), nil
}

func (*mockSpineExecutorBlockPopSuccessPoppedBlocks) BeginTx(bool, int) error {
	return nil
}

func (*mockSpineExecutorBlockPopSuccessPoppedBlocks) CommitTx(bool) error {
	return nil
}
func (*mockSpineExecutorBlockPopSuccessPoppedBlocks) ExecuteTransactions([][]interface{}) error {
	return nil
}
func (*mockSpineExecutorBlockPopSuccessPoppedBlocks) RollbackTx(bool) error {
	return nil
}

func (*mockSpineExecutorBlockPopSuccessPoppedBlocks) ExecuteSelectRow(qStr string, _ bool, _ ...interface{}) (*sql.Row, error) {
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

func TestBlockSpineService_PopOffToBlock(t *testing.T) {
	type fields struct {
		Chaintype                 chaintype.ChainType
		QueryExecutor             query.ExecutorInterface
		BlockQuery                query.BlockQueryInterface
		SpinePublicKeyQuery       query.SpinePublicKeyQueryInterface
		MempoolQuery              query.MempoolQueryInterface
		TransactionQuery          query.TransactionQueryInterface
		MerkleTreeQuery           query.MerkleTreeQueryInterface
		PublishedReceiptQuery     query.PublishedReceiptQueryInterface
		SkippedBlocksmithQuery    query.SkippedBlocksmithQueryInterface
		Signature                 crypto.SignatureInterface
		MempoolService            MempoolServiceInterface
		ReceiptService            ReceiptServiceInterface
		NodeRegistrationService   NodeRegistrationServiceInterface
		ActionTypeSwitcher        transaction.TypeActionSwitcher
		AccountBalanceQuery       query.AccountBalanceQueryInterface
		ParticipationScoreQuery   query.ParticipationScoreQueryInterface
		NodeRegistrationQuery     query.NodeRegistrationQueryInterface
		Observer                  *observer.Observer
		Logger                    *log.Logger
		SpinePublicKeyService     BlockSpinePublicKeyServiceInterface
		SpineBlockManifestService SpineBlockManifestServiceInterface
		SnapshotMainBlockService  SnapshotBlockServiceInterface
		BlockStateStorage         storage.CacheStorageInterface
		BlocksStorage             storage.CacheStackStorageInterface
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
				Chaintype:               &chaintype.SpineChain{},
				QueryExecutor:           &mockSpineExecutorBlockPopGetLastBlockFail{},
				BlockQuery:              query.NewBlockQuery(&chaintype.SpineChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.SpineChain{}),
				MerkleTreeQuery:         nil,
				PublishedReceiptQuery:   nil,
				SkippedBlocksmithQuery:  nil,
				Signature:               nil,
				MempoolService:          &mockSpineMempoolServiceBlockPopSuccess{},
				ReceiptService:          &mockSpineReceiptSuccess{},
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				ParticipationScoreQuery: nil,
				Observer:                nil,
				Logger:                  log.New(),
				SpinePublicKeyService: &BlockSpinePublicKeyService{
					Logger:                log.New(),
					NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
					QueryExecutor:         &mockSpineExecutorBlockPopGetLastBlockFail{},
					Signature:             nil,
					SpinePublicKeyQuery:   query.NewSpinePublicKeyQuery(),
				},
				SpineBlockManifestService: &mockSpineBlockManifestService{},
				BlockStateStorage:         &mockSpinePopOffBlockBlockStateStorageFail{},
			},
			args: args{
				commonBlock: mockSpineGoodCommonBlock,
			},
			want:    make([]*model.Block, 0),
			wantErr: true,
		},
		{
			name: "Fail-HardFork",
			fields: fields{
				Chaintype:               &chaintype.SpineChain{},
				QueryExecutor:           &mockSpineExecutorBlockPopSuccess{},
				BlockQuery:              query.NewBlockQuery(&chaintype.SpineChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.SpineChain{}),
				MerkleTreeQuery:         nil,
				PublishedReceiptQuery:   nil,
				SkippedBlocksmithQuery:  nil,
				Signature:               nil,
				MempoolService:          &mockSpineMempoolServiceBlockPopSuccess{},
				ReceiptService:          &mockSpineReceiptSuccess{},
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				ParticipationScoreQuery: nil,
				Observer:                nil,
				Logger:                  log.New(),
				SpinePublicKeyService: &BlockSpinePublicKeyService{
					Logger:                log.New(),
					NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
					QueryExecutor:         &mockSpineExecutorBlockPopSuccess{},
					Signature:             nil,
					SpinePublicKeyQuery:   query.NewSpinePublicKeyQuery(),
				},
				SpineBlockManifestService: &mockSpineBlockManifestService{},
				BlockStateStorage:         &mockSpinePopOffBlockBlockStateStorageSuccess{},
			},
			args: args{
				commonBlock: mockSpineBadCommonBlockHardFork,
			},
			want:    make([]*model.Block, 0),
			wantErr: false,
		},
		{
			name: "Fail-CommonBlockNotFound",
			fields: fields{
				Chaintype:               &chaintype.SpineChain{},
				QueryExecutor:           &mockSpineExecutorBlockPopFailCommonNotFound{},
				BlockQuery:              query.NewBlockQuery(&chaintype.SpineChain{}),
				MempoolQuery:            nil,
				TransactionQuery:        query.NewTransactionQuery(&chaintype.SpineChain{}),
				MerkleTreeQuery:         nil,
				PublishedReceiptQuery:   nil,
				SkippedBlocksmithQuery:  nil,
				Signature:               nil,
				MempoolService:          &mockSpineMempoolServiceBlockPopSuccess{},
				ReceiptService:          &mockSpineReceiptSuccess{},
				ActionTypeSwitcher:      nil,
				AccountBalanceQuery:     nil,
				ParticipationScoreQuery: nil,
				NodeRegistrationQuery:   nil,
				Observer:                nil,
				Logger:                  log.New(),
				SpinePublicKeyService: &BlockSpinePublicKeyService{
					Logger:                log.New(),
					NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
					QueryExecutor:         &mockSpineExecutorBlockPopFailCommonNotFound{},
					Signature:             nil,
					SpinePublicKeyQuery:   query.NewSpinePublicKeyQuery(),
				},
				SpineBlockManifestService: &mockSpineBlockManifestService{},
				BlockStateStorage:         &mockSpinePopOffBlockBlockStateStorageSuccess{},
				BlocksStorage:             &mockSpinePopOffBlockBlocksStorageSuccess{},
			},
			args: args{
				commonBlock: mockSpineGoodCommonBlock,
			},
			want:    make([]*model.Block, 0),
			wantErr: true,
		},
		{
			name: "GetManifestFromSpineBlockHeight-Success",
			fields: fields{
				Chaintype:                 &chaintype.SpineChain{},
				QueryExecutor:             &mockSpineExecutorBlockPopSuccess{},
				BlockQuery:                query.NewBlockQuery(&chaintype.SpineChain{}),
				MempoolQuery:              nil,
				TransactionQuery:          query.NewTransactionQuery(&chaintype.SpineChain{}),
				MerkleTreeQuery:           nil,
				PublishedReceiptQuery:     nil,
				SkippedBlocksmithQuery:    nil,
				Signature:                 nil,
				MempoolService:            &mockSpineMempoolServiceBlockPopSuccess{},
				ReceiptService:            &mockSpineReceiptSuccess{},
				ActionTypeSwitcher:        nil,
				AccountBalanceQuery:       nil,
				ParticipationScoreQuery:   nil,
				Observer:                  nil,
				Logger:                    log.New(),
				SpinePublicKeyService:     &mockBlockSpinePublicKeyService{},
				SpineBlockManifestService: &mockSpineBlockManifestServiceSuccesGetManifestFromHeight{},
				SnapshotMainBlockService:  &mockSnapshotMainBlockServiceDeleteSuccess{},
				BlockStateStorage:         &mockSpinePopOffBlockBlockStateStorageSuccess{},
				BlocksStorage:             &mockSpinePopOffBlockBlocksStorageSuccess{},
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
				Chaintype:                 tt.fields.Chaintype,
				QueryExecutor:             tt.fields.QueryExecutor,
				BlockQuery:                tt.fields.BlockQuery,
				Signature:                 tt.fields.Signature,
				Observer:                  tt.fields.Observer,
				Logger:                    tt.fields.Logger,
				SpinePublicKeyService:     tt.fields.SpinePublicKeyService,
				SpineBlockManifestService: tt.fields.SpineBlockManifestService,
				SnapshotMainBlockService:  tt.fields.SnapshotMainBlockService,
				BlockStateStorage:         tt.fields.BlockStateStorage,
				BlocksStorage:             tt.fields.BlocksStorage,
			}
			got, err := bs.PopOffToBlock(tt.args.commonBlock)
			if (err != nil) != tt.wantErr {
				t.Errorf("PopOffToBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PopOffToBlock() got = \n%v, want \n%v", got, tt.want)
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
	return nil, errors.New("MockError")
}

func (*mockSpineExecutorPopulateBlockDataSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mockSpine, _ := sqlmock.New()
	defer db.Close()
	switch qStr {
	case "SELECT node_public_key, node_id, public_key_action, main_block_height, latest, height FROM spine_public_key " +
		"WHERE height = 0":
		mockSpine.ExpectQuery(regexp.QuoteMeta(qStr)).
			WillReturnRows(sqlmock.NewRows(
				query.NewSpinePublicKeyQuery().Fields,
			).AddRow(
				mockSpinePublicKey.NodePublicKey,
				mockSpinePublicKey.NodeID,
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
		Chaintype                 chaintype.ChainType
		QueryExecutor             query.ExecutorInterface
		BlockQuery                query.BlockQueryInterface
		SpinePublicKeyQuery       query.SpinePublicKeyQueryInterface
		Signature                 crypto.SignatureInterface
		NodeRegistrationQuery     query.NodeRegistrationQueryInterface
		BlocksmithStrategy        strategy.BlocksmithStrategyInterface
		Observer                  *observer.Observer
		Logger                    *log.Logger
		SpinePublicKeyService     BlockSpinePublicKeyServiceInterface
		SpineBlockManifestService SpineBlockManifestServiceInterface
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
				Chaintype:     &chaintype.SpineChain{},
				QueryExecutor: &mockSpineExecutorPopulateBlockDataFail{},
				Logger:        log.New(),
				SpinePublicKeyService: &BlockSpinePublicKeyService{
					Logger:                log.New(),
					NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
					QueryExecutor:         &mockSpineExecutorPopulateBlockDataFail{},
					Signature:             nil,
					SpinePublicKeyQuery:   query.NewSpinePublicKeyQuery(),
				},
				SpineBlockManifestService: &mockSpineBlockManifestService{},
			},
			args: args{
				block: &model.Block{},
			},
			wantErr: true,
		},
		{
			name: "PopulateBlockData:success",
			fields: fields{
				Chaintype:     &chaintype.SpineChain{},
				QueryExecutor: &mockSpineExecutorPopulateBlockDataSuccess{},
				Logger:        log.New(),
				SpinePublicKeyService: &BlockSpinePublicKeyService{
					Logger:                log.New(),
					NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
					QueryExecutor:         &mockSpineExecutorPopulateBlockDataSuccess{},
					Signature:             nil,
					SpinePublicKeyQuery:   query.NewSpinePublicKeyQuery(),
				},
				SpineBlockManifestService: &mockSpineBlockManifestService{},
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
				SpineBlockManifests: make([]*model.SpineBlockManifest, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:                 tt.fields.Chaintype,
				QueryExecutor:             tt.fields.QueryExecutor,
				BlockQuery:                tt.fields.BlockQuery,
				Signature:                 tt.fields.Signature,
				BlocksmithStrategy:        tt.fields.BlocksmithStrategy,
				Observer:                  tt.fields.Observer,
				Logger:                    tt.fields.Logger,
				SpinePublicKeyService:     tt.fields.SpinePublicKeyService,
				SpineBlockManifestService: tt.fields.SpineBlockManifestService,
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
	mockSpineExecutorValidateSpineBlockManifest struct {
		query.Executor
		success bool
		noRows  bool
	}
)

func (msExQ *mockSpineExecutorValidateSpineBlockManifest) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	if !msExQ.success {
		return nil, errors.New("ExecuteSelectRowFailed")
	}
	if msExQ.noRows {
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(query.NewBlockQuery(&chaintype.SpineChain{}).Fields))
	}
	switch qStr {
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, " +
		"version, merkle_root, merkle_tree, reference_block_height FROM spine_block WHERE timestamp >= 15875392 ORDER BY timestamp LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows(query.NewBlockQuery(&chaintype.SpineChain{}).Fields).
			AddRow(
				mockSpineBlockData.GetHeight(),
				mockSpineBlockData.GetID(),
				mockSpineBlockData.GetBlockHash(),
				mockSpineBlockData.GetPreviousBlockHash(),
				mockSpineBlockData.GetTimestamp(),
				mockSpineBlockData.GetBlockSeed(),
				mockSpineBlockData.GetBlockSignature(),
				mockSpineBlockData.GetCumulativeDifficulty(),
				uint32(8),
				[]byte{28, 122, 181, 212, 11, 147, 147, 173, 220, 102, 150, 8, 100, 164, 82, 120, 228, 253, 53, 160, 5, 21,
					103, 1, 127, 243, 215, 57, 88, 97, 137, 113},
				mockSpineBlockData.GetBlocksmithPublicKey(),
				mockSpineBlockData.GetTotalAmount(),
				mockSpineBlockData.GetTotalFee(),
				mockSpineBlockData.GetTotalCoinBase(),
				mockSpineBlockData.GetVersion(),
				mockSpineBlockData.GetMerkleRoot(),
				mockSpineBlockData.GetMerkleTree(),
				mockSpineBlockData.GetReferenceBlockHeight(),
			))
	default:
		return nil, errors.New("UnmockedQuery")
	}
	row := db.QueryRow(qStr)
	return row, nil
}

func (ss *mockSpineBlockManifestService) GetSpineBlockManifestBytes(spineBlockManifest *model.SpineBlockManifest) []byte {
	if spineBlockManifest.ID == 12345678 || ss.ResSpineBlockManifestBytes != nil {
		return ss.ResSpineBlockManifestBytes
	}
	return []byte{}
}

func TestBlockSpineService_ValidateSpineBlockManifest(t *testing.T) {
	type fields struct {
		Chaintype                 chaintype.ChainType
		QueryExecutor             query.ExecutorInterface
		BlockQuery                query.BlockQueryInterface
		Signature                 crypto.SignatureInterface
		BlocksmithStrategy        strategy.BlocksmithStrategyInterface
		Observer                  *observer.Observer
		Logger                    *log.Logger
		SpinePublicKeyService     BlockSpinePublicKeyServiceInterface
		SpineBlockManifestService SpineBlockManifestServiceInterface
	}
	type args struct {
		spineBlockManifest *model.SpineBlockManifest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ValidateSpineBlockManifest:success",
			fields: fields{
				QueryExecutor: &mockSpineExecutorValidateSpineBlockManifest{
					success: true,
				},
				BlockQuery:            query.NewBlockQuery(&chaintype.SpineChain{}),
				Logger:                log.New(),
				SpinePublicKeyService: &mockBlockSpinePublicKeyService{},
				SpineBlockManifestService: &mockSpineBlockManifestService{
					ResSpineBlockManifestBytes: []byte{1, 1, 1, 1, 1, 1, 1, 1},
					ResSpineBlockManifests: []*model.SpineBlockManifest{
						{
							ID:                      12345678,
							FullFileHash:            make([]byte, 64),
							FileChunkHashes:         make([]byte, 0),
							ManifestReferenceHeight: 720,
							SpineBlockManifestType:  model.SpineBlockManifestType_Snapshot,
							ExpirationTimestamp:     int64(1000),
						},
					},
				},
			},
			args: args{
				spineBlockManifest: &model.SpineBlockManifest{
					ID:                  12345678,
					ExpirationTimestamp: 15875392,
				},
			},
			wantErr: false,
		},
		{
			name: "ValidateSpineBlockManifest:fail-{InvalidSpineBlockManifestTimestamp}",
			fields: fields{
				QueryExecutor: &mockSpineExecutorValidateSpineBlockManifest{
					success: true,
					noRows:  true,
				},
				BlockQuery:            query.NewBlockQuery(&chaintype.SpineChain{}),
				Logger:                log.New(),
				SpinePublicKeyService: &mockBlockSpinePublicKeyService{},
				SpineBlockManifestService: &mockSpineBlockManifestService{
					ResSpineBlockManifestBytes: []byte{1, 1, 1, 1, 1, 1, 1, 1},
					ResSpineBlockManifests: []*model.SpineBlockManifest{
						{
							ID:                     12345678,
							SpineBlockManifestType: model.SpineBlockManifestType_Snapshot,
							ExpirationTimestamp:    int64(1000),
						},
					},
				},
			},
			args: args{
				spineBlockManifest: &model.SpineBlockManifest{
					ID:                  12345678,
					ExpirationTimestamp: 15875392,
				},
			},
			wantErr: true,
		},
		{
			name: "ValidateSpineBlockManifest:fail-{InvalidSpineBlockManifestData}",
			fields: fields{
				QueryExecutor: &mockSpineExecutorValidateSpineBlockManifest{
					success: true,
				},
				BlockQuery:            query.NewBlockQuery(&chaintype.SpineChain{}),
				Logger:                log.New(),
				SpinePublicKeyService: &mockBlockSpinePublicKeyService{},
				SpineBlockManifestService: &mockSpineBlockManifestService{
					ResSpineBlockManifestBytes: []byte{1, 1, 1, 1, 1, 1, 1, 1},
				},
			},
			args: args{
				spineBlockManifest: &model.SpineBlockManifest{
					ID:                  11111111,
					ExpirationTimestamp: 15875392,
				},
			},
			wantErr: true,
		},
		{
			name: "ValidateSpineBlockManifest:fail-{InvalidComputedSpineBlockPayloadHash}",
			fields: fields{
				QueryExecutor: &mockSpineExecutorValidateSpineBlockManifest{
					success: true,
				},
				BlockQuery:            query.NewBlockQuery(&chaintype.SpineChain{}),
				Logger:                log.New(),
				SpinePublicKeyService: &mockBlockSpinePublicKeyService{},
				SpineBlockManifestService: &mockSpineBlockManifestService{
					ResSpineBlockManifestBytes: []byte{1, 1, 1, 1, 1, 1, 1, 1},
					ResSpineBlockManifests: []*model.SpineBlockManifest{
						{
							ID:                     11111111,
							SpineBlockManifestType: model.SpineBlockManifestType_Snapshot,
							ExpirationTimestamp:    int64(1000),
						},
						{
							ID:                     22222222,
							SpineBlockManifestType: model.SpineBlockManifestType_Snapshot,
							ExpirationTimestamp:    int64(1000),
						},
					},
				},
			},
			args: args{
				spineBlockManifest: &model.SpineBlockManifest{
					ID:                  11111111,
					ExpirationTimestamp: 15875392,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockSpineService{
				Chaintype:                 tt.fields.Chaintype,
				QueryExecutor:             tt.fields.QueryExecutor,
				BlockQuery:                tt.fields.BlockQuery,
				Signature:                 tt.fields.Signature,
				BlocksmithStrategy:        tt.fields.BlocksmithStrategy,
				Observer:                  tt.fields.Observer,
				Logger:                    tt.fields.Logger,
				SpinePublicKeyService:     tt.fields.SpinePublicKeyService,
				SpineBlockManifestService: tt.fields.SpineBlockManifestService,
			}
			if err := bs.ValidateSpineBlockManifest(tt.args.spineBlockManifest); (err != nil) != tt.wantErr {
				t.Errorf("BlockSpineService.ValidateSpineBlockManifest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
