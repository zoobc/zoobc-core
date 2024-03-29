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
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/zoobc/zoobc-core/common/signaturetype"
	"reflect"
	"regexp"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/util"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
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
	mockQueryExecutorSuccessSelectLinked struct {
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
			RMR:                  make([]byte, 32),
			RecipientSignature:   make([]byte, 64),
		},
		RMRBatch:      make([]byte, 64),
		RMRBatchIndex: 0,
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
			RMR:                  make([]byte, 32),
			RecipientSignature:   make([]byte, 64),
		},
		RMRBatch:      make([]byte, 64),
		RMRBatchIndex: 0,
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
			RMR:                  make([]byte, 32),
			RecipientSignature:   make([]byte, 64),
		},
		RMRBatch:      make([]byte, 64),
		RMRBatchIndex: 0,
	}

	mockReceiptRMR  *bytes.Buffer
	mockFlattenTree []byte

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
	mockScrambledNodesWithNodePublicKeyToIDMap1 = &model.ScrambledNodes{
		AddressNodes: []*model.Peer{
			0: {
				Info: &model.Node{
					ID:        int64(111),
					Address:   "127.0.0.1",
					Port:      8000,
					PublicKey: signaturetype.NewEd25519Signature().GetPublicKeyFromSeed("test"),
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
			RMR:                  []byte{1, 2, 3, 4, 5, 6},
			RecipientSignature:   []byte{1, 2, 3, 4, 5, 6},
		},
		RMRBatch:      []byte{1, 2, 3, 4, 5, 6},
		RMRBatchIndex: uint32(4),
	}
	mockReceiptToPublish = &model.PublishedReceipt{
		Receipt:        mockBatchReceipt.Receipt,
		RMRLinked:      []byte{2, 3, 4, 5, 6, 7},
		RMRLinkedIndex: uint32(1),
	}
	mockRecSrvTransaction = &model.Transaction{
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
		TransactionBodyBytes:  mockSendZBCTxBodyBytes,
		Signature:             []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Version:               1,
		TransactionIndex:      1,
	}
	mockRecSrvBlock1 = &model.Block{
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
	mockRecSrvBlockData = model.Block{
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
	mockRecSrvPublishedReceipt = []*model.PublishedReceipt{
		{
			Receipt: &model.Receipt{
				SenderPublicKey:      make([]byte, 32),
				RecipientPublicKey:   make([]byte, 32),
				DatumType:            0,
				DatumHash:            make([]byte, 32),
				ReferenceBlockHeight: 0,
				ReferenceBlockHash:   make([]byte, 32),
				RMR:                  nil,
				RecipientSignature:   make([]byte, 64),
			},
			IntermediateHashes: nil,
			BlockHeight:        1,
			PublishedIndex:     0,
			RMRLinked:          make([]byte, 32),
			RMRLinkedIndex:     uint32(0),
		},
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
			mockRecSrvTransaction.ID,
			mockRecSrvTransaction.BlockID,
			mockRecSrvTransaction.Height,
			mockRecSrvTransaction.SenderAccountAddress,
			mockRecSrvTransaction.RecipientAccountAddress,
			mockRecSrvTransaction.TransactionType,
			mockRecSrvTransaction.Fee,
			mockRecSrvTransaction.Timestamp,
			mockRecSrvTransaction.TransactionHash,
			mockRecSrvTransaction.TransactionBodyLength,
			mockRecSrvTransaction.TransactionBodyBytes,
			mockRecSrvTransaction.Signature,
			mockRecSrvTransaction.Version,
			mockRecSrvTransaction.TransactionIndex,
			mockRecSrvTransaction.MultisigChild,
			mockRecSrvTransaction.Message,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, reference_block_hash, rmr, " +
		"recipient_signature, rmr_batch, rmr_batch_index FROM node_receipt AS rc WHERE rc.rmr_batch = ? AND rc.datum_hash = ? AND rc." +
		"datum_type = ? ORDER BY recipient_signature":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
			mockReceiptQuery.Fields,
		).AddRow(
			mockBatchReceipt.Receipt.SenderPublicKey,
			mockBatchReceipt.Receipt.RecipientPublicKey,
			mockBatchReceipt.Receipt.DatumType,
			mockBatchReceipt.Receipt.DatumHash,
			mockBatchReceipt.Receipt.ReferenceBlockHeight,
			mockBatchReceipt.Receipt.ReferenceBlockHash,
			mockBatchReceipt.Receipt.RMR,
			mockBatchReceipt.Receipt.RecipientSignature,
			mockBatchReceipt.RMRBatch,
			mockBatchReceipt.RMRBatchIndex,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, reference_block_hash, rmr, " +
		"recipient_signature, intermediate_hashes, block_height, rmr_linked, rmr_linked_index, " +
		"published_index FROM published_receipt WHERE block_height = ? ORDER BY published_index ASC":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
			query.NewPublishedReceiptQuery().Fields,
		).AddRow(
			mockRecSrvPublishedReceipt[0].Receipt.SenderPublicKey,
			mockRecSrvPublishedReceipt[0].Receipt.RecipientPublicKey,
			mockRecSrvPublishedReceipt[0].Receipt.DatumType,
			mockRecSrvPublishedReceipt[0].Receipt.DatumHash,
			mockRecSrvPublishedReceipt[0].Receipt.ReferenceBlockHeight,
			mockRecSrvPublishedReceipt[0].Receipt.ReferenceBlockHash,
			mockRecSrvPublishedReceipt[0].Receipt.RMR,
			mockRecSrvPublishedReceipt[0].Receipt.RecipientSignature,
			mockRecSrvPublishedReceipt[0].IntermediateHashes,
			mockRecSrvPublishedReceipt[0].BlockHeight,
			mockRecSrvPublishedReceipt[0].RMRLinked,
			mockRecSrvPublishedReceipt[0].RMRLinkedIndex,
			mockRecSrvPublishedReceipt[0].PublishedIndex,
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
			mockRecSrvBlock1.GetHeight(),
			mockRecSrvBlock1.GetID(),
			mockRecSrvBlock1.GetBlockHash(),
			mockRecSrvBlock1.GetPreviousBlockHash(),
			mockRecSrvBlock1.GetTimestamp(),
			mockRecSrvBlock1.GetBlockSeed(),
			mockRecSrvBlock1.GetBlockSignature(),
			mockRecSrvBlock1.GetCumulativeDifficulty(),
			mockRecSrvBlock1.GetPayloadLength(),
			mockRecSrvBlock1.GetPayloadHash(),
			mockRecSrvBlock1.GetBlocksmithPublicKey(),
			mockRecSrvBlock1.GetTotalAmount(),
			mockRecSrvBlock1.GetTotalFee(),
			mockRecSrvBlock1.GetTotalCoinBase(),
			mockRecSrvBlock1.GetVersion(),
			mockRecSrvBlock1.GetMerkleRoot(),
			mockRecSrvBlock1.GetMerkleTree(),
			mockRecSrvBlock1.GetReferenceBlockHeight(),
		))
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version, merkle_root, merkle_tree, " +
		"reference_block_height FROM main_block WHERE height = 960":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.MainChain{}).Fields,
		).AddRow(
			mockRecSrvBlock1.GetHeight(),
			mockRecSrvBlock1.GetID(),
			mockRecSrvBlock1.GetBlockHash(),
			mockRecSrvBlock1.GetPreviousBlockHash(),
			mockRecSrvBlock1.GetTimestamp(),
			mockRecSrvBlock1.GetBlockSeed(),
			mockRecSrvBlock1.GetBlockSignature(),
			mockRecSrvBlock1.GetCumulativeDifficulty(),
			mockRecSrvBlock1.GetPayloadLength(),
			mockRecSrvBlock1.GetPayloadHash(),
			mockRecSrvBlock1.GetBlocksmithPublicKey(),
			mockRecSrvBlock1.GetTotalAmount(),
			mockRecSrvBlock1.GetTotalFee(),
			mockRecSrvBlock1.GetTotalCoinBase(),
			mockRecSrvBlock1.GetVersion(),
			mockRecSrvBlock1.GetMerkleRoot(),
			mockRecSrvBlock1.GetMerkleTree(),
			mockRecSrvBlock1.GetReferenceBlockHeight(),
		))
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version, merkle_root, merkle_tree, " +
		"reference_block_height FROM main_block WHERE height = 60":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.MainChain{}).Fields,
		).AddRow(
			mockRecSrvBlock1.GetHeight(),
			mockRecSrvBlock1.GetID(),
			mockRecSrvBlock1.GetBlockHash(),
			mockRecSrvBlock1.GetPreviousBlockHash(),
			mockRecSrvBlock1.GetTimestamp(),
			mockRecSrvBlock1.GetBlockSeed(),
			mockRecSrvBlock1.GetBlockSignature(),
			mockRecSrvBlock1.GetCumulativeDifficulty(),
			mockRecSrvBlock1.GetPayloadLength(),
			mockRecSrvBlock1.GetPayloadHash(),
			mockRecSrvBlock1.GetBlocksmithPublicKey(),
			mockRecSrvBlock1.GetTotalAmount(),
			mockRecSrvBlock1.GetTotalFee(),
			mockRecSrvBlock1.GetTotalCoinBase(),
			mockRecSrvBlock1.GetVersion(),
			mockRecSrvBlock1.GetMerkleRoot(),
			mockRecSrvBlock1.GetMerkleTree(),
			mockRecSrvBlock1.GetReferenceBlockHeight(),
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
		return nil, errors.New(qe)
	}
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockQueryExecutorSuccessSelectLinked) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT height, blocksmith_public_key FROM main_block WHERE height >= 79 AND height <= 5 ORDER BY height DESC":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
			[]string{
				"height",
				"blocksmith_public_key",
			},
		).AddRow(
			79,
			signaturetype.NewEd25519Signature().GetPublicKeyFromSeed("test"),
		))
	case "SELECT id, block_id, block_height, sender_account_address, recipient_account_address, transaction_type, fee, timestamp, " +
		"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version, transaction_index, multisig_child, " +
		"message FROM \"transaction\" WHERE block_id = ? AND multisig_child = false ORDER BY transaction_index ASC":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
			query.NewTransactionQuery(&chaintype.MainChain{}).Fields,
		).AddRow(
			mockRecSrvTransaction.ID,
			mockRecSrvTransaction.BlockID,
			mockRecSrvTransaction.Height,
			mockRecSrvTransaction.SenderAccountAddress,
			mockRecSrvTransaction.RecipientAccountAddress,
			mockRecSrvTransaction.TransactionType,
			mockRecSrvTransaction.Fee,
			mockRecSrvTransaction.Timestamp,
			mockRecSrvTransaction.TransactionHash,
			mockRecSrvTransaction.TransactionBodyLength,
			mockRecSrvTransaction.TransactionBodyBytes,
			mockRecSrvTransaction.Signature,
			mockRecSrvTransaction.Version,
			mockRecSrvTransaction.TransactionIndex,
			mockRecSrvTransaction.MultisigChild,
			mockRecSrvTransaction.Message,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, reference_block_hash, rmr," +
		" recipient_signature, rmr_batch, rmr_batch_index FROM node_receipt AS rc WHERE rc.rmr = ? AND rc.datum_hash = ? AND rc." +
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
			mockBatchReceipt.Receipt.RMR,
			mockBatchReceipt.Receipt.RecipientSignature,
			mockBatchReceipt.RMRBatch,
			mockBatchReceipt.RMRBatchIndex,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, reference_block_hash, rmr, " +
		"recipient_signature, intermediate_hashes, block_height, rmr_linked, rmr_linked_index, " +
		"published_index FROM published_receipt WHERE block_height = ? ORDER BY published_index ASC":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
			query.NewPublishedReceiptQuery().Fields,
		).AddRow(
			mockRecSrvPublishedReceipt[0].Receipt.SenderPublicKey,
			mockRecSrvPublishedReceipt[0].Receipt.RecipientPublicKey,
			mockRecSrvPublishedReceipt[0].Receipt.DatumType,
			mockRecSrvPublishedReceipt[0].Receipt.DatumHash,
			mockRecSrvPublishedReceipt[0].Receipt.ReferenceBlockHeight,
			mockRecSrvPublishedReceipt[0].Receipt.ReferenceBlockHash,
			mockRecSrvPublishedReceipt[0].Receipt.RMR,
			mockRecSrvPublishedReceipt[0].Receipt.RecipientSignature,
			mockRecSrvPublishedReceipt[0].IntermediateHashes,
			mockRecSrvPublishedReceipt[0].BlockHeight,
			mockRecSrvPublishedReceipt[0].RMRLinked,
			mockRecSrvPublishedReceipt[0].RMRLinkedIndex,
			mockRecSrvPublishedReceipt[0].PublishedIndex,
		))
	default:
		return nil, errors.New("QueryNotMocked")
	}

	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockQueryExecutorSuccessSelectLinked) ExecuteSelectRow(
	qe string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, reference_block_hash, rmr," +
		" recipient_signature, rmr_batch, rmr_batch_index FROM node_receipt AS rc WHERE rc.datum_hash = ? AND rc.datum_type = ? AND rc." +
		"recipient_public_key = ? LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
			query.NewBatchReceiptQuery().Fields,
		).AddRow(
			mockBatchReceipt.Receipt.SenderPublicKey,
			mockBatchReceipt.Receipt.RecipientPublicKey,
			mockBatchReceipt.Receipt.DatumType,
			mockBatchReceipt.Receipt.DatumHash,
			mockBatchReceipt.Receipt.ReferenceBlockHeight,
			mockBatchReceipt.Receipt.ReferenceBlockHash,
			mockBatchReceipt.Receipt.RMR,
			mockBatchReceipt.Receipt.RecipientSignature,
			mockBatchReceipt.RMRBatch,
			mockBatchReceipt.RMRBatchIndex,
		))
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version, merkle_root, merkle_tree, " +
		"reference_block_height FROM main_block WHERE height = 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.MainChain{}).Fields,
		).AddRow(
			mockRecSrvBlockData.GetHeight(),
			mockRecSrvBlockData.GetID(),
			mockRecSrvBlockData.GetBlockHash(),
			mockRecSrvBlockData.GetPreviousBlockHash(),
			mockRecSrvBlockData.GetTimestamp(),
			mockRecSrvBlockData.GetBlockSeed(),
			mockRecSrvBlockData.GetBlockSignature(),
			mockRecSrvBlockData.GetCumulativeDifficulty(),
			mockRecSrvBlockData.GetPayloadLength(),
			mockRecSrvBlockData.GetPayloadHash(),
			mockRecSrvBlockData.GetBlocksmithPublicKey(),
			mockRecSrvBlockData.GetTotalAmount(),
			mockRecSrvBlockData.GetTotalFee(),
			mockRecSrvBlockData.GetTotalCoinBase(),
			mockRecSrvBlockData.GetVersion(),
			mockRecSrvBlockData.GetMerkleRoot(),
			mockRecSrvBlockData.GetMerkleTree(),
			mockRecSrvBlockData.GetReferenceBlockHeight(),
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, reference_block_hash, rmr," +
		" recipient_signature, intermediate_hashes, block_height, rmr_linked, rmr_linked_index, " +
		"published_index FROM published_receipt WHERE block_height = ? AND recipient_public_key = ? AND rmr_linked IS NULL LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
			query.NewPublishedReceiptQuery().Fields,
		).AddRow(
			mockRecSrvPublishedReceipt[0].Receipt.SenderPublicKey,
			mockRecSrvPublishedReceipt[0].Receipt.RecipientPublicKey,
			mockRecSrvPublishedReceipt[0].Receipt.DatumType,
			mockRecSrvPublishedReceipt[0].Receipt.DatumHash,
			mockRecSrvPublishedReceipt[0].Receipt.ReferenceBlockHeight,
			mockRecSrvPublishedReceipt[0].Receipt.ReferenceBlockHash,
			mockRecSrvPublishedReceipt[0].Receipt.RMR,
			mockRecSrvPublishedReceipt[0].Receipt.RecipientSignature,
			mockRecSrvPublishedReceipt[0].IntermediateHashes,
			mockRecSrvPublishedReceipt[0].BlockHeight,
			mockRecSrvPublishedReceipt[0].RMRLinked,
			mockRecSrvPublishedReceipt[0].RMRLinkedIndex,
			mockRecSrvPublishedReceipt[0].PublishedIndex,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, reference_block_hash, " +
		"rmr, recipient_signature, rmr_batch, rmr_batch_index FROM node_receipt AS rc WHERE rc.reference_block_height = ? AND rc." +
		"reference_block_hash = ? LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
			query.NewBatchReceiptQuery().Fields,
		).AddRow(
			mockBatchReceipt.Receipt.SenderPublicKey,
			mockBatchReceipt.Receipt.RecipientPublicKey,
			mockBatchReceipt.Receipt.DatumType,
			mockBatchReceipt.Receipt.DatumHash,
			mockBatchReceipt.Receipt.ReferenceBlockHeight,
			mockBatchReceipt.Receipt.ReferenceBlockHash,
			mockBatchReceipt.Receipt.RMR,
			mockBatchReceipt.Receipt.RecipientSignature,
			mockBatchReceipt.RMRBatch,
			mockBatchReceipt.RMRBatchIndex,
		))
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version, merkle_root, merkle_tree, " +
		"reference_block_height FROM main_block WHERE height = 0":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
			query.NewBlockQuery(&chaintype.MainChain{}).Fields,
		).AddRow(
			mockRecSrvBlock1.GetHeight(),
			mockRecSrvBlock1.GetID(),
			mockRecSrvBlock1.GetBlockHash(),
			mockRecSrvBlock1.GetPreviousBlockHash(),
			mockRecSrvBlock1.GetTimestamp(),
			mockRecSrvBlock1.GetBlockSeed(),
			mockRecSrvBlock1.GetBlockSignature(),
			mockRecSrvBlock1.GetCumulativeDifficulty(),
			mockRecSrvBlock1.GetPayloadLength(),
			mockRecSrvBlock1.GetPayloadHash(),
			mockRecSrvBlock1.GetBlocksmithPublicKey(),
			mockRecSrvBlock1.GetTotalAmount(),
			mockRecSrvBlock1.GetTotalFee(),
			mockRecSrvBlock1.GetTotalCoinBase(),
			mockRecSrvBlock1.GetVersion(),
			mockRecSrvBlock1.GetMerkleRoot(),
			mockRecSrvBlock1.GetMerkleTree(),
			mockRecSrvBlock1.GetReferenceBlockHeight(),
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
		"reference_block_hash, rmr, recipient_signature, rmr_batch, rmr_batch_index FROM node_receipt AS rc " +
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
		"reference_block_hash, rmr, recipient_signature, rmr_batch, rmr_batch_index FROM node_receipt AS rc " +
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
			mockLinkedReceipt.Receipt.RMR,
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
			"rmr",
			"recipient_signature",
			"rmr_batch",
			"rmr_batch_index",
		}).AddRow(
			mockLinkedReceipt.Receipt.SenderPublicKey,
			mockLinkedReceipt.Receipt.RecipientPublicKey,
			mockLinkedReceipt.Receipt.DatumType,
			mockLinkedReceipt.Receipt.DatumHash,
			mockLinkedReceipt.Receipt.ReferenceBlockHeight,
			mockLinkedReceipt.Receipt.ReferenceBlockHash,
			mockLinkedReceipt.Receipt.RMR,
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
		"reference_block_hash, rmr, recipient_signature, rmr_batch, rmr_batch_index FROM node_receipt AS rc " +
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
			"rmr",
			"recipient_signature",
			"rmr_batch",
			"rmr_batch_index",
		}).AddRow(
			mockLinkedReceipt.Receipt.SenderPublicKey,
			mockLinkedReceipt.Receipt.RecipientPublicKey,
			mockLinkedReceipt.Receipt.DatumType,
			mockLinkedReceipt.Receipt.DatumHash,
			mockLinkedReceipt.Receipt.ReferenceBlockHeight,
			mockLinkedReceipt.Receipt.ReferenceBlockHash,
			mockLinkedReceipt.Receipt.RMR,
			mockLinkedReceipt.Receipt.RecipientSignature,
			mockReceiptRMR.Bytes(),
			0,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
		"reference_block_hash, rmr, recipient_signature, rmr_batch, rmr_batch_index FROM node_receipt AS rc WHERE NOT " +
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
			mockUnlinkedReceiptWithLinkedRMR.Receipt.RMR,
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
			mockUnlinkedReceipt.Receipt.RMR,
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
	mockMerkleRootUtilsSuccess struct {
		util.MerkleRoot
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

func (*mockMerkleRootUtilsSuccess) GenerateMerkleRoot(items []*bytes.Buffer) (*bytes.Buffer, error) {
	return &bytes.Buffer{}, nil
}

func (*mockMerkleRootUtilsSuccess) ToBytes() (root, tree []byte) {
	return []byte{1, 2, 3, 4, 5}, []byte{2, 3, 4, 5, 6}
}

func (*mockGenerateReceiptsMerkleRootMainBlockStateStorageSuccess) GetItem(lastChange, item interface{}) error {
	var blockCopy, _ = item.(*model.Block)
	*blockCopy = mockRecSrvBlockData
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
				mockRecSrvBlockData.GetHeight(),
				mockRecSrvBlockData.GetID(),
				mockRecSrvBlockData.GetBlockHash(),
				mockRecSrvBlockData.GetPreviousBlockHash(),
				mockRecSrvBlockData.GetTimestamp(),
				mockRecSrvBlockData.GetBlockSeed(),
				mockRecSrvBlockData.GetBlockSignature(),
				mockRecSrvBlockData.GetCumulativeDifficulty(),
				mockRecSrvBlockData.GetPayloadLength(),
				mockRecSrvBlockData.GetPayloadHash(),
				mockRecSrvBlockData.GetBlocksmithPublicKey(),
				mockRecSrvBlockData.GetTotalAmount(),
				mockRecSrvBlockData.GetTotalFee(),
				mockRecSrvBlockData.GetTotalCoinBase(),
				mockRecSrvBlockData.GetVersion(),
				mockRecSrvBlockData.GetMerkleRoot(),
				mockRecSrvBlockData.GetMerkleTree(),
				mockRecSrvBlockData.GetReferenceBlockHeight(),
			))
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version, merkle_root, merkle_tree, " +
		"reference_block_height FROM main_block WHERE height = 1":
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).
			WillReturnRows(sqlmock.NewRows(
				query.NewBlockQuery(&chaintype.MainChain{}).Fields,
			).AddRow(
				mockRecSrvBlockData.GetHeight(),
				mockRecSrvBlockData.GetID(),
				mockRecSrvBlockData.GetBlockHash(),
				mockRecSrvBlockData.GetPreviousBlockHash(),
				mockRecSrvBlockData.GetTimestamp(),
				mockRecSrvBlockData.GetBlockSeed(),
				mockRecSrvBlockData.GetBlockSignature(),
				mockRecSrvBlockData.GetCumulativeDifficulty(),
				mockRecSrvBlockData.GetPayloadLength(),
				mockRecSrvBlockData.GetPayloadHash(),
				mockRecSrvBlockData.GetBlocksmithPublicKey(),
				mockRecSrvBlockData.GetTotalAmount(),
				mockRecSrvBlockData.GetTotalFee(),
				mockRecSrvBlockData.GetTotalCoinBase(),
				mockRecSrvBlockData.GetVersion(),
				mockRecSrvBlockData.GetMerkleRoot(),
				mockRecSrvBlockData.GetMerkleTree(),
				mockRecSrvBlockData.GetReferenceBlockHeight(),
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
		mockRecSrvPublishedReceipt[0].Receipt.SenderPublicKey,
		mockRecSrvPublishedReceipt[0].Receipt.RecipientPublicKey,
		mockRecSrvPublishedReceipt[0].Receipt.DatumType,
		mockRecSrvPublishedReceipt[0].Receipt.DatumHash,
		mockRecSrvPublishedReceipt[0].Receipt.ReferenceBlockHeight,
		mockRecSrvPublishedReceipt[0].Receipt.ReferenceBlockHash,
		mockRecSrvPublishedReceipt[0].Receipt.RMR,
		mockRecSrvPublishedReceipt[0].Receipt.RecipientSignature,
		mockRecSrvPublishedReceipt[0].IntermediateHashes,
		mockRecSrvPublishedReceipt[0].BlockHeight,
		mockRecSrvPublishedReceipt[0].RMRLinked,
		mockRecSrvPublishedReceipt[0].RMRLinkedIndex,
		mockRecSrvPublishedReceipt[0].PublishedIndex,
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
			want:    mockRecSrvPublishedReceipt,
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
			ReferenceBlockHeight: mockRecSrvBlockData.Height,
			DatumHash:            make([]byte, 32),
			ReferenceBlockHash:   mockRecSrvBlockData.BlockHash,
			RecipientPublicKey:   make([]byte, 32),
			RecipientSignature:   make([]byte, 64),
		}
		mockReceipt2 = model.Receipt{
			ReferenceBlockHeight: mockRecSrvBlockData.Height - 1,
			DatumHash:            make([]byte, 32),
			ReferenceBlockHash:   make([]byte, 32),
			RecipientPublicKey:   make([]byte, 32),
			RecipientSignature:   make([]byte, 64),
		}
	)
	_ = mockReceiptCacheStorage.SetItem("a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a", mockReceipt1)
	_ = mockReceiptCacheStorage.SetItem("01020304", mockReceipt2)

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
		MerkleRootUtil           util.MerkleRootInterface
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
				block: &mockRecSrvBlockData,
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
				MerkleRootUtil:           &mockMerkleRootUtilsSuccess{},
				Logger:                   log.New(),
			},
			args: args{
				block: &mockRecSrvBlockData,
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
				MerkleRootUtil:           tt.fields.MerkleRootUtil,
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
	if blockHeight == 960 ||
		blockHeight == 80 ||
		blockHeight == 39 ||
		blockHeight == 20 ||
		blockHeight == 0 ||
		blockHeight == 1 {
		return mockScrambledNodesWithNodePublicKeyToIDMap, nil
	}
	return &model.ScrambledNodes{
		AddressNodes: []*model.Peer{},
		IndexNodes:   map[string]*int{},
	}, nil
}

