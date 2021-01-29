// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.

package fee

import (
	"testing"
	"time"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestCalculateTxMinimumFee(t *testing.T) {
	uniformTimestamp := time.Now().Unix()

	type args struct {
		tx       *model.Transaction
		feeScale int64
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "wantError:escrowTimeouthasPassed",
			args: args{
				tx: &model.Transaction{
					Timestamp: uniformTimestamp,
					Escrow: &model.Escrow{
						Timeout: uniformTimestamp - 1000,
					},
				},
				feeScale: InitialFeeScale,
			},
			wantErr: true,
		},
		{
			name: "wantSuccess:minimumFee",
			args: args{
				tx:       &model.Transaction{},
				feeScale: InitialFeeScale,
			},
			want: InitialFeeScale,
		},
		{
			name: "wantSuccess:changedFeeScale",
			args: args{
				tx:       &model.Transaction{},
				feeScale: InitialFeeScale * 5,
			},
			want: InitialFeeScale * 5,
		},
		{
			name: "wantSuccess:withTxMesssage",
			args: args{
				tx: &model.Transaction{
					Message: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
				},
				feeScale: InitialFeeScale,
			},
			want: InitialFeeScale * 2,
		},
		{
			name: "wantSuccess:withEscrowInstruction",
			args: args{
				tx: &model.Transaction{
					Timestamp: uniformTimestamp,
					Escrow: &model.Escrow{
						Instruction: "1234567890",
						Timeout:     uniformTimestamp,
					},
				},
				feeScale: InitialFeeScale,
			},
			want: InitialFeeScale * 2,
		},
		{
			name: "wantSuccess:withTxMessage&EscrowInstruction",
			args: args{
				tx: &model.Transaction{
					Timestamp: uniformTimestamp,
					Message:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
					Escrow: &model.Escrow{
						Instruction: "1234567890",
						Timeout:     uniformTimestamp,
					},
				},
				feeScale: InitialFeeScale,
			},
			want: InitialFeeScale * 3,
		},
		{
			name: "wantSuccess:withTxMessage&EscrowInstruction&escrowTimeoutMoreThanEscrowLifetimeDivider",
			args: args{
				tx: &model.Transaction{
					Timestamp: uniformTimestamp,
					Message:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
					Escrow: &model.Escrow{
						Instruction: "1234567890",
						Timeout:     time.Unix(uniformTimestamp, 0).Add(EscrowLifetimeDivider*time.Hour).Unix() + 1,
					},
				},
				feeScale: InitialFeeScale,
			},
			want: InitialFeeScale * 3 * 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateTxMinimumFee(tt.args.tx, tt.args.feeScale)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateTxMinimumFee() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CalculateTxMinimumFee() = %v, want %v", got, tt.want)
			}
		})
	}
}
