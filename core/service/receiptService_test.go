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
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	log "github.com/sirupsen/logrus"
	"reflect"
	"regexp"
	"testing"

	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/signaturetype"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/util"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"golang.org/x/crypto/sha3"
)

type (
	mockQueryExecutorFailExecuteSelect struct {
		query.Executor
	}
	mockQueryExecutorFailExecuteSelectReceipt struct {
		query.Executor
	}
	mockQueryExecutorSuccessMerkle struct {
		query.Executor
	}
	mockMerkleTreeQueryFailBuildTree struct {
		query.MerkleTreeQuery
	}
	mockQueryExecutorSuccessOneLinkedReceipts struct {
		query.Executor
	}
	mockQueryExecutorSuccessOneLinkedReceiptsAndMore struct {
		query.Executor
	}

	mockScrambleNodeServiceGetPriorityPeersSuccess struct {
		ScrambleNodeService
	}
	mockQueryExecutorSuccessSelectUnlinked struct {
		query.Executor
	}
)

var (
	mockBlockDataSelectReceipt = model.Block{
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
	mockLinkedReceipt = &model.BatchReceipt{
		Receipt: &model.Receipt{
			SenderPublicKey: []byte{
				8, 8, 8, 8, 7, 7, 7, 7, 6, 6, 6, 6, 5, 5, 5, 5, 4, 4, 4, 4, 3, 3, 3, 3, 2, 2, 2, 2, 1, 1, 1, 1,
			},
			RecipientPublicKey: []byte{
				1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5, 6, 6, 6, 6, 7, 7, 7, 7, 8, 8, 8, 8,
			},
			DatumType:            1,
			DatumHash:            make([]byte, 32),
			ReferenceBlockHeight: 10,
			ReferenceBlockHash:   make([]byte, 32),
			RMRLinked:            make([]byte, 32),
			RecipientSignature:   make([]byte, 64),
		},
		RMR:      make([]byte, 64),
		RMRIndex: 0,
	}
	mockUnlinkedReceiptWithLinkedRMR = &model.BatchReceipt{
		Receipt: &model.Receipt{
			SenderPublicKey: make([]byte, 32),
			RecipientPublicKey: []byte{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2,
			},
			DatumType:            1,
			DatumHash:            make([]byte, 32),
			ReferenceBlockHeight: 10,
			ReferenceBlockHash:   make([]byte, 32),
			RMRLinked:            make([]byte, 32),
			RecipientSignature:   make([]byte, 64),
		},
		RMR:      make([]byte, 64),
		RMRIndex: 0,
	}
	mockUnlinkedReceipt = &model.BatchReceipt{
		Receipt: &model.Receipt{
			SenderPublicKey: make([]byte, 32),
			RecipientPublicKey: []byte{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 2, 1,
			},
			DatumType:            1,
			DatumHash:            make([]byte, 32),
			ReferenceBlockHeight: 10,
			ReferenceBlockHash:   make([]byte, 32),
			RMRLinked:            make([]byte, 32),
			RecipientSignature:   make([]byte, 64),
		},
		RMR:      make([]byte, 64),
		RMRIndex: 0,
	}

	mockMerkle                  *util.MerkleRoot
	mockReceiptRMR              *bytes.Buffer
	mockMerkleHashes            []*bytes.Buffer
	mockFlattenTree             []byte
	mockFlattenIntermediateHash []byte

	// node registry
	mockNodeRegistrationData = model.NodeRegistration{
		NodeID:             111,
		NodePublicKey:      mockLinkedReceipt.Receipt.SenderPublicKey,
		AccountAddress:     nil,
		RegistrationHeight: 0,
		LockedBalance:      0,
		RegistrationStatus: 0,
		Latest:             false,
		Height:             0,
		NodeAddressInfo: &model.NodeAddressInfo{
			Address: "127.0.0.1",
			Port:    8001,
			Status:  model.NodeAddressStatus_NodeAddressConfirmed,
		},
	}
	mockNodeRegistrationDataB = model.NodeRegistration{
		NodeID:             222,
		NodePublicKey:      mockLinkedReceipt.Receipt.RecipientPublicKey,
		AccountAddress:     nil,
		RegistrationHeight: 0,
		LockedBalance:      0,
		RegistrationStatus: 0,
		Latest:             false,
		Height:             0,
		NodeAddressInfo: &model.NodeAddressInfo{
			Address: "127.0.0.1",
			Port:    8002,
			Status:  model.NodeAddressStatus_NodeAddressConfirmed,
		},
	}
	indexScramble = []int{
		0: 0,
		1: 1,
	}
	mockGoodScrambledNodes = &model.ScrambledNodes{
		AddressNodes: []*model.Peer{
			0: {
				Info: &model.Node{
					ID:      int64(111),
					Address: "127.0.0.1",
					Port:    8000,
				},
			},
			1: {
				Info: &model.Node{
					ID:      int64(222),
					Address: "127.0.0.1",
					Port:    3001,
				},
			},
		},
		IndexNodes: map[string]*int{
			"111": &indexScramble[0],
			"222": &indexScramble[1],
		},
	}
	mockScrambledNodesWithNodePublicKeyToIDMap = &model.ScrambledNodes{
		AddressNodes: []*model.Peer{
			0: {
				Info: &model.Node{
					ID:        int64(111),
					Address:   "127.0.0.1",
					Port:      8000,
					PublicKey: bcsNodePubKey1,
				},
			},
			1: {
				Info: &model.Node{
					ID:        int64(222),
					Address:   "127.0.0.1",
					Port:      3001,
					PublicKey: bcsNodePubKey2,
				},
			},
		},
		IndexNodes: map[string]*int{
			"111": &indexScramble[0],
			"222": &indexScramble[1],
		},
		NodePublicKeyToIDMap: map[string]int64{
			"dd45ce5b27aaec0877a2bb19cd5162aeab4672b4a9d08a597040026279590b3a": int64(111),
		},
	}
)

func (*mockScrambleNodeServiceGetPriorityPeersSuccess) GetScrambleNodesByHeight(
	blockHeight uint32,
) (*model.ScrambledNodes, error) {
	return mockGoodScrambledNodes, nil
}

func fixtureGenerateMerkle() {
	mockSeed := "mock seed"
	signature := crypto.NewSignature()
	receiptUtil := &coreUtil.ReceiptUtil{}
	// sign mock linked receipt and update the recipient public key
	mockLinkedReceipt.Receipt.RecipientPublicKey = signaturetype.NewEd25519Signature().GetPublicKeyFromSeed(mockSeed)
	mockSelectReceiptGoodScrambleNode.NodePublicKeyToIDMap[hex.EncodeToString(mockLinkedReceipt.Receipt.RecipientPublicKey)] =
		222
	unsignedReceiptByte := receiptUtil.GetUnsignedReceiptBytes(mockLinkedReceipt.Receipt)
	mockLinkedReceipt.Receipt.RecipientSignature = signature.SignByNode(unsignedReceiptByte, mockSeed)
	// sign rmr linked receipt
	mockUnlinkedReceiptWithLinkedRMR.Receipt.RecipientPublicKey = signaturetype.NewEd25519Signature().GetPublicKeyFromSeed(mockSeed)
	mockUnlinkedReceiptWithLinkedRMR.Receipt.SenderPublicKey = mockLinkedReceipt.Receipt.SenderPublicKey
	unsignedUnlinkedReceiptByte := receiptUtil.GetUnsignedReceiptBytes(mockUnlinkedReceiptWithLinkedRMR.Receipt)
	mockUnlinkedReceiptWithLinkedRMR.Receipt.RecipientSignature = signature.SignByNode(
		unsignedUnlinkedReceiptByte, mockSeed)
	// sign no rmr linked
	mockUnlinkedReceipt.Receipt.RecipientPublicKey = signaturetype.NewEd25519Signature().GetPublicKeyFromSeed(mockSeed)
	mockUnlinkedReceipt.Receipt.SenderPublicKey = mockLinkedReceipt.Receipt.SenderPublicKey
	unsignedNoRMRReceiptByte := receiptUtil.GetUnsignedReceiptBytes(mockUnlinkedReceipt.Receipt)
	mockUnlinkedReceipt.Receipt.RecipientSignature = signature.SignByNode(
		unsignedNoRMRReceiptByte, mockSeed,
	)
	mockNodeRegistrationDataB.NodePublicKey = mockLinkedReceipt.Receipt.RecipientPublicKey
	mockMerkle = &util.MerkleRoot{}
	receiptBytes := receiptUtil.GetSignedReceiptBytes(mockLinkedReceipt.Receipt)
	receiptHash := sha3.Sum256(receiptBytes)
	mockMerkleHashes = append(mockMerkleHashes, bytes.NewBuffer(receiptHash[:]))
	// generate random data for the hashes
	for i := 1; i < 8; i++ {
		var randomHash = make([]byte, 32)
		_, _ = rand.Read(randomHash)
		mockMerkleHashes = append(mockMerkleHashes, bytes.NewBuffer(randomHash))
	}
	// calculate the tree and root
	mockReceiptRMR, _ = mockMerkle.GenerateMerkleRoot(mockMerkleHashes)
	_, mockFlattenTree = mockMerkle.ToBytes()
	intermediateHashBuffer := mockMerkle.GetIntermediateHashes(bytes.NewBuffer(receiptHash[:]), 0)
	var intermediateHashes [][]byte
	for _, ihb := range intermediateHashBuffer {
		intermediateHashes = append(intermediateHashes, ihb.Bytes())
	}

	mockFlattenIntermediateHash = mockMerkle.FlattenIntermediateHashes(intermediateHashes)
}

func (*mockQueryExecutorFailExecuteSelect) ExecuteSelect(
	query string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	return nil, errors.New("mockError")
}

var (
	mockReceiptQuery = query.NewBatchReceiptQuery()
	mockBatchReceipt = &model.BatchReceipt{
		Receipt: &model.Receipt{
			SenderPublicKey:      []byte("BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"),
			RecipientPublicKey:   bcsNodePubKey2,
			DatumType:            uint32(1),
			DatumHash:            []byte{1, 2, 3, 4, 5, 6},
			ReferenceBlockHeight: uint32(1),
			ReferenceBlockHash:   []byte{1, 2, 3, 4, 5, 6},
			RMRLinked:            []byte{1, 2, 3, 4, 5, 6},
			RecipientSignature:   []byte{1, 2, 3, 4, 5, 6},
		},
		RMR:      []byte{1, 2, 3, 4, 5, 6},
		RMRIndex: uint32(4),
	}
)

func (*mockQueryExecutorSuccessSelectUnlinked) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, block_id, block_height, sender_account_address, recipient_account_address, transaction_type, fee, timestamp, " +
		"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version, transaction_index, multisig_child, " +
		"message FROM \"transaction\" WHERE block_id = ? AND multisig_child = false ORDER BY transaction_index ASC":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
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
			mockTransaction.MultisigChild,
			mockTransaction.Message,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, reference_block_hash, rmr_linked," +
		" recipient_signature, rmr, rmr_index FROM node_receipt AS rc WHERE rc.rmr = ? AND rc.datum_hash = ? AND rc." +
		"datum_type = ? ORDER BY recipient_public_key, reference_block_height":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
			mockReceiptQuery.Fields,
		).AddRow(
			mockBatchReceipt.Receipt.SenderPublicKey,
			mockBatchReceipt.Receipt.RecipientPublicKey,
			mockBatchReceipt.Receipt.DatumType,
			mockBatchReceipt.Receipt.DatumHash,
			mockBatchReceipt.Receipt.ReferenceBlockHeight,
			mockBatchReceipt.Receipt.ReferenceBlockHash,
			mockBatchReceipt.Receipt.RMRLinked,
			mockBatchReceipt.Receipt.RecipientSignature,
			mockBatchReceipt.RMR,
			mockBatchReceipt.RMRIndex,
		))
	default:
		return nil, errors.New("QueryNotMocked")
	}

	rows, _ := db.Query(qe)
	return rows, nil
}

