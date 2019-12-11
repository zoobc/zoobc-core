package util

import (
	"database/sql"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dgraph-io/badger"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	mockGenerateBatchReceiptWithReminderKVExecutorSuccess struct {
		kvdb.KVExecutor
	}
	mockGenerateBatchReceiptWithReminderKVExecutorFailOtherError struct {
		kvdb.KVExecutor
	}
	mockGenerateBatchReceiptWithReminderQueryExecutorSuccess struct {
		query.Executor
	}
)

func (*mockGenerateBatchReceiptWithReminderKVExecutorSuccess) Insert(key string, value []byte, expiry int) error {
	return nil
}
func (*mockGenerateBatchReceiptWithReminderKVExecutorFailOtherError) Insert(key string, value []byte, expiry int) error {
	return badger.ErrInvalidKey
}

func (*mockGenerateBatchReceiptWithReminderQueryExecutorSuccess) ExecuteSelectRow(
	qStr string,
	tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{
		"ID", "Tree", "Timestamp",
	}))
	row := db.QueryRow(qStr)
	return row, nil
}

func TestGenerateBatchReceiptWithReminder(t *testing.T) {
	var (
		mockSecretPhrase = ""
		mockBlock        = &model.Block{
			ID:                   0,
			PreviousBlockHash:    nil,
			Height:               0,
			Timestamp:            0,
			BlockSeed:            nil,
			BlockSignature:       nil,
			CumulativeDifficulty: "",
			SmithScale:           0,
			BlocksmithPublicKey:  []byte{},
			TotalAmount:          0,
			TotalFee:             0,
			TotalCoinBase:        0,
			Version:              0,
			PayloadLength:        0,
			PayloadHash:          nil,
			Transactions:         nil,
		}
		mockSenderPublicKey, mockReceivedDatumHash = make([]byte, 32), make([]byte, 32)
		mockReceiptKey, _                          = util.GetReceiptKey(mockReceivedDatumHash, mockSenderPublicKey)

		mockNodePublicKey          = util.GetPublicKeyFromSeed(mockSecretPhrase)
		mockSuccessBatchReceipt, _ = util.GenerateBatchReceipt(
			mockBlock,
			mockSenderPublicKey,
			mockNodePublicKey,
			mockReceivedDatumHash,
			nil,
			constant.ReceiptDatumTypeBlock,
		)
	)

	mockSuccessBatchReceipt.RecipientSignature = (&crypto.Signature{}).SignByNode(
		util.GetUnsignedBatchReceiptBytes(mockSuccessBatchReceipt),
		mockSecretPhrase,
	)

	type args struct {
		receivedDatumHash []byte
		lastBlock         *model.Block
		senderPublicKey   []byte
		nodeSecretPhrase  string
		receiptKey        string
		datumType         uint32
		signature         crypto.SignatureInterface
		queryExecutor     query.ExecutorInterface
		kvExecutor        kvdb.KVExecutorInterface
	}
	tests := []struct {
		name    string
		args    args
		want    *model.BatchReceipt
		wantErr bool
	}{
		{
			name: "wantSuccess",
			args: args{
				receivedDatumHash: mockReceivedDatumHash,
				lastBlock:         mockBlock,
				senderPublicKey:   mockSenderPublicKey,
				nodeSecretPhrase:  "",
				receiptKey:        "test_" + string(mockReceiptKey),
				datumType:         constant.ReceiptDatumTypeBlock,
				signature:         &crypto.Signature{},
				queryExecutor:     &mockGenerateBatchReceiptWithReminderQueryExecutorSuccess{},
				kvExecutor:        &mockGenerateBatchReceiptWithReminderKVExecutorSuccess{},
			},
			want:    mockSuccessBatchReceipt,
			wantErr: false,
		},
		{
			name: "wantFail:KVDBInsertFail",
			args: args{
				receivedDatumHash: mockReceivedDatumHash,
				lastBlock:         mockBlock,
				senderPublicKey:   mockSenderPublicKey,
				nodeSecretPhrase:  "",
				receiptKey:        "test_" + string(mockReceiptKey),
				datumType:         constant.ReceiptDatumTypeBlock,
				signature:         &crypto.Signature{},
				queryExecutor:     &mockGenerateBatchReceiptWithReminderQueryExecutorSuccess{},
				kvExecutor:        &mockGenerateBatchReceiptWithReminderKVExecutorFailOtherError{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateBatchReceiptWithReminder(
				tt.args.receivedDatumHash,
				tt.args.lastBlock,
				tt.args.senderPublicKey,
				tt.args.nodeSecretPhrase,
				tt.args.receiptKey,
				tt.args.datumType,
				tt.args.signature,
				tt.args.queryExecutor,
				tt.args.kvExecutor)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateBatchReceiptWithReminder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateBatchReceiptWithReminder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNumberOfMaxReceipts(t *testing.T) {
	type args struct {
		numberOfSortedBlocksmiths int
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{
			name: "TotalBlocksmiths < PriorityConstant",
			args: args{numberOfSortedBlocksmiths: constant.PriorityStrategyMaxPriorityPeers - 1},
			want: constant.PriorityStrategyMaxPriorityPeers - 2,
		},
		{
			name: "TotalBlocksmiths > PriorityConstant",
			args: args{numberOfSortedBlocksmiths: constant.PriorityStrategyMaxPriorityPeers + 2},
			want: constant.PriorityStrategyMaxPriorityPeers,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNumberOfMaxReceipts(tt.args.numberOfSortedBlocksmiths); got != tt.want {
				t.Errorf("GetNumberOfMaxReceipts() = %v, want %v", got, tt.want)
			}
		})
	}
}
