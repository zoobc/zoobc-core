package util

import (
	"bytes"
	"database/sql"
	"math"
	"reflect"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type mockQueryExecutorSuccess struct {
	query.Executor
}

func (*mockQueryExecutorSuccess) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()

	getAccountBalanceByAccountID := "SELECT account_address,block_height,spendable_balance,balance,pop_revenue," +
		"latest FROM account_balance WHERE account_address = ? AND latest = 1"
	defer db.Close()
	switch qe {
	case getAccountBalanceByAccountID:
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"account_address", "block_height", "spendable_balance", "balance", "pop_revenue", "latest"},
		).AddRow("BCZ", 1, 10000, 10000, 0, 1))
	default:
		return nil, nil
	}

	rows, _ := db.Query(qe)
	return rows, nil
}

func TestGetTransactionBytes(t *testing.T) {
	type args struct {
		transaction *model.Transaction
		sign        bool
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "GetTransactionBytes:success",
			args: args{
				transaction: &model.Transaction{
					TransactionType:               2,
					Version:                       1,
					Timestamp:                     1562806389280,
					SenderAccountAddressLength:    uint32(len([]byte("BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7"))),
					SenderAccountAddress:          "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
					RecipientAccountAddressLength: uint32(len([]byte("BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J"))),
					RecipientAccountAddress:       "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
					Fee:                           1000000,
					TransactionBodyLength:         8,
					TransactionBodyBytes:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
					Signature: []byte{4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78,
						68, 90, 83, 142, 11, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184,
						77, 80, 80, 39, 254, 173, 28, 169},
				},
				sign: true,
			},
			want: []byte{
				2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 44, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50, 83, 57, 97, 122, 105,
				73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80, 57, 56, 71, 69, 65, 85, 67, 55, 44,
				0, 0, 0, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80,
				118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4,
				5, 6, 7, 8, 4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78,
				68, 90, 83, 142, 11, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184,
				77, 80, 80, 39, 254, 173, 28, 169,
			},
			wantErr: false,
		},
		{
			name: "GetTransactionBytes:success-{without-signature}",
			args: args{
				transaction: &model.Transaction{
					Version:                       1,
					TransactionType:               2,
					Timestamp:                     1562806389280,
					SenderAccountAddressLength:    uint32(len([]byte("BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7"))),
					SenderAccountAddress:          "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
					RecipientAccountAddressLength: uint32(len([]byte("BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J"))),
					RecipientAccountAddress:       "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
					Fee:                           1000000,
					TransactionBodyLength:         8,
					TransactionBodyBytes:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
				},
				sign: false,
			},
			want: []byte{
				2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 44, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50, 83, 57, 97, 122,
				105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80, 57, 56, 71, 69, 65, 85,
				67, 55, 44, 0, 0, 0, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115,
				107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0,
				8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8,
			},
			wantErr: false,
		},
		{
			name: "GetTransactionBytes:fail-{sign:true, no signature}",
			args: args{
				transaction: &model.Transaction{
					TransactionType:               2,
					Version:                       1,
					Timestamp:                     1562806389280,
					SenderAccountAddressLength:    uint32(len([]byte("BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7"))),
					SenderAccountAddress:          "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
					RecipientAccountAddressLength: uint32(len([]byte("BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J"))),
					RecipientAccountAddress:       "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
					Fee:                           1000000,
					TransactionBodyLength:         8,
					TransactionBodyBytes:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
				},
				sign: true,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransactionBytes:success-{without recipient}",
			args: args{
				transaction: &model.Transaction{
					Version:                    1,
					TransactionType:            2,
					Timestamp:                  1562806389280,
					SenderAccountAddressLength: uint32(len([]byte("BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7"))),
					SenderAccountAddress:       "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
					Fee:                        1000000,
					TransactionBodyLength:      8,
					TransactionBodyBytes:       []byte{1, 2, 3, 4, 5, 6, 7, 8},
				},
				sign: false,
			},
			want: []byte{
				2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 44, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50, 83,
				57, 97, 122, 105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80, 57,
				56, 71, 69, 65, 85, 67, 55, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5,
				6, 7, 8,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTransactionBytes(tt.args.transaction, tt.args.sign)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTransactionBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTransactionBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTransactionBytes(t *testing.T) {
	type args struct {
		transactionBytes []byte
		sign             bool
	}
	tests := []struct {
		name    string
		args    args
		want    *model.Transaction
		wantErr bool
	}{
		{
			name: "ParseTransactionBytes:success",
			args: args{
				transactionBytes: []byte{
					2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 44, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50, 83, 57, 97, 122, 105,
					73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80, 57, 56, 71, 69, 65, 85, 67, 55, 44,
					0, 0, 0, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80,
					118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4,
					5, 6, 7, 8, 4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78,
					68, 90, 83, 142, 11, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184,
					77, 80, 80, 39, 254, 173, 28, 169,
				},
				sign: true,
			},
			want: &model.Transaction{
				ID:                            6829445217349326123,
				Version:                       1,
				TransactionType:               2,
				BlockID:                       0,
				Height:                        0,
				Timestamp:                     1562806389280,
				SenderAccountAddressLength:    44,
				SenderAccountAddress:          "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
				RecipientAccountAddressLength: 44,
				RecipientAccountAddress:       "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				Fee:                           1000000,
				TransactionHash: []byte{
					43, 213, 24, 193, 61, 16, 199, 94, 120, 255, 42, 44, 48, 36, 196, 250, 207, 25, 199, 203, 87,
					156, 26, 146, 232, 171, 179, 215, 12, 71, 124, 129,
				},
				TransactionBodyLength: 8,
				TransactionBodyBytes:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Signature: []byte{4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78,
					68, 90, 83, 142, 11, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184,
					77, 80, 80, 39, 254, 173, 28, 169},
			},
			wantErr: false,
		},
		{
			name: "ParseTransactionBytes:success-{without-signature}",
			args: args{
				transactionBytes: []byte{
					2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 44, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50, 83, 57, 97, 122,
					105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80, 57, 56, 71, 69, 65, 85,
					67, 55, 44, 0, 0, 0, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115,
					107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0,
					8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8,
				},
				sign: false,
			},
			want: &model.Transaction{
				ID:                            3602917061812791438,
				Version:                       1,
				TransactionType:               2,
				BlockID:                       0,
				Height:                        0,
				Timestamp:                     1562806389280,
				SenderAccountAddressLength:    uint32(len([]byte("BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7"))),
				SenderAccountAddress:          "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
				RecipientAccountAddressLength: uint32(len([]byte("BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J"))),
				RecipientAccountAddress:       "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				Fee:                           1000000,
				TransactionHash: []byte{
					142, 168, 139, 136, 250, 33, 0, 50, 23, 64, 67, 232, 186, 253, 134, 19, 59, 149, 71, 238, 156, 123, 77,
					203, 244, 160, 81, 179, 35, 24, 5, 81,
				},
				TransactionBodyLength: 8,
				TransactionBodyBytes:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			},
			wantErr: false,
		},
		{
			name: "ParseTransactionBytes:fail",
			args: args{
				transactionBytes: []byte{2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 44, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50, 83,
					57, 97, 122, 105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80, 57,
					56, 71, 69, 65, 85, 67, 55, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5,
					6, 7, 8},
				sign: true,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTransactionBytes(tt.args.transactionBytes, tt.args.sign)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTransactionBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTransactionBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTransactionID(t *testing.T) {
	type args struct {
		tx *model.Transaction
		ct contract.ChainType
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "GetTransactionID:success",
			args: args{
				tx: &model.Transaction{
					TransactionHash: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
					Signature:       make([]byte, 64),
				},
				ct: &chaintype.MainChain{},
			},
			wantErr: false,
			want:    72340172838076673,
		},
		{
			name: "GetTransactionID:fail",
			args: args{
				tx: &model.Transaction{
					TransactionHash: []byte{},
				},
				ct: &chaintype.MainChain{},
			},
			wantErr: true,
			want:    -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTransactionID(tt.args.tx.TransactionHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTransactionID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetTransactionID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateTransaction(t *testing.T) {
	type args struct {
		tx                  *model.Transaction
		queryExecutor       query.ExecutorInterface
		accountBalanceQuery query.AccountBalanceQueryInterface
		verifySignature     bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TestValidateTransaction:success",
			args: args{
				tx: buildTransaction(1562893303, "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
					"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE"),
				queryExecutor:       &mockQueryExecutorSuccess{},
				accountBalanceQuery: query.NewAccountBalanceQuery(),
				verifySignature:     false,
			},
		},
		{
			name: "ValidateTransaction:Genesis",
			args: args{
				tx: &model.Transaction{
					Height: 0,
				},
			},
		},
		{
			name: "ValidateTransaction:Fee<0",
			args: args{
				tx: &model.Transaction{
					Height: 1,
					Fee:    0,
				},
			},
			wantErr: true,
		},
		{
			name: "ValidateTransaction:SenderAddressEmpty",
			args: args{
				tx: &model.Transaction{
					Height: 1,
					Fee:    1,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateTransaction(tt.args.tx, tt.args.queryExecutor, tt.args.accountBalanceQuery,
				tt.args.verifySignature); (err != nil) != tt.wantErr {
				t.Errorf("ValidateTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func buildTransaction(timestamp int64, sender, recipient string) *model.Transaction {
	return &model.Transaction{
		Version:                       1,
		ID:                            2774809487,
		BlockID:                       1,
		Height:                        1,
		SenderAccountAddressLength:    uint32(len([]byte(sender))),
		SenderAccountAddress:          sender,
		RecipientAccountAddressLength: uint32(len([]byte(recipient))),
		RecipientAccountAddress:       recipient,
		TransactionType:               0,
		Fee:                           1,
		Timestamp:                     timestamp,
		TransactionHash:               make([]byte, 32),
		TransactionBodyLength:         0,
		TransactionBodyBytes:          make([]byte, 0),
		TransactionBody:               nil,
		Signature:                     make([]byte, 64),
	}
}

func TestReadAccountAddress(t *testing.T) {
	type args struct {
		accountType uint32
		buf         *bytes.Buffer
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "TestReadAccountAddress:defult",
			args: args{
				accountType: math.MaxUint32,
				buf: bytes.NewBuffer([]byte{2, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50, 83, 57, 97, 122,
					105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80, 57, 56, 71, 69, 65, 85, 67, 55,
					0, 0, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80,
					118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4,
					5, 6, 7, 8, 4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190,
					78, 68, 90, 83, 142, 11, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81,
					229, 184, 77, 80, 80, 39, 254, 173, 28, 169}),
			},
			want: []byte{
				2, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50, 83, 57, 97, 122, 105, 73, 76, 51, 99, 110,
				95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReadAccountAddress(tt.args.accountType, tt.args.buf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadAccountAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
