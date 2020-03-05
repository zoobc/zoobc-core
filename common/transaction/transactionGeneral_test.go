package transaction

import (
	"bytes"
	"database/sql"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
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
				2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 44, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50, 83, 57, 97, 122, 105, 73,
				76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80, 57, 56, 71, 69, 65, 85, 67, 55, 44, 0, 0,
				0, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66,
				57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8,
				4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9,
				76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78, 68, 90, 83, 142, 11, 4, 38, 68, 24, 230, 247, 88, 220, 119,
				124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169,
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
				2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 44, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79,
				50, 83, 57, 97, 122, 105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114,
				110, 88, 117, 80, 57, 56, 71, 69, 65, 85, 67, 55, 44, 0, 0, 0, 66, 67, 90, 75, 76, 118, 103, 85,
				89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57,
				106, 112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0,
				0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0,
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
				2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 44, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102,
				79, 50, 83, 57, 97, 122, 105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80,
				79, 114, 110, 88, 117, 80, 57, 56, 71, 69, 65, 85, 67, 55, 0, 0, 0, 0, 64, 66, 15, 0, 0, 0, 0,
				0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0,
			},
			wantErr: false,
		},
		{
			name: "Success:WithEscrow",
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
					Escrow: &model.Escrow{
						ApproverAddress: "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
						Commission:      24,
						Timeout:         100,
					},
					Signature: []byte{0, 0, 0, 0, 4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174,
						239, 46, 190, 78, 68, 90, 83, 142, 11, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56,
						139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				},
				sign: true,
			},
			want: []byte{
				2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 44, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50, 83, 57, 97, 122, 105, 73, 76, 51,
				99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80, 57, 56, 71, 69, 65, 85, 67, 55, 44, 0, 0, 0, 66, 67, 90,
				75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106,
				102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 44, 0, 0, 0, 66, 67, 90, 68,
				95, 86, 120, 102, 79, 50, 83, 57, 97, 122, 105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80,
				57, 56, 71, 69, 65, 85, 67, 55, 24, 0, 0, 0, 0, 0, 0, 0, 100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 38, 103, 73, 250, 169, 63,
				155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78, 68, 90, 83, 142, 11, 4, 38, 68, 24, 230, 247, 88,
				220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169,
			},
		},
		{
			name: "SuccessNoSigned:WithEscrow",
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
					Escrow: &model.Escrow{
						ApproverAddress: "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
						Commission:      24,
						Timeout:         100,
					},
					Signature: []byte{0, 0, 0, 0, 4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174,
						239, 46, 190, 78, 68, 90, 83, 142, 11, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56,
						139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				},
				sign: false,
			},
			want: []byte{
				2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 44, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50, 83, 57, 97, 122, 105, 73, 76, 51,
				99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80, 57, 56, 71, 69, 65, 85, 67, 55, 44, 0, 0, 0, 66, 67, 90,
				75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106,
				102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 44, 0, 0, 0, 66, 67, 90, 68,
				95, 86, 120, 102, 79, 50, 83, 57, 97, 122, 105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80,
				57, 56, 71, 69, 65, 85, 67, 55, 24, 0, 0, 0, 0, 0, 0, 0, 100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		{
			name: "EscrowApproval",
			args: args{
				transaction: &model.Transaction{
					Version:                 1,
					ID:                      1,
					BlockID:                 1,
					Height:                  1,
					SenderAccountAddress:    "GHI",
					RecipientAccountAddress: "",
					TransactionType:         4,
					Fee:                     1,
					Timestamp:               1562806389280,
					TransactionHash:         nil,
					TransactionBodyLength:   12,
					TransactionBodyBytes:    []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					TransactionIndex:        0,
					Signature:               nil,
					Escrow:                  nil,
				},
			},
			want: []byte{4, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 3, 0, 0, 0, 71, 72, 73, 0, 0, 0, 0, 1, 0, 0,
				0, 0, 0, 0, 0, 12, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name: "EscrowApproval:Signed",
			args: args{
				transaction: &model.Transaction{
					Version:                 1,
					ID:                      1,
					BlockID:                 1,
					Height:                  1,
					SenderAccountAddress:    "GHI",
					RecipientAccountAddress: "",
					TransactionType:         4,
					Fee:                     1,
					Timestamp:               1562806389280,
					TransactionHash:         nil,
					TransactionBodyLength:   12,
					TransactionBodyBytes:    []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					TransactionIndex:        0,
					Signature:               nil,
					Escrow:                  nil,
				},
			},
			want: []byte{4, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 3, 0, 0, 0, 71, 72, 73, 0, 0, 0, 0, 1, 0,
				0, 0, 0, 0, 0, 0, 12, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (&Util{}).GetTransactionBytes(tt.args.transaction, tt.args.sign)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTransactionBytes() error = \n%v, wantErr \n%v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				var byteStrArr []string
				for _, bt := range got {
					byteStrArr = append(byteStrArr, fmt.Sprintf("%v", bt))
				}
				t.Logf(strings.Join(byteStrArr, ", "))
				t.Errorf("GetTransactionBytes() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestParseTransactionBytes(t *testing.T) {
	transactionBytes, transactionHashed := GetFixturesForTransactionBytes(&model.Transaction{
		ID:                      670925173877174625,
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
		Escrow: &model.Escrow{
			ApproverAddress: "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
			Commission:      24,
			Timeout:         100,
		},
	}, true)
	approvalTX, approvalTXBytes := GetFixtureForSpecificTransaction(
		-5003742102621241788,
		12345678,
		"BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
		"",
		constant.EscrowApprovalBytesLength,
		model.TransactionType_ApprovalEscrowTransaction,
		&model.ApprovalEscrowTransactionBody{
			Approval:      model.EscrowApproval_Approve,
			TransactionID: 0,
		},
		false,
		true,
	)

	successWithoutSig, successWithoutSigHashed := GetFixturesForTransactionBytes(&model.Transaction{
		ID:                      670925173877174625,
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
		Escrow: &model.Escrow{
			ApproverAddress: "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
			Commission:      24,
			Timeout:         100,
		},
	}, false)
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
			name: "ParseTransactionBytes:withEscrow",
			args: args{
				transactionBytes: transactionBytes,
				sign:             true,
			},
			want: &model.Transaction{
				ID:                      8974217473277679238,
				Version:                 1,
				TransactionType:         2,
				BlockID:                 0,
				Height:                  0,
				Timestamp:               1562806389280,
				SenderAccountAddress:    "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
				RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				Fee:                     1000000,
				TransactionHash:         transactionHashed[:],
				TransactionBodyLength:   8,
				TransactionBodyBytes:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Signature: []byte{0, 0, 0, 0, 4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174,
					239, 46, 190, 78, 68, 90, 83, 142, 11, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56,
					139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				Escrow: &model.Escrow{
					ApproverAddress: "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
					Commission:      24,
					Timeout:         100,
				},
			},
			wantErr: false,
		},
		{
			name: "ParseTransactionBytes:success-{without-signature}",
			args: args{
				transactionBytes: successWithoutSig,
				sign:             false,
			},
			want: &model.Transaction{
				ID:                      4956766951297472907,
				Version:                 1,
				TransactionType:         2,
				BlockID:                 0,
				Height:                  0,
				Timestamp:               1562806389280,
				SenderAccountAddress:    "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
				RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				Fee:                     1000000,
				TransactionHash:         successWithoutSigHashed[:],
				TransactionBodyLength:   8,
				TransactionBodyBytes:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Escrow: &model.Escrow{
					ApproverAddress: "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
					Commission:      24,
					Timeout:         100,
				},
			},
			wantErr: false,
		},
		{
			name: "Ups",
			args: args{
				transactionBytes: approvalTXBytes,
				sign:             true,
			},
			want: approvalTX,
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
			got, err := (&Util{}).ParseTransactionBytes(tt.args.transactionBytes, tt.args.sign)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTransactionBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTransactionBytes() = \n%v, want \n%v", got, tt.want)
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
			if got := (&Util{}).ReadAccountAddress(tt.args.accountType, tt.args.buf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadAccountAddress() = %v, want %v", got, tt.want)
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
			got, err := (&Util{}).GetTransactionID(tt.args.tx.TransactionHash)
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

type mockQueryExecutorSuccess struct {
	query.Executor
}

func (*mockQueryExecutorSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mockRow := mock.NewRows(query.NewAccountBalanceQuery().Fields)
	mockRow.AddRow(
		"BCZ", 1, 10000, 10000, 0, 1,
	)
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(mockRow)
	row := db.QueryRow(qStr)
	return row, nil
}

func TestValidateTransaction(t *testing.T) {
	transactionUtil := &Util{}
	type args struct {
		tx                  *model.Transaction
		queryExecutor       query.ExecutorInterface
		accountBalanceQuery query.AccountBalanceQueryInterface
		verifySignature     bool
	}
	txEscrowValidate := GetFixturesForTransaction(
		1562893303,
		"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		true,
	)
	txBytesEscrow, _ := transactionUtil.GetTransactionBytes(txEscrowValidate, false)
	signatureEscrow := (&crypto.Signature{}).Sign(txBytesEscrow, constant.SignatureTypeDefault,
		"concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved")
	txEscrowValidate.Signature = signatureEscrow

	txValidate := GetFixturesForTransaction(
		1562893303,
		"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		false,
	)
	txBytes, _ := transactionUtil.GetTransactionBytes(txValidate, false)
	signature := (&crypto.Signature{}).Sign(txBytes, constant.SignatureTypeDefault,
		"concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved")
	txValidate.Signature = signature

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TestValidateTransaction:success",
			args: args{
				tx: GetFixturesForTransaction(
					time.Now().Unix()+int64(constant.TransactionTimeOffset)-1,
					"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
					"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
					false,
				),
				queryExecutor:       &mockQueryExecutorSuccess{},
				accountBalanceQuery: query.NewAccountBalanceQuery(),
				verifySignature:     false,
			},
			wantErr: false,
		},
		{
			name: "TestValidateTransactionWithEscrow:success",
			args: args{
				tx:                  txEscrowValidate,
				queryExecutor:       &mockQueryExecutorSuccess{},
				accountBalanceQuery: query.NewAccountBalanceQuery(),
				verifySignature:     true,
			},
		},
		{
			name: "TestValidateTransaction:success - verify signature",
			args: args{
				tx:                  txValidate,
				queryExecutor:       &mockQueryExecutorSuccess{},
				accountBalanceQuery: query.NewAccountBalanceQuery(),
				verifySignature:     true,
			},
			wantErr: false,
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
			if err := transactionUtil.ValidateTransaction(
				tt.args.tx,
				tt.args.queryExecutor,
				tt.args.accountBalanceQuery,
				tt.args.verifySignature,
			); (err != nil) != tt.wantErr {
				t.Errorf("ValidateTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUtil_GenerateMultiSigAddress(t *testing.T) {
	type args struct {
		info *model.MultiSignatureInfo
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "wantSuccess",
			args: args{info: &model.MultiSignatureInfo{
				MinimumSignatures: 2,
				Nonce:             12,
				Addresses: []string{
					"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					"BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
					"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				},
			}},
			want: "C1Jgm37Y-xW6ls1l9JW5XsRnsifX0CkWB4QTwZE9keVt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tu := &Util{}
			got, err := tu.GenerateMultiSigAddress(tt.args.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateMultiSigAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GenerateMultiSigAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}