type (
	receiptSrvMockReceiptUtilSuccess struct {
		coreUtil.ReceiptUtilInterface
	}
	mockNodeConfigurationService struct {
		nodePubKey []byte
		NodeConfigurationService
	}
)

func (mockReceiptCfgSrv *mockNodeConfigurationService) GetNodePublicKey() []byte {
	if mockReceiptCfgSrv.nodePubKey != nil {
		return mockReceiptCfgSrv.nodePubKey
	}
	return signaturetype.NewEd25519Signature().GetPublicKeyFromSeed("test")

}
func (*receiptSrvMockReceiptUtilSuccess) BuildBlockDatumHashes(
	block *model.Block,
	executor query.ExecutorInterface,
	transactionQuery query.TransactionQueryInterface,
) ([][]byte, error) {
	return [][]byte{
		{1, 1, 1, 1, 1, 1, 1, 1},
		{2, 2, 2, 2, 2, 2, 2, 2},
	}, nil

}

func (*receiptSrvMockReceiptUtilSuccess) GetRandomDatumHash(hashList [][]byte, blockSeed []byte) (rndDatumHash []byte, rndDatumType uint32,
	err error) {
	return []byte{1, 1, 1, 1, 1, 1, 1, 1}, constant.ReceiptDatumTypeBlock, nil

}

