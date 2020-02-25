package fee

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestBlockLifeTimeFeeModel_CalculateTxMinimumFee(t *testing.T) {
	type fields struct {
		blockPeriod       int64
		feePerBlockPeriod int64
	}
	type args struct {
		txBody model.TransactionBodyInterface
		escrow *model.Escrow
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "CalculateTxMinimumFee-1",
			fields: fields{
				blockPeriod:       5,
				feePerBlockPeriod: constant.OneZBC / 100,
			},
			args: args{
				txBody: nil,
				escrow: &model.Escrow{
					Timeout: 17,
				},
			},
			want:    4 * (constant.OneZBC / 100),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blt := &BlockLifeTimeFeeModel{
				blockPeriod:       tt.fields.blockPeriod,
				feePerBlockPeriod: tt.fields.feePerBlockPeriod,
			}
			got, err := blt.CalculateTxMinimumFee(tt.args.txBody, tt.args.escrow)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateTxMinimumFee() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CalculateTxMinimumFee() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBlockLifeTimeFeeModel(t *testing.T) {
	type args struct {
		blockPeriod       int64
		feePerBlockPeriod int64
	}
	tests := []struct {
		name string
		args args
		want *BlockLifeTimeFeeModel
	}{
		{
			name: "NewBlockLifeTimeFeeModel-Success",
			args: args{
				blockPeriod:       0,
				feePerBlockPeriod: 0,
			},
			want: &BlockLifeTimeFeeModel{
				feePerBlockPeriod: 0,
				blockPeriod:       0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockLifeTimeFeeModel(tt.args.blockPeriod, tt.args.feePerBlockPeriod); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockLifeTimeFeeModel() = %v, want %v", got, tt.want)
			}
		})
	}
}
