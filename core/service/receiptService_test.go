package service

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"golang.org/x/crypto/sha3"

	"github.com/zoobc/zoobc-core/common/util"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
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
	mockLinkedReceipt = &model.Receipt{
		BatchReceipt: &model.BatchReceipt{
			SenderPublicKey:      make([]byte, 32),
			RecipientPublicKey:   make([]byte, 32),
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
)

func fixtureGenerateMerkle() {
	mockMerkle = &util.MerkleRoot{}
	receiptBytes := util.GetSignedBatchReceiptBytes(mockLinkedReceipt.BatchReceipt)
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
	case "SELECT id, tree, timestamp FROM merkle_tree AS mt WHERE EXISTS " +
		"(SELECT rmr_linked FROM published_receipt AS pr WHERE mt.id = pr.rmr_linked AND " +
		"block_height >= 0 AND block_height <= 1000 ) LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "Tree", "Timestamp",
		}).AddRow(
			[]byte{},
			[]byte{},
			1*constant.OneZBC,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
		"reference_block_hash, rmr_linked, recipient_signature, rmr, rmr_index FROM node_receipt AS rc " +
		"WHERE rc.rmr = ? AND NOT EXISTS (SELECT datum_hash FROM published_receipt AS pr WHERE " +
		"pr.datum_hash = rc.datum_hash AND pr.recipient_public_key = rc.recipient_public_key) " +
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
	case "SELECT id, tree, timestamp FROM merkle_tree AS mt WHERE EXISTS " +
		"(SELECT rmr_linked FROM published_receipt AS pr WHERE mt.id = pr.rmr_linked AND " +
		"block_height >= 0 AND block_height <= 1000 ) LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "Tree", "Timestamp",
		}).AddRow(
			mockReceiptRMR.Bytes(),
			mockFlattenTree,
			1*constant.OneZBC,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
		"reference_block_hash, rmr_linked, recipient_signature, rmr, rmr_index FROM node_receipt AS rc " +
		"WHERE rc.rmr = ? AND NOT EXISTS (SELECT datum_hash FROM published_receipt AS pr WHERE " +
		"pr.datum_hash = rc.datum_hash AND pr.recipient_public_key = rc.recipient_public_key) " +
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
	}

	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockQueryExecutorSuccessOneLinkedReceiptsAndMore) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, tree, timestamp FROM merkle_tree AS mt WHERE EXISTS " +
		"(SELECT rmr_linked FROM published_receipt AS pr WHERE mt.id = pr.rmr_linked AND " +
		"block_height >= 0 AND block_height <= 1000 ) LIMIT 3":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "Tree", "Timestamp",
		}).AddRow(
			mockReceiptRMR.Bytes(),
			mockFlattenTree,
			1*constant.OneZBC,
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
		"reference_block_hash, rmr_linked, recipient_signature, rmr, rmr_index FROM node_receipt AS rc " +
		"WHERE rc.rmr = ? AND NOT EXISTS (SELECT datum_hash FROM published_receipt AS pr WHERE " +
		"pr.datum_hash = rc.datum_hash AND pr.recipient_public_key = rc.recipient_public_key) " +
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
		"reference_block_hash, rmr_linked, recipient_signature, rmr, rmr_index FROM node_receipt AS rc WHERE " +
		"rmr_linked IS NOT NULL AND NOT EXISTS (SELECT datum_hash FROM published_receipt AS pr WHERE " +
		"pr.datum_hash == rc.datum_hash) GROUP BY recipient_public_key LIMIT 0, 2":
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
		))
	case "SELECT sender_public_key, recipient_public_key, datum_type, datum_hash, reference_block_height, " +
		"reference_block_hash, rmr_linked, recipient_signature, rmr, rmr_index FROM node_receipt AS rc WHERE " +
		"NOT EXISTS (SELECT datum_hash FROM published_receipt AS pr WHERE pr.datum_hash == rc.datum_hash) " +
		"GROUP BY recipient_public_key LIMIT 0, 1":
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

func TestReceiptService_SelectReceipts(t *testing.T) {
	// prepare testing env
	fixtureGenerateMerkle()
	type fields struct {
		ReceiptQuery    query.ReceiptQueryInterface
		MerkleTreeQuery query.MerkleTreeQueryInterface
		KVExecutor      kvdb.KVExecutorInterface
		QueryExecutor   query.ExecutorInterface
	}
	type args struct {
		blockTimestamp  int64
		numberOfReceipt int
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
				ReceiptQuery:    nil,
				MerkleTreeQuery: query.NewMerkleTreeQuery(),
				KVExecutor:      nil,
				QueryExecutor:   &mockQueryExecutorFailExecuteSelect{},
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
				ReceiptQuery:    nil,
				MerkleTreeQuery: &mockMerkleTreeQueryFailBuildTree{},
				KVExecutor:      nil,
				QueryExecutor:   &mockQueryExecutorSuccessMerkle{},
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
				ReceiptQuery:    query.NewReceiptQuery(),
				MerkleTreeQuery: query.NewMerkleTreeQuery(),
				KVExecutor:      nil,
				QueryExecutor:   &mockQueryExecutorFailExecuteSelectReceipt{},
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
				ReceiptQuery:    query.NewReceiptQuery(),
				MerkleTreeQuery: query.NewMerkleTreeQuery(),
				KVExecutor:      nil,
				QueryExecutor:   &mockQueryExecutorSuccessOneLinkedReceipts{},
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
				ReceiptQuery:    query.NewReceiptQuery(),
				MerkleTreeQuery: query.NewMerkleTreeQuery(),
				KVExecutor:      nil,
				QueryExecutor:   &mockQueryExecutorSuccessOneLinkedReceiptsAndMore{},
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
				{
					BatchReceipt:       mockUnlinkedReceiptWithLinkedRMR.BatchReceipt,
					IntermediateHashes: nil,
					BlockHeight:        0,
					ReceiptIndex:       mockUnlinkedReceiptWithLinkedRMR.RMRIndex,
					PublishedIndex:     0,
				},
				{
					BatchReceipt:       mockUnlinkedReceipt.BatchReceipt,
					IntermediateHashes: nil,
					BlockHeight:        0,
					ReceiptIndex:       mockUnlinkedReceipt.RMRIndex,
					PublishedIndex:     0,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ReceiptService{
				ReceiptQuery:    tt.fields.ReceiptQuery,
				MerkleTreeQuery: tt.fields.MerkleTreeQuery,
				KVExecutor:      tt.fields.KVExecutor,
				QueryExecutor:   tt.fields.QueryExecutor,
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