func (*receiptSrvMockReceiptUtilSuccess) GetMaxLookBackHeight(firstLookBackBlockHeight uint32) (uint32, error) {
	return 5, nil

}

func (*receiptSrvMockReceiptUtilSuccess) ValidateReceiptHelper(
	receipt *model.Receipt,
	validateRefBlock bool,
	executor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
	mainBlockStorage storage.CacheStackStorageInterface,
	signature crypto.SignatureInterface,
	scrambleNodesAtHeight *model.ScrambledNodes,
) error {
	return nil
}

func (*receiptSrvMockReceiptUtilSuccess) GeneratePublishedReceipt(
	batchReceipt *model.Receipt,
	PublishedIndex uint32,
	RMRLinked []byte,
	RMRLinkedIndex uint32,
	executor query.ExecutorInterface,
	merkleTreeQuery query.MerkleTreeQueryInterface,
) (*model.PublishedReceipt, error) {
	return mockReceiptToPublish, nil
}

func (*receiptSrvMockReceiptUtilSuccess) GetPriorityPeersAtHeight(
	nodePubKey []byte,
	scrambleNodes *model.ScrambledNodes,
) (map[string]*model.Peer, error) {
	var (
		res = make(map[string]*model.Peer)
	)

	for _, peer := range mockScrambledNodesWithNodePublicKeyToIDMap1.AddressNodes {
		res[hex.EncodeToString(make([]byte, 32))] = peer
	}
	return res, nil
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
		NodeConfiguration        NodeConfigurationServiceInterface
		Logger                   *log.Logger
	}
	type args struct {
		numberOfReceipt uint32
		blockHeight     uint32
		blockSeed       []byte
		secretPhrase    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.PublishedReceipt
		wantErr bool
	}{
		{
			name:   "SelectUnlinkedReceipts:success-{blockTooLow}",
			fields: fields{},
			args: args{
				blockHeight: constant.BatchReceiptLookBackHeight - 1,
			},
			want: make([]*model.PublishedReceipt, 0),
		},
		{
			name:   "SelectUnlinkedReceipts:success-{noReceipts}",
			fields: fields{},
			args: args{
				numberOfReceipt: 0,
			},
			want: make([]*model.PublishedReceipt, 0),
		},
		{
			name: "SelectUnlinkedReceipts:success",
			fields: fields{
				QueryExecutor:         &mockQueryExecutorSuccessSelectUnlinked{},
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				NodeReceiptQuery:      query.NewBatchReceiptQuery(),
				TransactionQuery:      query.NewTransactionQuery(&chaintype.MainChain{}),
				PublishedReceiptQuery: query.NewPublishedReceiptQuery(),
				MerkleTreeQuery:       query.NewMerkleTreeQuery(),
				ScrambleNodeService:   &receiptSrvMockScrambleNodeService{},
				ReceiptUtil:           &receiptSrvMockReceiptUtilSuccess{},
				NodeConfiguration: &mockNodeConfigurationService{
					nodePubKey: make([]byte, 32),
				},
			},
			args: args{
				numberOfReceipt: 2,
				blockHeight:     constant.BatchReceiptLookBackHeight,
				secretPhrase:    "test",
			},
			want: []*model.PublishedReceipt{
				mockReceiptToPublish,
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
				NodeConfiguration:        tt.fields.NodeConfiguration,
				Logger:                   tt.fields.Logger,
			}
			got, err := rs.SelectUnlinkedReceipts(tt.args.numberOfReceipt, tt.args.blockHeight, tt.args.blockSeed)
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

// TODO: uncomment when fixing linked receipts logic
// func TestReceiptService_SelectLinkedReceipts(t *testing.T) {
// 	type fields struct {
// 		NodeReceiptQuery         query.BatchReceiptQueryInterface
// 		MerkleTreeQuery          query.MerkleTreeQueryInterface
// 		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
// 		BlockQuery               query.BlockQueryInterface
// 		TransactionQuery         query.TransactionQueryInterface
// 		QueryExecutor            query.ExecutorInterface
// 		NodeRegistrationService  NodeRegistrationServiceInterface
// 		Signature                crypto.SignatureInterface
// 		PublishedReceiptQuery    query.PublishedReceiptQueryInterface
// 		ReceiptUtil              coreUtil.ReceiptUtilInterface
// 		MainBlockStateStorage    storage.CacheStorageInterface
// 		ScrambleNodeService      ScrambleNodeServiceInterface
// 		ReceiptReminderStorage   storage.CacheStorageInterface
// 		BatchReceiptCacheStorage storage.CacheStorageInterface
// 		MainBlocksStorage        storage.CacheStackStorageInterface
// 		LastMerkleRoot           []byte
// 		MerkleRootUtil           util.MerkleRootInterface
// 		NodeConfiguration        NodeConfigurationServiceInterface
// 		Logger                   *log.Logger
// 	}
// 	type args struct {
// 		numberOfReceipt          uint32
// 		numberOfUnlinkedReceipts uint32
// 		blockHeight              uint32
// 		blockSeed                []byte
// 		secretPhrase             string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    []*model.PublishedReceipt
// 		wantErr bool
// 	}{
// 		{
// 			name: "SelectLinkedReceipts:success-{blockTooLow}",
// 			fields: fields{
// 				NodeConfiguration: &mockNodeConfigurationService{},
// 			},
// 			args: args{
// 				blockHeight: constant.BatchReceiptLookBackHeight - 1,
// 			},
// 			want: nil,
// 		},
// 		{
// 			name: "SelectLinkedReceipts:success-{noReceipts}",
// 			fields: fields{
// 				NodeConfiguration: &mockNodeConfigurationService{},
// 			},
// 			args: args{
// 				numberOfReceipt: 0,
// 			},
// 			want: nil,
// 		},
// 		{
// 			name: "SelectLinkedReceipts:success",
// 			fields: fields{
// 				QueryExecutor:         &mockQueryExecutorSuccessSelectLinked{},
// 				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
// 				NodeReceiptQuery:      query.NewBatchReceiptQuery(),
// 				TransactionQuery:      query.NewTransactionQuery(&chaintype.MainChain{}),
// 				MerkleTreeQuery:       query.NewMerkleTreeQuery(),
// 				PublishedReceiptQuery: query.NewPublishedReceiptQuery(),
// 				ScrambleNodeService:   &receiptSrvMockScrambleNodeService{},
// 				ReceiptUtil:           &receiptSrvMockReceiptUtilSuccess{},
// 				NodeConfiguration:     &mockNodeConfigurationService{},
// 				Logger:                log.New(),
// 			},
// 			args: args{
// 				numberOfReceipt: 2,
// 				blockHeight:     2 * constant.BatchReceiptLookBackHeight,
// 				secretPhrase:    "test",
// 			},
// 			want: []*model.PublishedReceipt{
// 				mockReceiptToPublish,
// 				mockReceiptToPublish,
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			rs := &ReceiptService{
// 				NodeReceiptQuery:         tt.fields.NodeReceiptQuery,
// 				MerkleTreeQuery:          tt.fields.MerkleTreeQuery,
// 				NodeRegistrationQuery:    tt.fields.NodeRegistrationQuery,
// 				BlockQuery:               tt.fields.BlockQuery,
// 				TransactionQuery:         tt.fields.TransactionQuery,
// 				QueryExecutor:            tt.fields.QueryExecutor,
// 				NodeRegistrationService:  tt.fields.NodeRegistrationService,
// 				Signature:                tt.fields.Signature,
// 				PublishedReceiptQuery:    tt.fields.PublishedReceiptQuery,
// 				ReceiptUtil:              tt.fields.ReceiptUtil,
// 				MainBlockStateStorage:    tt.fields.MainBlockStateStorage,
// 				ScrambleNodeService:      tt.fields.ScrambleNodeService,
// 				ReceiptReminderStorage:   tt.fields.ReceiptReminderStorage,
// 				BatchReceiptCacheStorage: tt.fields.BatchReceiptCacheStorage,
// 				MainBlocksStorage:        tt.fields.MainBlocksStorage,
// 				LastMerkleRoot:           tt.fields.LastMerkleRoot,
// 				MerkleRootUtil:           tt.fields.MerkleRootUtil,
// 				NodeConfiguration:        tt.fields.NodeConfiguration,
// 				Logger:                   tt.fields.Logger,
// 			}
// 			got, err := rs.SelectLinkedReceipts(tt.args.numberOfUnlinkedReceipts, tt.args.numberOfReceipt, tt.args.blockHeight,
// 				tt.args.blockSeed)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("SelectLinkedReceipts() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("SelectLinkedReceipts() got = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestReceiptService_ValidateUnlinkedReceipts(t *testing.T) {
	var mockRecSrvPublishedReceipt = []*model.PublishedReceipt{
		{
			Receipt: &model.Receipt{
				SenderPublicKey:      make([]byte, 32),
				RecipientPublicKey:   make([]byte, 32),
				DatumType:            1,
				DatumHash:            []byte{1, 1, 1, 1, 1, 1, 1, 1},
				ReferenceBlockHeight: 0,
				ReferenceBlockHash:   make([]byte, 32),
				RMR:                  nil,
				RecipientSignature:   make([]byte, 64),
			},
			IntermediateHashes: nil,
			BlockHeight:        1,
			PublishedIndex:     0,
			RMRLinked:          make([]byte, 32),
			RMRLinkedIndex:     uint32(0),
		},
	}
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
		MerkleRootUtil           util.MerkleRootInterface
		NodeConfiguration        NodeConfigurationServiceInterface
		Logger                   *log.Logger
	}
	type args struct {
		receiptsToValidate []*model.PublishedReceipt
		blockToValidate    *model.Block
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		wantValidReceipts []*model.PublishedReceipt
		wantErr           bool
	}{
		{
			name: "ValidateUnlinkedReceipts:fail-{BlockHeightTooLow}",
			fields: fields{
				NodeConfiguration: &mockNodeConfigurationService{},
			},
			args: args{
				receiptsToValidate: mockRecSrvPublishedReceipt,
				blockToValidate: &model.Block{
					Height: 10,
				},
			},
		},
		{
			name: "ValidateUnlinkedReceipts:success",
			fields: fields{
				QueryExecutor:         &mockQueryExecutorSuccessSelectUnlinked{},
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				NodeReceiptQuery:      query.NewBatchReceiptQuery(),
				TransactionQuery:      query.NewTransactionQuery(&chaintype.MainChain{}),
				PublishedReceiptQuery: query.NewPublishedReceiptQuery(),
				MerkleTreeQuery:       query.NewMerkleTreeQuery(),
				ScrambleNodeService:   &receiptSrvMockScrambleNodeService{},
				ReceiptUtil:           &receiptSrvMockReceiptUtilSuccess{},
				NodeConfiguration:     &mockNodeConfigurationService{},
			},
			args: args{
				receiptsToValidate: mockRecSrvPublishedReceipt,
				blockToValidate:    mockRecSrvBlock1,
			},
			wantValidReceipts: mockRecSrvPublishedReceipt,
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
				MerkleRootUtil:           tt.fields.MerkleRootUtil,
				NodeConfiguration:        tt.fields.NodeConfiguration,
				Logger:                   tt.fields.Logger,
			}
			gotValidReceipts, err := rs.ValidateUnlinkedReceipts(tt.args.receiptsToValidate, tt.args.blockToValidate)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUnlinkedReceipts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotValidReceipts, tt.wantValidReceipts) {
				t.Errorf("ValidateUnlinkedReceipts() gotValidReceipts = %v, want %v", gotValidReceipts, tt.wantValidReceipts)
			}
		})
	}
}