var (
	mockMerkleTreeQuery = query.NewMerkleTreeQuery()
	mockRoot            = make([]byte, 32)
	mockBlockHeight     = uint32(0)
	mockTree            = make([]byte, 14*32)
)

func (*mockQueryExecutorSuccessSelectUnlinked) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {

	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version, merkle_root, merkle_tree, " +
		"reference_block_height FROM main_block WHERE height = 0":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.MainChain{}).Fields,
		).AddRow(
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
		))
	case "SELECT id, block_height, tree, timestamp FROM merkle_tree WHERE block_height = 0":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
			mockMerkleTreeQuery.Fields,
		).AddRow(
			mockRoot,
			mockBlockHeight,
			mockTree,
			int64(0),
		))

	default:
		return nil, errors.New("QueryNotMocked")
	}
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockQueryExecutorFailExecuteSelectReceipt) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, block_height, tree, timestamp FROM merkle_tree AS mt WHERE EXISTS " +
		"(SELECT rmr_linked FROM published_receipt AS pr WHERE mt.id = pr.rmr_linked)" +
		" AND block_height BETWEEN 0 AND 1000 ORDER BY block_height ASC LIMIT 5":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "BlockHeight", "Tree", "Timestamp",
		}).AddRow(
			[]byte{},
			uint32(0),
			[]byte{},
			1*constant.OneZBC,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
		"reference_block_hash, rmr_linked, recipient_signature, rmr, rmr_index FROM node_receipt AS rc " +
		"WHERE rc.rmr = ? AND NOT EXISTS (SELECT datum_hash FROM published_receipt AS pr WHERE " +
		"pr.datum_hash = rc.datum_hash AND pr.recipient_public_key = rc.recipient_public_key) AND " +
		"reference_block_height BETWEEN 0 AND 1000 " +
		"GROUP BY recipient_public_key":
		return nil, errors.New("mockError")
	}

	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockQueryExecutorSuccessOneLinkedReceipts) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, block_height, tree, timestamp FROM merkle_tree AS mt WHERE EXISTS " +
		"(SELECT rmr_linked FROM published_receipt AS pr WHERE mt.id = pr.rmr_linked) AND block_height " +
		"BETWEEN 0 AND 1000 ORDER BY block_height ASC LIMIT 5":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "BlockHeight", "Tree", "Timestamp",
		}).AddRow(
			mockReceiptRMR.Bytes(),
			uint32(0),
			mockFlattenTree,
			1*constant.OneZBC,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
		"reference_block_hash, rmr_linked, recipient_signature, rmr, rmr_index FROM node_receipt AS rc " +
		"WHERE rc.rmr = ? AND NOT EXISTS (SELECT datum_hash FROM published_receipt AS pr WHERE " +
		"pr.datum_hash = rc.datum_hash AND pr.recipient_public_key = rc.recipient_public_key) AND " +
		"reference_block_height BETWEEN 0 AND 1000 " +
		"GROUP BY recipient_public_key":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"sender_public_key",
			"recipient_public_key",
			"datum_type",
			"datum_hash",
			"reference_block_height",
			"reference_block_hash",
			"rmr_linked",
			"recipient_signature",
			"rmr",
			"rmr_index",
		}).AddRow(
			mockLinkedReceipt.Receipt.SenderPublicKey,
			mockLinkedReceipt.Receipt.RecipientPublicKey,
			mockLinkedReceipt.Receipt.DatumType,
			mockLinkedReceipt.Receipt.DatumHash,
			mockLinkedReceipt.Receipt.ReferenceBlockHeight,
			mockLinkedReceipt.Receipt.ReferenceBlockHash,
			mockLinkedReceipt.Receipt.RMRLinked,
			mockLinkedReceipt.Receipt.RecipientSignature,
			mockReceiptRMR.Bytes(),
			0,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
		"reference_block_hash, rmr_linked, recipient_signature, rmr, rmr_index FROM node_receipt AS rc WHERE NOT EXISTS " +
		"(SELECT datum_hash FROM published_receipt AS pr WHERE pr.datum_hash == rc.datum_hash) AND reference_block_height " +
		"BETWEEN 0 AND 1000 GROUP BY recipient_public_key ORDER BY reference_block_height ASC LIMIT 5":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"sender_public_key",
			"recipient_public_key",
			"datum_type",
			"datum_hash",
			"reference_block_height",
			"reference_block_hash",
			"rmr_linked",
			"recipient_signature",
			"rmr",
			"rmr_index",
		}).AddRow(
			mockLinkedReceipt.Receipt.SenderPublicKey,
			mockLinkedReceipt.Receipt.RecipientPublicKey,
			mockLinkedReceipt.Receipt.DatumType,
			mockLinkedReceipt.Receipt.DatumHash,
			mockLinkedReceipt.Receipt.ReferenceBlockHeight,
			mockLinkedReceipt.Receipt.ReferenceBlockHash,
			mockLinkedReceipt.Receipt.RMRLinked,
			mockLinkedReceipt.Receipt.RecipientSignature,
			mockReceiptRMR.Bytes(),
			0,
		))

	}

	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockQueryExecutorSuccessOneLinkedReceipts) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {

	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, registration_status, latest, " +
		"height FROM node_registry WHERE node_public_key = ? AND height <= ? ORDER BY height DESC LIMIT 1":
		nodePublicKey := args[0].([]byte)
		if !reflect.DeepEqual(nodePublicKey, mockNodeRegistrationData.NodePublicKey) {
			mock.ExpectQuery(regexp.QuoteMeta(qe)).
				WillReturnRows(sqlmock.NewRows(
					query.NewNodeRegistrationQuery().Fields,
				).AddRow(
					mockNodeRegistrationData.NodeID,
					mockNodeRegistrationData.NodePublicKey,
					mockNodeRegistrationData.AccountAddress,
					mockNodeRegistrationData.RegistrationHeight,
					mockNodeRegistrationData.LockedBalance,
					mockNodeRegistrationData.RegistrationStatus,
					mockNodeRegistrationData.Latest,
					mockNodeRegistrationData.Height,
				))
		} else {
			mock.ExpectQuery(regexp.QuoteMeta(qe)).
				WillReturnRows(sqlmock.NewRows(
					query.NewNodeRegistrationQuery().Fields,
				).AddRow(
					mockNodeRegistrationDataB.NodeID,
					mockNodeRegistrationDataB.NodePublicKey,
					mockNodeRegistrationDataB.AccountAddress,
					mockNodeRegistrationDataB.RegistrationHeight,
					mockNodeRegistrationDataB.LockedBalance,
					mockNodeRegistrationDataB.RegistrationStatus,
					mockNodeRegistrationDataB.Latest,
					mockNodeRegistrationDataB.Height,
				))
		}

	default:
		mock.ExpectQuery(regexp.QuoteMeta(qe)).
			WillReturnRows(sqlmock.NewRows(
				query.NewBlockQuery(&chaintype.MainChain{}).Fields,
			).AddRow(
				mockBlockDataSelectReceipt.GetHeight(),
				mockBlockDataSelectReceipt.GetID(),
				mockBlockDataSelectReceipt.GetBlockHash(),
				mockBlockDataSelectReceipt.GetPreviousBlockHash(),
				mockBlockDataSelectReceipt.GetTimestamp(),
				mockBlockDataSelectReceipt.GetBlockSeed(),
				mockBlockDataSelectReceipt.GetBlockSignature(),
				mockBlockDataSelectReceipt.GetCumulativeDifficulty(),
				mockBlockDataSelectReceipt.GetPayloadLength(),
				mockBlockDataSelectReceipt.GetPayloadHash(),
				mockBlockDataSelectReceipt.GetBlocksmithPublicKey(),
				mockBlockDataSelectReceipt.GetTotalAmount(),
				mockBlockDataSelectReceipt.GetTotalFee(),
				mockBlockDataSelectReceipt.GetTotalCoinBase(),
				mockBlockDataSelectReceipt.GetVersion(),
				mockBlockDataSelectReceipt.GetMerkleRoot(),
				mockBlockDataSelectReceipt.GetMerkleTree(),
				mockBlockDataSelectReceipt.GetReferenceBlockHeight(),
			))
	}
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockQueryExecutorSuccessOneLinkedReceiptsAndMore) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, registration_status, latest, " +
		"height FROM node_registry WHERE node_public_key = ? AND height <= ? ORDER BY height DESC LIMIT 1":
		nodePublicKey := args[0].([]byte)
		if !reflect.DeepEqual(nodePublicKey, mockNodeRegistrationData.NodePublicKey) {
			mock.ExpectQuery(regexp.QuoteMeta(qe)).
				WillReturnRows(sqlmock.NewRows(
					query.NewNodeRegistrationQuery().Fields,
				).AddRow(
					mockNodeRegistrationData.NodeID,
					mockNodeRegistrationData.NodePublicKey,
					mockNodeRegistrationData.AccountAddress,
					mockNodeRegistrationData.RegistrationHeight,
					mockNodeRegistrationData.LockedBalance,
					mockNodeRegistrationData.RegistrationStatus,
					mockNodeRegistrationData.Latest,
					mockNodeRegistrationData.Height,
				))
		} else {
			mock.ExpectQuery(regexp.QuoteMeta(qe)).
				WillReturnRows(sqlmock.NewRows(
					query.NewNodeRegistrationQuery().Fields,
				).AddRow(
					mockNodeRegistrationDataB.NodeID,
					mockNodeRegistrationDataB.NodePublicKey,
					mockNodeRegistrationDataB.AccountAddress,
					mockNodeRegistrationDataB.RegistrationHeight,
					mockNodeRegistrationDataB.LockedBalance,
					mockNodeRegistrationDataB.RegistrationStatus,
					mockNodeRegistrationDataB.Latest,
					mockNodeRegistrationDataB.Height,
				))
		}

	default:
		mock.ExpectQuery(regexp.QuoteMeta(qe)).
			WillReturnRows(sqlmock.NewRows(
				query.NewBlockQuery(&chaintype.MainChain{}).Fields,
			).AddRow(
				mockBlockDataSelectReceipt.GetHeight(),
				mockBlockDataSelectReceipt.GetID(),
				mockBlockDataSelectReceipt.GetBlockHash(),
				mockBlockDataSelectReceipt.GetPreviousBlockHash(),
				mockBlockDataSelectReceipt.GetTimestamp(),
				mockBlockDataSelectReceipt.GetBlockSeed(),
				mockBlockDataSelectReceipt.GetBlockSignature(),
				mockBlockDataSelectReceipt.GetCumulativeDifficulty(),
				mockBlockDataSelectReceipt.GetPayloadLength(),
				mockBlockDataSelectReceipt.GetPayloadHash(),
				mockBlockDataSelectReceipt.GetBlocksmithPublicKey(),
				mockBlockDataSelectReceipt.GetTotalAmount(),
				mockBlockDataSelectReceipt.GetTotalFee(),
				mockBlockDataSelectReceipt.GetTotalCoinBase(),
				mockBlockDataSelectReceipt.GetVersion(),
				mockBlockDataSelectReceipt.GetMerkleRoot(),
				mockBlockDataSelectReceipt.GetMerkleTree(),
				mockBlockDataSelectReceipt.GetReferenceBlockHeight(),
			))
	}
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockQueryExecutorSuccessOneLinkedReceiptsAndMore) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, block_height, tree, timestamp FROM merkle_tree AS mt WHERE EXISTS " +
		"(SELECT rmr_linked FROM published_receipt AS pr WHERE mt.id = pr.rmr_linked) " +
		"AND block_height BETWEEN 0 AND 1000 ORDER BY block_height ASC LIMIT 15":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "BlockHeight", "Tree", "Timestamp",
		}).AddRow(
			mockReceiptRMR.Bytes(),
			uint32(0),
			mockFlattenTree,
			1*constant.OneZBC,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
		"reference_block_hash, rmr_linked, recipient_signature, rmr, rmr_index FROM node_receipt AS rc " +
		"WHERE rc.rmr = ? AND NOT EXISTS (SELECT datum_hash FROM published_receipt AS pr WHERE " +
		"pr.datum_hash = rc.datum_hash AND pr.recipient_public_key = rc.recipient_public_key) AND " +
		"reference_block_height BETWEEN 0 AND 1000 " +
		"GROUP BY recipient_public_key":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"sender_public_key",
			"recipient_public_key",
			"datum_type",
			"datum_hash",
			"reference_block_height",
			"reference_block_hash",
			"rmr_linked",
			"recipient_signature",
			"rmr",
			"rmr_index",
		}).AddRow(
			mockLinkedReceipt.Receipt.SenderPublicKey,
			mockLinkedReceipt.Receipt.RecipientPublicKey,
			mockLinkedReceipt.Receipt.DatumType,
			mockLinkedReceipt.Receipt.DatumHash,
			mockLinkedReceipt.Receipt.ReferenceBlockHeight,
			mockLinkedReceipt.Receipt.ReferenceBlockHash,
			mockLinkedReceipt.Receipt.RMRLinked,
			mockLinkedReceipt.Receipt.RecipientSignature,
			mockReceiptRMR.Bytes(),
			0,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
		"reference_block_hash, rmr_linked, recipient_signature, rmr, rmr_index FROM node_receipt AS rc WHERE NOT " +
		"EXISTS (SELECT datum_hash FROM published_receipt AS pr WHERE pr.datum_hash == rc.datum_hash) AND " +
		"reference_block_height BETWEEN 0 AND 1000 GROUP BY recipient_public_key ORDER BY reference_block_height " +
		"ASC LIMIT 15":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"sender_public_key",
			"recipient_public_key",
			"datum_type",
			"datum_hash",
			"reference_block_height",
			"reference_block_hash",
			"rmr_linked",
			"recipient_signature",
			"rmr",
			"rmr_index",
		}).AddRow(
			mockUnlinkedReceiptWithLinkedRMR.Receipt.SenderPublicKey,
			mockUnlinkedReceiptWithLinkedRMR.Receipt.RecipientPublicKey,
			mockUnlinkedReceiptWithLinkedRMR.Receipt.DatumType,
			mockUnlinkedReceiptWithLinkedRMR.Receipt.DatumHash,
			mockUnlinkedReceiptWithLinkedRMR.Receipt.ReferenceBlockHeight,
			mockUnlinkedReceiptWithLinkedRMR.Receipt.ReferenceBlockHash,
			mockUnlinkedReceiptWithLinkedRMR.Receipt.RMRLinked,
			mockUnlinkedReceiptWithLinkedRMR.Receipt.RecipientSignature,
			make([]byte, 32),
			0,
		).AddRow(
			mockUnlinkedReceipt.Receipt.SenderPublicKey,
			mockUnlinkedReceipt.Receipt.RecipientPublicKey,
			mockUnlinkedReceipt.Receipt.DatumType,
			mockUnlinkedReceipt.Receipt.DatumHash,
			mockUnlinkedReceipt.Receipt.ReferenceBlockHeight,
			mockUnlinkedReceipt.Receipt.ReferenceBlockHash,
			mockUnlinkedReceipt.Receipt.RMRLinked,
			mockUnlinkedReceipt.Receipt.RecipientSignature,
			make([]byte, 32),
			0,
		))
	}

	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockQueryExecutorSuccessMerkle) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
		"ID", "Tree", "Timestamp",
	}).AddRow(
		[]byte{},
		[]byte{},
		1*constant.OneZBC,
	))
	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockMerkleTreeQueryFailBuildTree) BuildTree(rows *sql.Rows) (map[string][]byte, error) {
	return nil, errors.New("mockError")
}

