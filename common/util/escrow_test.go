package util

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestValidateBasicEscrow(t *testing.T) {
	type args struct {
		tx *model.Transaction
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "wantFailed:InvalidEscrowObject",
			args: args{
				tx: &model.Transaction{},
			},
			wantErr: true,
		},
		{
			name: "wantFailed:ApproverAddressRequired",
			args: args{
				tx: &model.Transaction{
					Escrow: &model.Escrow{
						ApproverAddress: []byte{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "wantFailed:TimeoutHasPassed",
			args: args{
				tx: &model.Transaction{
					Escrow: &model.Escrow{
						ApproverAddress: []byte{1, 2, 3},
						Timeout:         99,
					},
					Timestamp: 100,
				},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			args: args{
				tx: &model.Transaction{
					Escrow: &model.Escrow{
						ApproverAddress: []byte{1, 2, 3},
						Timeout:         100,
					},
					Timestamp: 100,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateBasicEscrow(tt.args.tx); (err != nil) != tt.wantErr {
				t.Errorf("ValidateBasicEscrow() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrepareEscrowObjectForAction(t *testing.T) {
	type args struct {
		tx *model.Transaction
	}
	tests := []struct {
		name string
		args args
		want *model.Escrow
	}{
		{
			name: "wantSuccess:PreparingEscrowObjectSuccess",
			args: args{
				tx: &model.Transaction{
					ID:                      123,
					SenderAccountAddress:    []byte{1},
					RecipientAccountAddress: []byte{2},
					Height:                  3,
					Escrow: &model.Escrow{
						ApproverAddress: []byte{1, 2, 3},
						Commission:      0,
						Timeout:         99,
						Status:          model.EscrowStatus_Pending,
						Instruction:     "abc",
					},
				},
			},
			want: &model.Escrow{
				ID:               123,
				ApproverAddress:  []byte{1, 2, 3},
				Commission:       0,
				Timeout:          99,
				Status:           model.EscrowStatus_Pending,
				Instruction:      "abc",
				SenderAddress:    []byte{1},
				RecipientAddress: []byte{2},
				BlockHeight:      3,
				Latest:           true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PrepareEscrowObjectForAction(tt.args.tx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PrepareEscrowObjectForAction() = %v, want %v", got, tt.want)
			}
		})
	}
}
