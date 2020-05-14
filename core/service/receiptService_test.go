package service

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
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
	mockLinkedReceipt = &model.Receipt{
		BatchReceipt: &model.BatchReceipt{
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
	mockUnlinkedReceiptWithLinkedRMR = &model.Receipt{
		BatchReceipt: &model.BatchReceipt{
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
	mockUnlinkedReceipt = &model.Receipt{
		BatchReceipt: &model.BatchReceipt{
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
		NodeID:             0,
		NodePublicKey:      mockLinkedReceipt.BatchReceipt.SenderPublicKey,
		AccountAddress:     "",
		RegistrationHeight: 0,
		NodeAddress: &model.NodeAddress{
			Address: "0.0.0.0",
			Port:    8001,
		},
		LockedBalance:      0,
		RegistrationStatus: 0,
		Latest:             false,
		Height:             0,
	}
	mockNodeRegistrationDataB = model.NodeRegistration{
		NodeID:             0,
		NodePublicKey:      mockLinkedReceipt.BatchReceipt.RecipientPublicKey,
		AccountAddress:     "",
		RegistrationHeight: 0,
		NodeAddress: &model.NodeAddress{
			Address: "0.0.0.0",
			Port:    8002,
		},
		LockedBalance:      0,
		RegistrationStatus: 0,
		Latest:             false,
		Height:             0,
	}
)

func fixtureGenerateMerkle() {
	mockSeed := "mock seed"
	signature := crypto.NewSignature()
	receiptUtil := &coreUtil.ReceiptUtil{}
	// sign mock linked receipt and update the recipient public key
	mockLinkedReceipt.BatchReceipt.RecipientPublicKey = crypto.NewEd25519Signature().GetPublicKeyFromSeed(mockSeed)
	unsignedReceiptByte := receiptUtil.GetUnsignedBatchReceiptBytes(mockLinkedReceipt.BatchReceipt)
	mockLinkedReceipt.BatchReceipt.RecipientSignature = signature.SignByNode(unsignedReceiptByte, mockSeed)
	// sign rmr linked receipt
	mockUnlinkedReceiptWithLinkedRMR.BatchReceipt.RecipientPublicKey = crypto.NewEd25519Signature().GetPublicKeyFromSeed(mockSeed)
	mockUnlinkedReceiptWithLinkedRMR.BatchReceipt.SenderPublicKey = mockLinkedReceipt.BatchReceipt.SenderPublicKey
	unsignedUnlinkedReceiptByte := receiptUtil.GetUnsignedBatchReceiptBytes(mockUnlinkedReceiptWithLinkedRMR.BatchReceipt)
	mockUnlinkedReceiptWithLinkedRMR.BatchReceipt.RecipientSignature = signature.SignByNode(
		unsignedUnlinkedReceiptByte, mockSeed)
	// sign no rmr linked
	mockUnlinkedReceipt.BatchReceipt.RecipientPublicKey = crypto.NewEd25519Signature().GetPublicKeyFromSeed(mockSeed)
	mockUnlinkedReceipt.BatchReceipt.SenderPublicKey = mockLinkedReceipt.BatchReceipt.SenderPublicKey
	unsignedNoRMRReceiptByte := receiptUtil.GetUnsignedBatchReceiptBytes(mockUnlinkedReceipt.BatchReceipt)
	mockUnlinkedReceipt.BatchReceipt.RecipientSignature = signature.SignByNode(
		unsignedNoRMRReceiptByte, mockSeed,
	)
	mockNodeRegistrationDataB.NodePublicKey = mockLinkedReceipt.BatchReceipt.RecipientPublicKey
	mockMerkle = &util.MerkleRoot{}
	receiptBytes := receiptUtil.GetSignedBatchReceiptBytes(mockLinkedReceipt.BatchReceipt)
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
		"(SELECT rmr_linked FROM published_receipt AS pr WHERE mt.id = pr.rmr_linked) " +
		"AND block_height BETWEEN 280 AND 1000 ORDER BY block_height ASC LIMIT 5":
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
			mockLinkedReceipt.BatchReceipt.SenderPublicKey,
			mockLinkedReceipt.BatchReceipt.RecipientPublicKey,
			mockLinkedReceipt.BatchReceipt.DatumType,
			mockLinkedReceipt.BatchReceipt.DatumHash,
			mockLinkedReceipt.BatchReceipt.ReferenceBlockHeight,
			mockLinkedReceipt.BatchReceipt.ReferenceBlockHash,
			mockLinkedReceipt.BatchReceipt.RMRLinked,
			mockLinkedReceipt.BatchReceipt.RecipientSignature,
			mockReceiptRMR.Bytes(),
			0,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
		"reference_block_hash, rmr_linked, recipient_signature, rmr, rmr_index FROM node_receipt AS rc WHERE NOT " +
		"EXISTS (SELECT datum_hash FROM published_receipt AS pr WHERE pr.datum_hash == rc.datum_hash) AND " +
		"reference_block_height BETWEEN 0 AND 1000 GROUP BY recipient_public_key ORDER BY reference_block_height ASC " +
		"LIMIT 5":
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
		}))

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
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND height <= ? ORDER BY " +
		"height DESC LIMIT 1":
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
					fmt.Sprintf("%s:%d",
						mockNodeRegistrationData.NodeAddress.Address, mockNodeRegistrationData.NodeAddress.Port,
					),
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
					fmt.Sprintf("%s:%d",
						mockNodeRegistrationDataB.NodeAddress.Address, mockNodeRegistrationDataB.NodeAddress.Port,
					),
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
				mockBlockDataSelectReceipt.GetID(),
				mockBlockDataSelectReceipt.GetBlockHash(),
				mockBlockDataSelectReceipt.GetPreviousBlockHash(),
				mockBlockDataSelectReceipt.GetHeight(),
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
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, " +
		"registration_status, latest, height FROM node_registry WHERE node_public_key = ? AND height <= ? ORDER BY " +
		"height DESC LIMIT 1":
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
					fmt.Sprintf("%s:%d",
						mockNodeRegistrationData.NodeAddress.Address, mockNodeRegistrationData.NodeAddress.Port,
					),
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
					fmt.Sprintf("%s:%d",
						mockNodeRegistrationDataB.NodeAddress.Address, mockNodeRegistrationDataB.NodeAddress.Port,
					),
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
				mockBlockDataSelectReceipt.GetID(),
				mockBlockDataSelectReceipt.GetBlockHash(),
				mockBlockDataSelectReceipt.GetPreviousBlockHash(),
				mockBlockDataSelectReceipt.GetHeight(),
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
			mockLinkedReceipt.BatchReceipt.SenderPublicKey,
			mockLinkedReceipt.BatchReceipt.RecipientPublicKey,
			mockLinkedReceipt.BatchReceipt.DatumType,
			mockLinkedReceipt.BatchReceipt.DatumHash,
			mockLinkedReceipt.BatchReceipt.ReferenceBlockHeight,
			mockLinkedReceipt.BatchReceipt.ReferenceBlockHash,
			mockLinkedReceipt.BatchReceipt.RMRLinked,
			mockLinkedReceipt.BatchReceipt.RecipientSignature,
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
			mockUnlinkedReceiptWithLinkedRMR.BatchReceipt.SenderPublicKey,
			mockUnlinkedReceiptWithLinkedRMR.BatchReceipt.RecipientPublicKey,
			mockUnlinkedReceiptWithLinkedRMR.BatchReceipt.DatumType,
			mockUnlinkedReceiptWithLinkedRMR.BatchReceipt.DatumHash,
			mockUnlinkedReceiptWithLinkedRMR.BatchReceipt.ReferenceBlockHeight,
			mockUnlinkedReceiptWithLinkedRMR.BatchReceipt.ReferenceBlockHash,
			mockUnlinkedReceiptWithLinkedRMR.BatchReceipt.RMRLinked,
			mockUnlinkedReceiptWithLinkedRMR.BatchReceipt.RecipientSignature,
			make([]byte, 32),
			0,
		).AddRow(
			mockUnlinkedReceipt.BatchReceipt.SenderPublicKey,
			mockUnlinkedReceipt.BatchReceipt.RecipientPublicKey,
			mockUnlinkedReceipt.BatchReceipt.DatumType,
			mockUnlinkedReceipt.BatchReceipt.DatumHash,
			mockUnlinkedReceipt.BatchReceipt.ReferenceBlockHeight,
			mockUnlinkedReceipt.BatchReceipt.ReferenceBlockHash,
			mockUnlinkedReceipt.BatchReceipt.RMRLinked,
			mockUnlinkedReceipt.BatchReceipt.RecipientSignature,
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

func (*mockNodeRegistrationSelectReceiptSuccess) GetScrambleNodesByHeight(
	blockHeight uint32,
) (*model.ScrambledNodes, error) {
	indexA := 1
	indexB := 2
	indexC := 3
	indexD := 4
	indexE := 5
	return &model.ScrambledNodes{
		AddressNodes: []*model.Peer{
			{
				Info: &model.Node{
					Address: "0.0.0.0",
					Port:    8001,
				},
			},
			{
				Info: &model.Node{
					Address: "0.0.0.0",
					Port:    8002,
				},
			},
			{
				Info: &model.Node{
					Address: "0.0.0.0",
					Port:    8003,
				},
			},
			{
				Info: &model.Node{
					Address: "0.0.0.0",
					Port:    8004,
				},
			},
			{
				Info: &model.Node{
					Address: "0.0.0.0",
					Port:    8005,
				},
			},
		},
		IndexNodes: map[string]*int{
			"0.0.0.0:8001": &indexA,
			"0.0.0.0:8002": &indexB,
			"0.0.0.0:8003": &indexC,
			"0.0.0.0:8004": &indexD,
			"0.0.0.0:8005": &indexE,
		},
		BlockHeight: blockHeight,
	}, nil
}

func TestReceiptService_SelectReceipts(t *testing.T) {
	// prepare testing env
	fixtureGenerateMerkle()

	type fields struct {
		NodeReceiptQuery        query.NodeReceiptQueryInterface
		MerkleTreeQuery         query.MerkleTreeQueryInterface
		KVExecutor              kvdb.KVExecutorInterface
		QueryExecutor           query.ExecutorInterface
		NodeRegistrationService NodeRegistrationServiceInterface
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
				NodeReceiptQuery: nil,
				MerkleTreeQuery:  query.NewMerkleTreeQuery(),
				KVExecutor:       nil,
				QueryExecutor:    &mockQueryExecutorFailExecuteSelect{},
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
				NodeReceiptQuery: nil,
				MerkleTreeQuery:  &mockMerkleTreeQueryFailBuildTree{},
				KVExecutor:       nil,
				QueryExecutor:    &mockQueryExecutorSuccessMerkle{},
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
				NodeReceiptQuery: query.NewNodeReceiptQuery(),
				MerkleTreeQuery:  query.NewMerkleTreeQuery(),
				KVExecutor:       nil,
				QueryExecutor:    &mockQueryExecutorFailExecuteSelectReceipt{},
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
				NodeReceiptQuery:        query.NewNodeReceiptQuery(),
				MerkleTreeQuery:         query.NewMerkleTreeQuery(),
				KVExecutor:              nil,
				QueryExecutor:           &mockQueryExecutorSuccessOneLinkedReceipts{},
				NodeRegistrationService: &mockNodeRegistrationSelectReceiptSuccess{},
			},
			args: args{
				blockTimestamp:  0,
				numberOfReceipt: 1,
			},
			want: []*model.PublishedReceipt{
				{
					BatchReceipt:       mockLinkedReceipt.BatchReceipt,
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
				NodeReceiptQuery:        query.NewNodeReceiptQuery(),
				MerkleTreeQuery:         query.NewMerkleTreeQuery(),
				KVExecutor:              nil,
				NodeRegistrationService: &mockNodeRegistrationSelectReceiptSuccess{},
				QueryExecutor:           &mockQueryExecutorSuccessOneLinkedReceiptsAndMore{},
			},
			args: args{
				blockTimestamp:  0,
				numberOfReceipt: 3,
			},
			want: []*model.PublishedReceipt{
				{
					BatchReceipt:       mockLinkedReceipt.BatchReceipt,
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
				KVExecutor:              tt.fields.KVExecutor,
				QueryExecutor:           tt.fields.QueryExecutor,
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				Signature:               crypto.NewSignature(),
				NodeRegistrationService: tt.fields.NodeRegistrationService,
				ReceiptUtil:             &coreUtil.ReceiptUtil{},
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
)

func (*mockQueryExecutorGenerateReceiptsMerkleRootSuccess) ExecuteSelectRow(
	qStr string, tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	switch qStr {
	case "SELECT id, block_hash, previous_block_hash, height, timestamp, block_seed, block_signature, " +
		"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, " +
		"total_fee, total_coinbase, version FROM main_block ORDER BY height DESC LIMIT 1":
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
	default:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).
			WillReturnRows(sqlmock.NewRows([]string{"total_record"}).AddRow(constant.ReceiptBatchMaximum))
	}

	return db.QueryRow(qStr), nil
}

func (*mockQueryExecutorGenerateReceiptsMerkleRootSuccess) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mockedRows := sqlmock.NewRows(query.NewBatchReceiptQuery().Fields)
	mockedRows.AddRow(
		mockLinkedReceipt.BatchReceipt.GetSenderPublicKey(),
		mockLinkedReceipt.BatchReceipt.GetRecipientPublicKey(),
		mockLinkedReceipt.BatchReceipt.GetDatumType(),
		mockLinkedReceipt.BatchReceipt.GetDatumHash(),
		mockLinkedReceipt.BatchReceipt.GetReferenceBlockHeight(),
		mockLinkedReceipt.BatchReceipt.GetReferenceBlockHash(),
		mockLinkedReceipt.BatchReceipt.GetRMRLinked(),
		mockLinkedReceipt.BatchReceipt.GetRecipientSignature(),
	)

	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(mockedRows)
	rows, _ := db.Query(qe)
	return rows, nil
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
		NodeReceiptQuery  query.NodeReceiptQueryInterface
		BatchReceiptQuery query.BatchReceiptQueryInterface
		MerkleTreeQuery   query.MerkleTreeQueryInterface
		KVExecutor        kvdb.KVExecutorInterface
		QueryExecutor     query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				NodeReceiptQuery:  query.NewNodeReceiptQuery(),
				BatchReceiptQuery: query.NewBatchReceiptQuery(),
				MerkleTreeQuery:   query.NewMerkleTreeQuery(),
				KVExecutor:        nil,
				QueryExecutor:     &mockQueryExecutorGenerateReceiptsMerkleRootSuccess{},
			},
			wantErr: false,
		},
		{
			name: "wantError:SelectRowFail",
			fields: fields{
				NodeReceiptQuery:  nil,
				BatchReceiptQuery: query.NewBatchReceiptQuery(),
				MerkleTreeQuery:   nil,
				KVExecutor:        nil,
				QueryExecutor:     &mockQueryExecutorGenerateReceiptsMerkleRootSelectRowFail{},
			},
			wantErr: true,
		},
		{
			name: "wantError:SelectFail",
			fields: fields{
				NodeReceiptQuery:  nil,
				BatchReceiptQuery: query.NewBatchReceiptQuery(),
				MerkleTreeQuery:   nil,
				KVExecutor:        nil,
				QueryExecutor:     &mockQueryExecutorGenerateReceiptsMerkleRootSelectFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ReceiptService{
				NodeReceiptQuery:  tt.fields.NodeReceiptQuery,
				BatchReceiptQuery: tt.fields.BatchReceiptQuery,
				MerkleTreeQuery:   tt.fields.MerkleTreeQuery,
				BlockQuery:        query.NewBlockQuery(&chaintype.MainChain{}),
				KVExecutor:        tt.fields.KVExecutor,
				QueryExecutor:     tt.fields.QueryExecutor,
				ReceiptUtil:       &coreUtil.ReceiptUtil{},
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
		mockBlockDataSelectReceipt.GetID(),
		mockBlockDataSelectReceipt.GetBlockHash(),
		mockBlockDataSelectReceipt.GetPreviousBlockHash(),
		mockBlockDataSelectReceipt.GetHeight(),
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
	)
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(mockRow)
	return db.QueryRow(qStr), nil
}
func TestReceiptService_PruningNodeReceipts(t *testing.T) {
	type fields struct {
		NodeReceiptQuery        query.NodeReceiptQueryInterface
		BatchReceiptQuery       query.BatchReceiptQueryInterface
		MerkleTreeQuery         query.MerkleTreeQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		KVExecutor              kvdb.KVExecutorInterface
		QueryExecutor           query.ExecutorInterface
		NodeRegistrationService NodeRegistrationServiceInterface
		Signature               crypto.SignatureInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "WantSuccess",
			fields: fields{
				NodeReceiptQuery: query.NewNodeReceiptQuery(),
				MerkleTreeQuery:  query.NewMerkleTreeQuery(),
				BlockQuery:       query.NewBlockQuery(chaintype.GetChainType(0)),
				QueryExecutor:    &mockExecutorPruningNodeReceiptsSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ReceiptService{
				NodeReceiptQuery:        tt.fields.NodeReceiptQuery,
				BatchReceiptQuery:       tt.fields.BatchReceiptQuery,
				MerkleTreeQuery:         tt.fields.MerkleTreeQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				KVExecutor:              tt.fields.KVExecutor,
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeRegistrationService: tt.fields.NodeRegistrationService,
				Signature:               tt.fields.Signature,
			}
			if err := rs.PruningNodeReceipts(); (err != nil) != tt.wantErr {
				t.Errorf("PruningNodeReceipts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
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
	)
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(mockRows)
	return db.Query(qStr)
}
func TestReceiptService_GetPublishedReceiptsByHeight(t *testing.T) {
	type fields struct {
		NodeReceiptQuery        query.NodeReceiptQueryInterface
		BatchReceiptQuery       query.BatchReceiptQueryInterface
		MerkleTreeQuery         query.MerkleTreeQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		KVExecutor              kvdb.KVExecutorInterface
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
				BatchReceiptQuery:       tt.fields.BatchReceiptQuery,
				MerkleTreeQuery:         tt.fields.MerkleTreeQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				KVExecutor:              tt.fields.KVExecutor,
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