type (
	mockNodeRegistrationSelectReceiptSuccess struct {
		NodeRegistrationService
	}
)

var (
	indexA                            = 1
	indexB                            = 2
	indexC                            = 3
	indexD                            = 4
	indexE                            = 5
	mockSelectReceiptGoodScrambleNode = &model.ScrambledNodes{
		AddressNodes: []*model.Peer{
			{
				Info: &model.Node{
					ID:      111,
					Address: "0.0.0.0",
					Port:    8001,
				},
			},
			{
				Info: &model.Node{
					ID:      222,
					Address: "0.0.0.0",
					Port:    8002,
				},
			},
			{
				Info: &model.Node{
					ID:      333,
					Address: "0.0.0.0",
					Port:    8003,
				},
			},
			{
				Info: &model.Node{
					ID:      444,
					Address: "0.0.0.0",
					Port:    8004,
				},
			},
			{
				Info: &model.Node{
					ID:      555,
					Address: "0.0.0.0",
					Port:    8005,
				},
			},
		},
		IndexNodes: map[string]*int{
			"111": &indexA,
			"222": &indexB,
			"333": &indexC,
			"444": &indexD,
			"555": &indexE,
		},
		NodePublicKeyToIDMap: map[string]int64{
			hex.EncodeToString(mockLinkedReceipt.Receipt.SenderPublicKey):    111,
			hex.EncodeToString(mockLinkedReceipt.Receipt.RecipientPublicKey): 222,
			"333": 333,
			"444": 444,
			"555": 555,
		},
		BlockHeight: 10,
	}
)

