package util

import (
	"bytes"
	"math"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

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
					TransactionType:         2,
					Version:                 1,
					Timestamp:               1562806389280,
					SenderAccountAddress:    "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
					RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
					Fee:                     1000000,
					TransactionBodyLength:   8,
					TransactionBodyBytes:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
					Signature: []byte{0, 0, 0, 0, 4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174,
						239, 46, 190, 78, 68, 90, 83, 142, 11, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56,
						139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				},
				sign: true,
			},
			want: []byte{
				2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 44, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50, 83, 57, 97, 122, 105, 73, 76,
				51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80, 57, 56, 71, 69, 65, 85, 67, 55, 44, 0, 0, 0, 66,
				67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106,
				112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 0, 0, 0,
				0, 4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78, 68, 90, 83,
				142, 11, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80,
				39, 254, 173, 28, 169,
			},
			wantErr: false,
		},
		{
			name: "GetTransactionBytes:success-{without-signature}",
			args: args{
				transaction: &model.Transaction{
					Version:                 1,
					TransactionType:         2,
					Timestamp:               1562806389280,
					SenderAccountAddress:    "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
					RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
					Fee:                     1000000,
					TransactionBodyLength:   8,
					TransactionBodyBytes:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
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
					TransactionType:         2,
					Version:                 1,
					Timestamp:               1562806389280,
					SenderAccountAddress:    "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
					RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
					Fee:                     1000000,
					TransactionBodyLength:   8,
					TransactionBodyBytes:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
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
					Version:               1,
					TransactionType:       2,
					Timestamp:             1562806389280,
					SenderAccountAddress:  "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
					Fee:                   1000000,
					TransactionBodyLength: 8,
					TransactionBodyBytes:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
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
					2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 44, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50, 83, 57, 97, 122, 105, 73, 76,
					51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80, 57, 56, 71, 69, 65, 85, 67, 55, 44, 0, 0, 0, 66,
					67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106,
					112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 0, 0, 0,
					0, 4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78, 68, 90, 83,
					142, 11, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80,
					39, 254, 173, 28, 169,
				},
				sign: true,
			},
			want: &model.Transaction{
				ID:                      -4993024615358961093,
				Version:                 1,
				TransactionType:         2,
				BlockID:                 0,
				Height:                  0,
				Timestamp:               1562806389280,
				SenderAccountAddress:    "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
				RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				Fee:                     1000000,
				TransactionHash: []byte{
					59, 106, 191, 6, 145, 54, 181, 186, 75, 93, 234, 139, 131, 96, 153, 252, 40, 245, 235, 132,
					187, 45, 245, 113, 210, 87, 23, 67, 157, 117, 41, 143,
				},
				TransactionBodyLength: 8,
				TransactionBodyBytes:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Signature: []byte{
					0, 0, 0, 0, 4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78,
					68, 90, 83, 142, 11, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184,
					77, 80, 80, 39, 254, 173, 28, 169,
				},
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
				ID:                      3602917061812791438,
				Version:                 1,
				TransactionType:         2,
				BlockID:                 0,
				Height:                  0,
				Timestamp:               1562806389280,
				SenderAccountAddress:    "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
				RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				Fee:                     1000000,
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
		ct chaintype.ChainType
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

func TestReadTransactionBytes(t *testing.T) {
	type args struct {
		buf    *bytes.Buffer
		nBytes int
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "ReadTransactionBytes:wrong-bytes",
			args: args{
				buf:    bytes.NewBuffer([]byte{1, 2}),
				nBytes: 4,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadTransactionBytes(tt.args.buf, tt.args.nBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadTransactionBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadTransactionBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeePerByteTransaction(t *testing.T) {
	type args struct {
		feeTransaction   int64
		transactionBytes []byte
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "wantSuccess",
			args: args{
				feeTransaction:   1,
				transactionBytes: []byte{1},
			},
			want: constant.OneFeePerByteTransaction,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FeePerByteTransaction(tt.args.feeTransaction, tt.args.transactionBytes); got != tt.want {
				t.Errorf("FeePerByteTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}
