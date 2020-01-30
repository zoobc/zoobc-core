package query

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockEscrowQuery = NewEscrowTransactionQuery()
)

func TestEscrowTransactionQuery_InsertEscrowTransaction(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		escrow *model.Escrow
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantQStr string
		wantArgs []interface{}
	}{
		{
			name:   "wantSuccess",
			fields: fields(*mockEscrowQuery),
			args: args{
				escrow: &model.Escrow{
					ID:               0,
					SenderAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					RecipientAddress: "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
					ApproverAddress:  "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
					Amount:           10,
					Commission:       1,
					Timeout:          120,
					Status:           0,
					BlockHeight:      0,
					Latest:           true,
				},
			},
			wantArgs: []interface{}{
				int64(0),
				"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				"BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
				"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				int64(10),
				int64(1),
				uint64(120),
				model.EscrowStatus_Approved,
				uint32(0),
				true,
			},
			wantQStr: "INSERT INTO  (id,sender_address,recipient_address,approver_address,amount,commission,timeout,status,block_height,latest) " +
				"VALUES(? , ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et := &EscrowTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotQStr, gotArgs := et.InsertEscrowTransaction(tt.args.escrow)
			if gotQStr != tt.wantQStr {
				t.Errorf("InsertEscrowTransaction() gotQStr = %v, want %v", gotQStr, tt.wantQStr)
				return
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertEscrowTransaction() gotArgs = \n%v, want \n%v", gotArgs, tt.wantArgs)
			}
		})
	}
}