func (*mockNodeRegistrationSelectReceiptSuccess) GetScrambleNodesByHeight(
	blockHeight uint32,
) (*model.ScrambledNodes, error) {
	return mockSelectReceiptGoodScrambleNode, nil
}

type (
	mockScrambleNodeServiceSelectReceiptsSuccess struct {
		ScrambleNodeService
	}
	mockSelectReceiptsMainBlocksStorageSuccess struct {
		storage.CacheStackStorageInterface
	}
)

func (*mockScrambleNodeServiceSelectReceiptsSuccess) GetScrambleNodesByHeight(
	blockHeight uint32,
) (*model.ScrambledNodes, error) {
	return mockSelectReceiptGoodScrambleNode, nil
}

func (*mockSelectReceiptsMainBlocksStorageSuccess) GetAtIndex(height uint32, item interface{}) error {
	blockCacheObjCopy, ok := item.(*storage.BlockCacheObject)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "mockedErr")
	}
	blockCacheObjCopy.BlockHash = mockBlockDataSelectReceipt.BlockHash
	blockCacheObjCopy.Height = mockBlockDataSelectReceipt.Height
	blockCacheObjCopy.ID = mockBlockDataSelectReceipt.ID
	return nil
}

type (
	mockQueryExecutorGenerateReceiptsMerkleRootSuccess struct {
		query.Executor
	}
	mockQueryExecutorGenerateReceiptsMerkleRootSelectRowFail struct {
		query.Executor
	}
	mockQueryExecutorGenerateReceiptsMerkleRootSelectFail struct {
		query.Executor
	}
	mockGenerateReceiptsMerkleRootMainBlockStateStorageSuccess struct {
		storage.CacheStorageInterface
	}
	mockGenerateReceiptsMerkleRootMainBlockStateStorageFail struct {
		storage.CacheStorageInterface
	}
)

