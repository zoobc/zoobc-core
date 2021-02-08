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
package feedbacksystem

import (
	"reflect"
	"sync"
	"testing"
)

func TestNewDummyFeedbackStrategy(t *testing.T) {
	tests := []struct {
		name string
		want *DummyFeedbackStrategy
	}{
		{
			name: "NewAntiSpamStrategy:success",
			want: &DummyFeedbackStrategy{
				FeedbackVars: map[string]interface{}{
					"tpsReceived":         0,
					"tpsReceivedTmp":      0,
					"tpsProcessed":        0,
					"tpsProcessedTmp":     0,
					"txReceived":          0,
					"txProcessed":         0,
					"P2PIncomingRequests": 0,
					"P2POutgoingRequests": 0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDummyFeedbackStrategy(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDummyFeedbackStrategy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDummyFeedbackStrategy_GetFeedbackVar(t *testing.T) {
	type fields struct {
		FeedbackVars     map[string]interface{}
		FeedbackVarsLock sync.RWMutex
	}
	type args struct {
		k string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		{
			name: "GetFeedbackVar",
			fields: fields{
				FeedbackVars: map[string]interface{}{
					"testVar": 10,
				},
			},
			args: args{
				k: "testVar",
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dfs := &DummyFeedbackStrategy{
				FeedbackVars:     tt.fields.FeedbackVars,
				FeedbackVarsLock: tt.fields.FeedbackVarsLock,
			}
			if got := dfs.GetFeedbackVar(tt.args.k); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFeedbackVar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDummyFeedbackStrategy_IncrementVarCount(t *testing.T) {
	type fields struct {
		FeedbackVars     map[string]interface{}
		FeedbackVarsLock sync.RWMutex
	}
	type args struct {
		k string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		{
			name: "IncrementVarCount:success",
			fields: fields{
				FeedbackVars: map[string]interface{}{
					"testVar": 10,
				},
			},
			args: args{
				k: "testVar",
			},
			want: 11,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dfs := &DummyFeedbackStrategy{
				FeedbackVars:     tt.fields.FeedbackVars,
				FeedbackVarsLock: tt.fields.FeedbackVarsLock,
			}
			if got := dfs.IncrementVarCount(tt.args.k); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IncrementVarCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDummyFeedbackStrategy_DecrementVarCount(t *testing.T) {
	type fields struct {
		FeedbackVars     map[string]interface{}
		FeedbackVarsLock sync.RWMutex
	}
	type args struct {
		k string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		{
			name: "DecrementVarCount:success",
			fields: fields{
				FeedbackVars: map[string]interface{}{
					"testVar": 10,
				},
			},
			args: args{
				k: "testVar",
			},
			want: 9,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dfs := &DummyFeedbackStrategy{
				FeedbackVars:     tt.fields.FeedbackVars,
				FeedbackVarsLock: tt.fields.FeedbackVarsLock,
			}
			if got := dfs.DecrementVarCount(tt.args.k); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecrementVarCount() = %v, want %v", got, tt.want)
			}
		})
	}
}
