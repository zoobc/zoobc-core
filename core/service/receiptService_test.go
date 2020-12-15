package service

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/signaturetype"
	"reflect"
	"regexp"
	"testing"

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
)

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

func (*mockQueryExecutorFailExecuteSelectReceipt) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, block_height, tree, timestamp FROM merkle_tree AS mt WHERE EXISTS " +
		"(SELECT rmr_linked FROM published_receipt AS pr WHERE mt.id = pr.rmr_linked)" +
		" AND block_height BETWEEN 280 AND 1000 ORDER BY block_height ASC LIMIT 5":
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
		"reference_block_height BETWEEN 280 AND 1000 " +
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
		"BETWEEN 280 AND 1000 ORDER BY block_height ASC LIMIT 5":
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
		"reference_block_height BETWEEN 280 AND 1000 " +
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
		"BETWEEN 280 AND 1000 GROUP BY recipient_public_key ORDER BY reference_block_height ASC LIMIT 5":
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
		"AND block_height BETWEEN 280 AND 1000 ORDER BY block_height ASC LIMIT 15":
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
		"reference_block_height BETWEEN 280 AND 1000 " +
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
		"reference_block_height BETWEEN 280 AND 1000 GROUP BY recipient_public_key ORDER BY reference_block_height " +
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

func TestReceiptService_SelectReceipts(t *testing.T) {
	// prepare testing env
	fixtureGenerateMerkle()

	type fields struct {
		NodeReceiptQuery        query.BatchReceiptQueryInterface
		MerkleTreeQuery         query.MerkleTreeQueryInterface
		QueryExecutor           query.ExecutorInterface
		NodeRegistrationService NodeRegistrationServiceInterface
		ScrambleNodeService     ScrambleNodeServiceInterface
		MainBlocksStorage       storage.CacheStackStorageInterface
	}
	type args struct {
		blockTimestamp  int64
		numberOfReceipt uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.PublishedReceipt
		wantErr bool
	}{
		{
			name: "receiptService-selectReceipts-Fail:selectDB-error",
			fields: fields{
				MerkleTreeQuery:     query.NewMerkleTreeQuery(),
				ScrambleNodeService: &mockScrambleNodeServiceSelectReceiptsSuccess{},
				NodeReceiptQuery:    nil,
				QueryExecutor:       &mockQueryExecutorFailExecuteSelect{},
			},
			args: args{
				blockTimestamp:  0,
				numberOfReceipt: 1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "receiptService-selectReceipts-Fail:MerkleTreeQuery-BuildTree-Fail",
			fields: fields{
				QueryExecutor:       &mockQueryExecutorSuccessMerkle{},
				ScrambleNodeService: &mockScrambleNodeServiceSelectReceiptsSuccess{},
				NodeReceiptQuery:    nil,
				MerkleTreeQuery:     &mockMerkleTreeQueryFailBuildTree{},
			},
			args: args{
				blockTimestamp:  0,
				numberOfReceipt: 1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "receiptService-selectReceipts-Fail:ExecuteSelect-Fail_Receipt",
			fields: fields{
				ScrambleNodeService: &mockScrambleNodeServiceSelectReceiptsSuccess{},
				NodeReceiptQuery:    query.NewBatchReceiptQuery(),
				MerkleTreeQuery:     query.NewMerkleTreeQuery(),
				QueryExecutor:       &mockQueryExecutorFailExecuteSelectReceipt{},
			},
			args: args{
				blockTimestamp:  0,
				numberOfReceipt: 1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "receiptService-selectReceipts-success-one-linked",
			fields: fields{
				NodeReceiptQuery:        query.NewBatchReceiptQuery(),
				MerkleTreeQuery:         query.NewMerkleTreeQuery(),
				QueryExecutor:           &mockQueryExecutorSuccessOneLinkedReceipts{},
				NodeRegistrationService: &mockNodeRegistrationSelectReceiptSuccess{},
				ScrambleNodeService:     &mockScrambleNodeServiceSelectReceiptsSuccess{},
				MainBlocksStorage:       &mockSelectReceiptsMainBlocksStorageSuccess{},
			},
			args: args{
				blockTimestamp:  0,
				numberOfReceipt: 1,
			},
			want: []*model.PublishedReceipt{
				{
					Receipt:            mockLinkedReceipt.Receipt,
					IntermediateHashes: mockFlattenIntermediateHash,
					BlockHeight:        0,
					ReceiptIndex:       mockLinkedReceipt.RMRIndex,
					PublishedIndex:     0,
				},
			},
			wantErr: false,
		},
		{
			name: "receiptService-selectReceipts-success-one-linked-more-rmr-linked-and-unlinked",
			fields: fields{
				NodeReceiptQuery:        query.NewBatchReceiptQuery(),
				MerkleTreeQuery:         query.NewMerkleTreeQuery(),
				NodeRegistrationService: &mockNodeRegistrationSelectReceiptSuccess{},
				QueryExecutor:           &mockQueryExecutorSuccessOneLinkedReceiptsAndMore{},
				ScrambleNodeService:     &mockScrambleNodeServiceSelectReceiptsSuccess{},
				MainBlocksStorage:       &mockSelectReceiptsMainBlocksStorageSuccess{},
			},
			args: args{
				blockTimestamp:  0,
				numberOfReceipt: 3,
			},
			want: []*model.PublishedReceipt{
				{
					Receipt:            mockLinkedReceipt.Receipt,
					IntermediateHashes: mockFlattenIntermediateHash,
					BlockHeight:        0,
					ReceiptIndex:       mockLinkedReceipt.RMRIndex,
					PublishedIndex:     0,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ReceiptService{
				NodeReceiptQuery:        tt.fields.NodeReceiptQuery,
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				MerkleTreeQuery:         tt.fields.MerkleTreeQuery,
				QueryExecutor:           tt.fields.QueryExecutor,
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				Signature:               crypto.NewSignature(),
				NodeRegistrationService: tt.fields.NodeRegistrationService,
				ReceiptUtil:             &coreUtil.ReceiptUtil{},
				ScrambleNodeService:     tt.fields.ScrambleNodeService,
				MainBlocksStorage:       tt.fields.MainBlocksStorage,
			}
			got, err := rs.SelectReceipts(tt.args.blockTimestamp, tt.args.numberOfReceipt, 1000)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectReceipts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SelectReceipts() got = %v, want %v", got, tt.want)
			}
		})
	}
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
	default:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).
			WillReturnRows(sqlmock.NewRows([]string{"total_record"}).AddRow(constant.ReceiptBatchMaximum))
	}

	return db.QueryRow(qStr), nil
}

func (*mockQueryExecutorGenerateReceiptsMerkleRootSuccess) BeginTx() error {
	return nil
}
func (*mockQueryExecutorGenerateReceiptsMerkleRootSuccess) CommitTx() error {
	return nil
}
func (*mockQueryExecutorGenerateReceiptsMerkleRootSuccess) RollbackTx() error {
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
func (*mockQueryExecutorGenerateReceiptsMerkleRootSelectFail) BeginTx() error {
	return errors.New("mockError:BeginTxFail")
}

func (*mockQueryExecutorGenerateReceiptsMerkleRootSelectFail) CommitTx() error {
	return errors.New("mockError:CommitTxFail")
}
func (*mockQueryExecutorGenerateReceiptsMerkleRootSelectFail) RollbackTx() error {
	return errors.New("mockError:RollbackTxFail")
}
func (*mockQueryExecutorGenerateReceiptsMerkleRootSelectFail) ExecuteTransactions(queries [][]interface{}) error {
	return errors.New("mockError:ExecuteTransactionsFail")
}

func TestReceiptService_GenerateReceiptsMerkleRoot(t *testing.T) {
	type fields struct {
		NodeReceiptQuery      query.BatchReceiptQueryInterface
		MerkleTreeQuery       query.MerkleTreeQueryInterface
		QueryExecutor         query.ExecutorInterface
		MainBlockStateStorage storage.CacheStorageInterface
		BatchReceiptStorage   storage.CacheStorageInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				NodeReceiptQuery:      query.NewBatchReceiptQuery(),
				MerkleTreeQuery:       query.NewMerkleTreeQuery(),
				QueryExecutor:         &mockQueryExecutorGenerateReceiptsMerkleRootSuccess{},
				MainBlockStateStorage: &mockGenerateReceiptsMerkleRootMainBlockStateStorageSuccess{},
				BatchReceiptStorage:   storage.NewReceiptPoolCacheStorage(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ReceiptService{
				NodeReceiptQuery:         tt.fields.NodeReceiptQuery,
				MerkleTreeQuery:          tt.fields.MerkleTreeQuery,
				BlockQuery:               query.NewBlockQuery(&chaintype.MainChain{}),
				QueryExecutor:            tt.fields.QueryExecutor,
				ReceiptUtil:              &coreUtil.ReceiptUtil{},
				MainBlockStateStorage:    tt.fields.MainBlockStateStorage,
				BatchReceiptCacheStorage: tt.fields.BatchReceiptStorage,
			}
			if err := rs.GenerateReceiptsMerkleRoot(); (err != nil) != tt.wantErr {
				t.Errorf("ReceiptService.GenerateReceiptsMerkleRoot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockExecutorPruningNodeReceiptsSuccess struct {
		query.Executor
	}
)

func (*mockExecutorPruningNodeReceiptsSuccess) BeginTx() error {
	return nil
}
func (*mockExecutorPruningNodeReceiptsSuccess) CommitTx() error {
	return nil
}
func (*mockExecutorPruningNodeReceiptsSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}
func (*mockExecutorPruningNodeReceiptsSuccess) RollbackTx() error {
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