func (*mockGenerateReceiptsMerkleRootMainBlockStateStorageSuccess) GetItem(lastChange, item interface{}) error {
	var blockCopy, _ = item.(*model.Block)
	*blockCopy = mockBlockData
	return nil
}

func (*mockGenerateReceiptsMerkleRootMainBlockStateStorageFail) GetItem(lastChange, item interface{}) error {
	return errors.New("mockedError")
}

func (*mockQueryExecutorGenerateReceiptsMerkleRootSuccess) ExecuteSelectRow(
	qStr string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	switch qStr {
	case "SELECT MAX(height), id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version, merkle_root, merkle_tree, reference_block_height FROM main_block":
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
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version, merkle_root, merkle_tree, " +
		"reference_block_height FROM main_block WHERE height = 1":
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
	default:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).
			WillReturnRows(sqlmock.NewRows([]string{"total_record"}).AddRow(constant.ReceiptBatchMaximum))
	}

	return db.QueryRow(qStr), nil
}

func (*mockQueryExecutorGenerateReceiptsMerkleRootSuccess) BeginTx(bool, int) error {
	return nil
}
func (*mockQueryExecutorGenerateReceiptsMerkleRootSuccess) CommitTx(bool) error {
	return nil
}
func (*mockQueryExecutorGenerateReceiptsMerkleRootSuccess) RollbackTx(bool) error {
	return nil
}
func (*mockQueryExecutorGenerateReceiptsMerkleRootSuccess) ExecuteTransactions(
	queries [][]interface{},
) error {
	return nil
}

func (*mockQueryExecutorGenerateReceiptsMerkleRootSelectRowFail) ExecuteSelectRow(
	qStr string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, _, _ := sqlmock.New()
	return db.QueryRow(qStr), nil
}

func (*mockQueryExecutorGenerateReceiptsMerkleRootSelectFail) ExecuteSelectRow(
	qStr string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).
		WillReturnRows(sqlmock.NewRows([]string{"total_record"}).AddRow(constant.ReceiptBatchMaximum))
	return db.QueryRow(qStr), nil
}
func (*mockQueryExecutorGenerateReceiptsMerkleRootSelectFail) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	return nil, errors.New("mockError:executeSelectFail")
}
func (*mockQueryExecutorGenerateReceiptsMerkleRootSelectFail) BeginTx(bool, int) error {
	return errors.New("mockError:BeginTxFail")
}

func (*mockQueryExecutorGenerateReceiptsMerkleRootSelectFail) CommitTx(bool) error {
	return errors.New("mockError:CommitTxFail")
}
func (*mockQueryExecutorGenerateReceiptsMerkleRootSelectFail) RollbackTx(bool) error {
	return errors.New("mockError:RollbackTxFail")
}
func (*mockQueryExecutorGenerateReceiptsMerkleRootSelectFail) ExecuteTransactions(queries [][]interface{}) error {
	return errors.New("mockError:ExecuteTransactionsFail")
}

type (
	mockExecutorPruningNodeReceiptsSuccess struct {
		query.Executor
	}
)

func (*mockExecutorPruningNodeReceiptsSuccess) BeginTx(bool, int) error {
	return nil
}
func (*mockExecutorPruningNodeReceiptsSuccess) CommitTx(bool) error {
	return nil
}
func (*mockExecutorPruningNodeReceiptsSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}
func (*mockExecutorPruningNodeReceiptsSuccess) RollbackTx(bool) error {
	return nil
}
func (*mockExecutorPruningNodeReceiptsSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mockRow := mock.NewRows(query.NewBlockQuery(chaintype.GetChainType(0)).Fields)
	mockRow.AddRow(
		mockBlockDataSelectReceipt.GetHeight(),
		mockBlockDataSelectReceipt.GetID(),
		mockBlockDataSelectReceipt.GetBlockHash(),
		mockBlockDataSelectReceipt.GetPreviousBlockHash(),
		mockBlockDataSelectReceipt.GetTimestamp(),
		mockBlockDataSelectReceipt.GetBlockSeed(),
		mockBlockDataSelectReceipt.GetBlockSignature(),
		mockBlockDataSelectReceipt.GetCumulativeDifficulty(),
		mockBlockDataSelectReceipt.GetPayloadLength(),
		mockBlockDataSelectReceipt.GetPayloadHash(),
		mockBlockDataSelectReceipt.GetBlocksmithPublicKey(),
		mockBlockDataSelectReceipt.GetTotalAmount(),
		mockBlockDataSelectReceipt.GetTotalFee(),
		mockBlockDataSelectReceipt.GetTotalCoinBase(),
		mockBlockDataSelectReceipt.GetVersion(),
		mockBlockDataSelectReceipt.GetMerkleRoot(),
		mockBlockDataSelectReceipt.GetMerkleTree(),
		mockBlockDataSelectReceipt.GetReferenceBlockHeight(),
	)
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(mockRow)
	return db.QueryRow(qStr), nil
}

type (
	mockQueryExecutorGetPublishedReceiptsByHeight struct {
		query.Executor
	}
)

func (*mockQueryExecutorGetPublishedReceiptsByHeight) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mockRows := mock.NewRows(query.NewPublishedReceiptQuery().Fields)
	mockRows.AddRow(
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
	)
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(mockRows)
	return db.Query(qStr)
}
func TestReceiptService_GetPublishedReceiptsByHeight(t *testing.T) {
	type fields struct {
		NodeReceiptQuery        query.BatchReceiptQueryInterface
		MerkleTreeQuery         query.MerkleTreeQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		QueryExecutor           query.ExecutorInterface
		NodeRegistrationService NodeRegistrationServiceInterface
		Signature               crypto.SignatureInterface
		PublishedReceiptQuery   query.PublishedReceiptQueryInterface
	}
	type args struct {
		blockHeight uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.PublishedReceipt
		wantErr bool
	}{
		{
			name: "WantSuccess",
			fields: fields{
				PublishedReceiptQuery: query.NewPublishedReceiptQuery(),
				QueryExecutor:         &mockQueryExecutorGetPublishedReceiptsByHeight{},
			},
			args: args{
				blockHeight: 212,
			},
			want:    mockPublishedReceipt,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ReceiptService{
				NodeReceiptQuery:        tt.fields.NodeReceiptQuery,
				MerkleTreeQuery:         tt.fields.MerkleTreeQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeRegistrationService: tt.fields.NodeRegistrationService,
				Signature:               tt.fields.Signature,
				PublishedReceiptQuery:   tt.fields.PublishedReceiptQuery,
			}
			got, err := rs.GetPublishedReceiptsByHeight(tt.args.blockHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPublishedReceiptsByHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPublishedReceiptsByHeight() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockReceiptReminderStorageDuplicated struct {
		storage.ReceiptReminderStorage
	}
)

func (*mockReceiptReminderStorageDuplicated) GetItem(_, item interface{}) error {
	nItem, _ := item.(*chaintype.ChainType)
	*nItem = &chaintype.MainChain{}
	return nil
}

func TestReceiptService_IsDuplicated(t *testing.T) {
	type fields struct {
		NodeReceiptQuery        query.BatchReceiptQueryInterface
		MerkleTreeQuery         query.MerkleTreeQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		QueryExecutor           query.ExecutorInterface
		NodeRegistrationService NodeRegistrationServiceInterface
		Signature               crypto.SignatureInterface
		PublishedReceiptQuery   query.PublishedReceiptQueryInterface
		ReceiptUtil             coreUtil.ReceiptUtilInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ReceiptReminderStorage  storage.CacheStorageInterface
	}
	type args struct {
		publicKey []byte
		datumHash []byte
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantDuplicated bool
		wantErr        bool
	}{
		{
			name: "WantErr:InvalidKeyItem",
			fields: fields{
				ReceiptUtil:            &coreUtil.ReceiptUtil{},
				ReceiptReminderStorage: storage.NewReceiptReminderStorage(),
			},
			wantErr: true,
		},
		{
			name: "wantErr:Duplicated",
			fields: fields{
				ReceiptUtil:            &coreUtil.ReceiptUtil{},
				ReceiptReminderStorage: &mockReceiptReminderStorageDuplicated{},
			},
			args:           args{datumHash: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}, publicKey: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}},
			wantDuplicated: true,
			wantErr:        true,
		},
		{
			name: "want:Success",
			fields: fields{
				ReceiptUtil:            &coreUtil.ReceiptUtil{},
				ReceiptReminderStorage: storage.NewReceiptReminderStorage(),
			},
			args: args{datumHash: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}, publicKey: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ReceiptService{
				NodeReceiptQuery:        tt.fields.NodeReceiptQuery,
				MerkleTreeQuery:         tt.fields.MerkleTreeQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeRegistrationService: tt.fields.NodeRegistrationService,
				Signature:               tt.fields.Signature,
				PublishedReceiptQuery:   tt.fields.PublishedReceiptQuery,
				ReceiptUtil:             tt.fields.ReceiptUtil,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ReceiptReminderStorage:  tt.fields.ReceiptReminderStorage,
			}
			err := rs.CheckDuplication(tt.args.publicKey, tt.args.datumHash)
			if (err != nil) != tt.wantErr {
				b := err.(blocker.Blocker)
				if tt.wantDuplicated && b.Type != blocker.DuplicateReceiptErr {
					t.Errorf("CheckDuplication() gotDuplicated = %v, want %v", b.Type, tt.wantDuplicated)
					return
				}
				t.Errorf("CheckDuplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestReceiptService_GenerateReceiptsMerkleRoot(t *testing.T) {
	var (
		mockReceiptCacheStorage = storage.NewReceiptPoolCacheStorage()
		mockReceipt1            = model.Receipt{
			ReferenceBlockHeight: mockBlockData.Height,
			DatumHash:            make([]byte, 32),
			ReferenceBlockHash:   mockBlockData.BlockHash,
			RecipientPublicKey:   make([]byte, 32),
			RecipientSignature:   make([]byte, 64),
		}
		mockReceipt2 = model.Receipt{
			ReferenceBlockHeight: mockBlockData.Height - 1,
			DatumHash:            make([]byte, 32),
			ReferenceBlockHash:   make([]byte, 32),
			RecipientPublicKey:   make([]byte, 32),
			RecipientSignature:   make([]byte, 64),
		}
	)
	_ = mockReceiptCacheStorage.SetItem(nil, mockReceipt1)
	_ = mockReceiptCacheStorage.SetItem(nil, mockReceipt2)

	type fields struct {
		NodeReceiptQuery         query.BatchReceiptQueryInterface
		MerkleTreeQuery          query.MerkleTreeQueryInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		BlockQuery               query.BlockQueryInterface
		QueryExecutor            query.ExecutorInterface
		NodeRegistrationService  NodeRegistrationServiceInterface
		Signature                crypto.SignatureInterface
		PublishedReceiptQuery    query.PublishedReceiptQueryInterface
		ReceiptUtil              coreUtil.ReceiptUtilInterface
		MainBlockStateStorage    storage.CacheStorageInterface
		ScrambleNodeService      ScrambleNodeServiceInterface
		ReceiptReminderStorage   storage.CacheStorageInterface
		BatchReceiptCacheStorage storage.CacheStorageInterface
		MainBlocksStorage        storage.CacheStackStorageInterface
		LastMerkleRoot           []byte
		Logger                   *log.Logger
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
			name: "wantSuccess-{no-receipts}",
			fields: fields{
				BlockQuery:               query.NewBlockQuery(&chaintype.MainChain{}),
				NodeReceiptQuery:         query.NewBatchReceiptQuery(),
				MerkleTreeQuery:          query.NewMerkleTreeQuery(),
				QueryExecutor:            &mockQueryExecutorGenerateReceiptsMerkleRootSuccess{},
				MainBlockStateStorage:    &mockGenerateReceiptsMerkleRootMainBlockStateStorageSuccess{},
				BatchReceiptCacheStorage: storage.NewReceiptPoolCacheStorage(),
				Logger:                   log.New(),
			},
			args: args{
				block: &mockBlockData,
			},
			wantErr: false,
		},
		{
			name: "wantSuccess",
			fields: fields{
				ScrambleNodeService: &mockScrambleNodeServiceGetPriorityPeersSuccess{},
				ReceiptUtil: &mockReceiptUtil{
					validateSender: true,
					resSignetBytes: []byte{1, 1, 1, 1, 1},
				},
				Signature:                &mockSignature{},
				BlockQuery:               query.NewBlockQuery(&chaintype.MainChain{}),
				NodeReceiptQuery:         query.NewBatchReceiptQuery(),
				MerkleTreeQuery:          query.NewMerkleTreeQuery(),
				QueryExecutor:            &mockQueryExecutorGenerateReceiptsMerkleRootSuccess{},
				MainBlockStateStorage:    &mockGenerateReceiptsMerkleRootMainBlockStateStorageSuccess{},
				BatchReceiptCacheStorage: mockReceiptCacheStorage,
				Logger:                   log.New(),
			},
			args: args{
				block: &mockBlockData,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ReceiptService{
				NodeReceiptQuery:         tt.fields.NodeReceiptQuery,
				MerkleTreeQuery:          tt.fields.MerkleTreeQuery,
				NodeRegistrationQuery:    tt.fields.NodeRegistrationQuery,
				BlockQuery:               tt.fields.BlockQuery,
				QueryExecutor:            tt.fields.QueryExecutor,
				NodeRegistrationService:  tt.fields.NodeRegistrationService,
				Signature:                tt.fields.Signature,
				PublishedReceiptQuery:    tt.fields.PublishedReceiptQuery,
				ReceiptUtil:              tt.fields.ReceiptUtil,
				MainBlockStateStorage:    tt.fields.MainBlockStateStorage,
				ScrambleNodeService:      tt.fields.ScrambleNodeService,
				ReceiptReminderStorage:   tt.fields.ReceiptReminderStorage,
				BatchReceiptCacheStorage: tt.fields.BatchReceiptCacheStorage,
				MainBlocksStorage:        tt.fields.MainBlocksStorage,
				LastMerkleRoot:           tt.fields.LastMerkleRoot,
				Logger:                   tt.fields.Logger,
			}
			if err := rs.GenerateReceiptsMerkleRoot(tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("GenerateReceiptsMerkleRoot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	receiptSrvMockScrambleNodeService struct {
		ScrambleNodeService
	}
)

func (*receiptSrvMockScrambleNodeService) GetScrambleNodesByHeight(
	blockHeight uint32,
) (*model.ScrambledNodes, error) {
	return mockScrambledNodesWithNodePublicKeyToIDMap, nil
}

type (
	receiptSrvMockReceiptUtilSuccess struct {
		coreUtil.ReceiptUtilInterface
	}
)

func (*receiptSrvMockReceiptUtilSuccess) ValidateReceiptHelper(
	receipt *model.Receipt,
	executor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
	mainBlockStorage storage.CacheStackStorageInterface,
	signature crypto.SignatureInterface,
	scrambleNodesAtHeight *model.ScrambledNodes,
) error {
	return nil
}

func TestReceiptService_SelectUnlinkedReceipts(t *testing.T) {
	type fields struct {
		NodeReceiptQuery         query.BatchReceiptQueryInterface
		MerkleTreeQuery          query.MerkleTreeQueryInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		BlockQuery               query.BlockQueryInterface
		TransactionQuery         query.TransactionQueryInterface
		QueryExecutor            query.ExecutorInterface
		NodeRegistrationService  NodeRegistrationServiceInterface
		Signature                crypto.SignatureInterface
		PublishedReceiptQuery    query.PublishedReceiptQueryInterface
		ReceiptUtil              coreUtil.ReceiptUtilInterface
		MainBlockStateStorage    storage.CacheStorageInterface
		ScrambleNodeService      ScrambleNodeServiceInterface
		ReceiptReminderStorage   storage.CacheStorageInterface
		BatchReceiptCacheStorage storage.CacheStorageInterface
		MainBlocksStorage        storage.CacheStackStorageInterface
		LastMerkleRoot           []byte
		Logger                   *log.Logger
	}
	type args struct {
		numberOfReceipt   uint32
		blockHeight       uint32
		previousBlockHash []byte
		blockSeed         []byte
		secretPhrase      string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.BatchReceipt
		wantErr bool
	}{
		{
			name:   "SelectUnlinkedReceipts:success-{blockTooLow}",
			fields: fields{},
			args: args{
				blockHeight: constant.BatchReceiptLookBackHeight - 1,
			},
			want: make([]*model.BatchReceipt, 0),
		},
		{
			name:   "SelectUnlinkedReceipts:success-{noReceipts}",
			fields: fields{},
			args: args{
				numberOfReceipt: 0,
			},
			want: make([]*model.BatchReceipt, 0),
		},
		{
			name: "SelectUnlinkedReceipts:success",
			fields: fields{
				QueryExecutor:       &mockQueryExecutorSuccessSelectUnlinked{},
				BlockQuery:          query.NewBlockQuery(&chaintype.MainChain{}),
				NodeReceiptQuery:    query.NewBatchReceiptQuery(),
				TransactionQuery:    query.NewTransactionQuery(&chaintype.MainChain{}),
				MerkleTreeQuery:     query.NewMerkleTreeQuery(),
				ScrambleNodeService: &receiptSrvMockScrambleNodeService{},
				ReceiptUtil:         &receiptSrvMockReceiptUtilSuccess{},
			},
			args: args{
				numberOfReceipt: 2,
				blockHeight:     constant.BatchReceiptLookBackHeight,
				secretPhrase:    "test",
			},
			want: []*model.BatchReceipt{
				mockBatchReceipt,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ReceiptService{
				NodeReceiptQuery:         tt.fields.NodeReceiptQuery,
				MerkleTreeQuery:          tt.fields.MerkleTreeQuery,
				NodeRegistrationQuery:    tt.fields.NodeRegistrationQuery,
				BlockQuery:               tt.fields.BlockQuery,
				TransactionQuery:         tt.fields.TransactionQuery,
				QueryExecutor:            tt.fields.QueryExecutor,
				NodeRegistrationService:  tt.fields.NodeRegistrationService,
				Signature:                tt.fields.Signature,
				PublishedReceiptQuery:    tt.fields.PublishedReceiptQuery,
				ReceiptUtil:              tt.fields.ReceiptUtil,
				MainBlockStateStorage:    tt.fields.MainBlockStateStorage,
				ScrambleNodeService:      tt.fields.ScrambleNodeService,
				ReceiptReminderStorage:   tt.fields.ReceiptReminderStorage,
				BatchReceiptCacheStorage: tt.fields.BatchReceiptCacheStorage,
				MainBlocksStorage:        tt.fields.MainBlocksStorage,
				LastMerkleRoot:           tt.fields.LastMerkleRoot,
				Logger:                   tt.fields.Logger,
			}
			got, err := rs.SelectUnlinkedReceipts(tt.args.numberOfReceipt, tt.args.blockHeight, tt.args.previousBlockHash, tt.args.blockSeed, tt.args.secretPhrase)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectUnlinkedReceipts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SelectUnlinkedReceipts() got = %v, want %v", got, tt.want)
			}
		})
	}
}
