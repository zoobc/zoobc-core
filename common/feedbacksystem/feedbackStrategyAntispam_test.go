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
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"reflect"
	"sync"
	"testing"
)

func TestAntiSpamStrategy_DecrementVarCount(t *testing.T) {
	type fields struct {
		CPUPercentageSamples        []float64
		MemUsageSamples             []float64
		GoRoutineSamples            []int
		RunningCliP2PAPIRequests    []int
		RunningServerP2PAPIRequests []int
		FeedbackVars                map[string]interface{}
		FeedbackVarsLock            sync.RWMutex
		Logger                      *log.Logger
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
			ass := &AntiSpamStrategy{
				CPUPercentageSamples:        tt.fields.CPUPercentageSamples,
				MemUsageSamples:             tt.fields.MemUsageSamples,
				GoRoutineSamples:            tt.fields.GoRoutineSamples,
				RunningCliP2PAPIRequests:    tt.fields.RunningCliP2PAPIRequests,
				RunningServerP2PAPIRequests: tt.fields.RunningServerP2PAPIRequests,
				FeedbackVars:                tt.fields.FeedbackVars,
				FeedbackVarsLock:            tt.fields.FeedbackVarsLock,
				Logger:                      tt.fields.Logger,
			}
			if got := ass.DecrementVarCount(tt.args.k); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecrementVarCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAntiSpamStrategy_GetFeedbackVar(t *testing.T) {
	type fields struct {
		CPUPercentageSamples        []float64
		MemUsageSamples             []float64
		GoRoutineSamples            []int
		RunningCliP2PAPIRequests    []int
		RunningServerP2PAPIRequests []int
		FeedbackVars                map[string]interface{}
		FeedbackVarsLock            sync.RWMutex
		Logger                      *log.Logger
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
			ass := &AntiSpamStrategy{
				CPUPercentageSamples:        tt.fields.CPUPercentageSamples,
				MemUsageSamples:             tt.fields.MemUsageSamples,
				GoRoutineSamples:            tt.fields.GoRoutineSamples,
				RunningCliP2PAPIRequests:    tt.fields.RunningCliP2PAPIRequests,
				RunningServerP2PAPIRequests: tt.fields.RunningServerP2PAPIRequests,
				FeedbackVars:                tt.fields.FeedbackVars,
				FeedbackVarsLock:            tt.fields.FeedbackVarsLock,
				Logger:                      tt.fields.Logger,
			}
			if got := ass.GetFeedbackVar(tt.args.k); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFeedbackVar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAntiSpamStrategy_IncrementVarCount(t *testing.T) {
	type fields struct {
		CPUPercentageSamples        []float64
		MemUsageSamples             []float64
		GoRoutineSamples            []int
		RunningCliP2PAPIRequests    []int
		RunningServerP2PAPIRequests []int
		FeedbackVars                map[string]interface{}
		FeedbackVarsLock            sync.RWMutex
		Logger                      *log.Logger
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
			ass := &AntiSpamStrategy{
				CPUPercentageSamples:        tt.fields.CPUPercentageSamples,
				MemUsageSamples:             tt.fields.MemUsageSamples,
				GoRoutineSamples:            tt.fields.GoRoutineSamples,
				RunningCliP2PAPIRequests:    tt.fields.RunningCliP2PAPIRequests,
				RunningServerP2PAPIRequests: tt.fields.RunningServerP2PAPIRequests,
				FeedbackVars:                tt.fields.FeedbackVars,
				FeedbackVarsLock:            tt.fields.FeedbackVarsLock,
				Logger:                      tt.fields.Logger,
			}
			if got := ass.IncrementVarCount(tt.args.k); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IncrementVarCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAntiSpamStrategy_IsGoroutineLimitReached(t *testing.T) {
	type fields struct {
		CPUPercentageSamples        []float64
		MemUsageSamples             []float64
		GoRoutineSamples            []int
		RunningCliP2PAPIRequests    []int
		RunningServerP2PAPIRequests []int
		FeedbackVars                map[string]interface{}
		FeedbackVarsLock            sync.RWMutex
		Logger                      *log.Logger
	}
	type args struct {
		numSamples int
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantLimitReached bool
		wantLimitLevel   constant.FeedbackLimitLevel
	}{
		{
			name: "IsGoroutineLimitReached:success-{notEnoughSamples}",
			fields: fields{
				GoRoutineSamples: []int{
					10,
					10,
				},
			},
			args: args{
				numSamples: 3,
			},
			wantLimitLevel:   constant.FeedbackLimitNone,
			wantLimitReached: false,
		},
		{
			name: "IsGoroutineLimitReached:success-{noLimitReached}",
			fields: fields{
				GoRoutineSamples: []int{
					10,
					10,
					20,
					20,
					30,
					30,
				},
			},
			args: args{
				numSamples: 4,
			},
			wantLimitLevel:   constant.FeedbackLimitNone,
			wantLimitReached: false,
		},
		{
			name: "IsGoroutineLimitReached:success-{criticalLimitReached}",
			fields: fields{
				GoRoutineSamples: []int{
					constant.GoRoutineHardLimit,
					constant.GoRoutineHardLimit,
					constant.GoRoutineHardLimit + 100,
				},
				Logger: log.New(),
			},
			args: args{
				numSamples: 3,
			},
			wantLimitLevel:   constant.FeedbackLimitCritical,
			wantLimitReached: true,
		},
		{
			name: "IsGoroutineLimitReached:success-{highLimitReached}",
			fields: fields{
				GoRoutineSamples: []int{
					constant.GoRoutineHardLimit * constant.FeedbackLimitHighPerc / 100,
					constant.GoRoutineHardLimit * constant.FeedbackLimitHighPerc / 100,
				},
				Logger: log.New(),
			},
			args: args{
				numSamples: 2,
			},
			wantLimitLevel:   constant.FeedbackLimitHigh,
			wantLimitReached: true,
		},
		{
			name: "IsGoroutineLimitReached:success-{mediumLimitReached}",
			fields: fields{
				GoRoutineSamples: []int{
					constant.GoRoutineHardLimit * constant.FeedbackLimitMediumPerc / 100,
					constant.GoRoutineHardLimit * constant.FeedbackLimitMediumPerc / 100,
				},
				Logger: log.New(),
			},
			args: args{
				numSamples: 2,
			},
			wantLimitLevel:   constant.FeedbackLimitMedium,
			wantLimitReached: true,
		},
		{
			name: "IsGoroutineLimitReached:success-{mediumLimitReached}",
			fields: fields{
				GoRoutineSamples: []int{
					constant.GoRoutineHardLimit * constant.FeedbackLimitLowPerc / 100,
					constant.GoRoutineHardLimit * constant.FeedbackLimitLowPerc / 100,
				},
				Logger: log.New(),
			},
			args: args{
				numSamples: 2,
			},
			wantLimitLevel:   constant.FeedbackLimitLow,
			wantLimitReached: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ass := &AntiSpamStrategy{
				CPUPercentageSamples:        tt.fields.CPUPercentageSamples,
				MemUsageSamples:             tt.fields.MemUsageSamples,
				GoRoutineSamples:            tt.fields.GoRoutineSamples,
				RunningCliP2PAPIRequests:    tt.fields.RunningCliP2PAPIRequests,
				RunningServerP2PAPIRequests: tt.fields.RunningServerP2PAPIRequests,
				FeedbackVars:                tt.fields.FeedbackVars,
				FeedbackVarsLock:            tt.fields.FeedbackVarsLock,
				Logger:                      tt.fields.Logger,
			}
			gotLimitReached, gotLimitLevel := ass.IsGoroutineLimitReached(tt.args.numSamples)
			if gotLimitReached != tt.wantLimitReached {
				t.Errorf("IsGoroutineLimitReached() gotLimitReached = %v, want %v", gotLimitReached, tt.wantLimitReached)
			}
			if gotLimitLevel != tt.wantLimitLevel {
				t.Errorf("IsGoroutineLimitReached() gotLimitLevel = %v, want %v", gotLimitLevel, tt.wantLimitLevel)
			}
		})
	}
}

func TestAntiSpamStrategy_IsP2PRequestLimitReached(t *testing.T) {
	type fields struct {
		CPUPercentageSamples        []float64
		MemUsageSamples             []float64
		GoRoutineSamples            []int
		RunningCliP2PAPIRequests    []int
		RunningServerP2PAPIRequests []int
		FeedbackVars                map[string]interface{}
		FeedbackVarsLock            sync.RWMutex
		Logger                      *log.Logger
		P2PRequestsLimit            int
		CPULimit                    int
	}
	type args struct {
		numSamples int
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantLimitReached bool
		wantLimitLevel   constant.FeedbackLimitLevel
	}{
		{
			name: "IsP2PRequestLimitReached:success-{notEnoughSamples}",
			fields: fields{
				RunningServerP2PAPIRequests: []int{
					10,
					10,
				},
				P2PRequestsLimit: constant.P2PRequestHardLimit,
			},
			args: args{
				numSamples: 3,
			},
			wantLimitLevel:   constant.FeedbackLimitNone,
			wantLimitReached: false,
		},
		{
			name: "IsP2PRequestLimitReached:success-{noLimitReached}",
			fields: fields{
				RunningServerP2PAPIRequests: []int{
					10,
					10,
					20,
					20,
					30,
					30,
				},
				P2PRequestsLimit: constant.P2PRequestHardLimit,
			},
			args: args{
				numSamples: 4,
			},
			wantLimitLevel:   constant.FeedbackLimitNone,
			wantLimitReached: false,
		},
		{
			name: "IsP2PRequestLimitReached:success-{criticalLimitReached}",
			fields: fields{
				RunningServerP2PAPIRequests: []int{
					constant.P2PRequestHardLimit,
					constant.P2PRequestHardLimit,
					constant.P2PRequestHardLimit + 100,
				},
				Logger:           log.New(),
				P2PRequestsLimit: constant.P2PRequestHardLimit,
			},
			args: args{
				numSamples: 3,
			},
			wantLimitLevel:   constant.FeedbackLimitCritical,
			wantLimitReached: true,
		},
		{
			name: "IsP2PRequestLimitReached:success-{highLimitReached}",
			fields: fields{
				RunningServerP2PAPIRequests: []int{
					constant.P2PRequestHardLimit * constant.FeedbackLimitHighPerc / 100,
					constant.P2PRequestHardLimit * constant.FeedbackLimitHighPerc / 100,
				},
				Logger:           log.New(),
				P2PRequestsLimit: constant.P2PRequestHardLimit,
			},
			args: args{
				numSamples: 2,
			},
			wantLimitLevel:   constant.FeedbackLimitHigh,
			wantLimitReached: true,
		},
		{
			name: "IsP2PRequestLimitReached:success-{mediumLimitReached}",
			fields: fields{
				RunningServerP2PAPIRequests: []int{
					constant.P2PRequestHardLimit * constant.FeedbackLimitMediumPerc / 100,
					constant.P2PRequestHardLimit * constant.FeedbackLimitMediumPerc / 100,
				},
				Logger:           log.New(),
				P2PRequestsLimit: constant.P2PRequestHardLimit,
			},
			args: args{
				numSamples: 2,
			},
			wantLimitLevel:   constant.FeedbackLimitMedium,
			wantLimitReached: true,
		},
		{
			name: "IsP2PRequestLimitReached:success-{mediumLimitReached}",
			fields: fields{
				RunningServerP2PAPIRequests: []int{
					constant.P2PRequestHardLimit * constant.FeedbackLimitLowPerc / 100,
					constant.P2PRequestHardLimit * constant.FeedbackLimitLowPerc / 100,
				},
				Logger:           log.New(),
				P2PRequestsLimit: constant.P2PRequestHardLimit,
			},
			args: args{
				numSamples: 2,
			},
			wantLimitLevel:   constant.FeedbackLimitLow,
			wantLimitReached: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ass := &AntiSpamStrategy{
				CPUPercentageSamples:        tt.fields.CPUPercentageSamples,
				MemUsageSamples:             tt.fields.MemUsageSamples,
				GoRoutineSamples:            tt.fields.GoRoutineSamples,
				RunningCliP2PAPIRequests:    tt.fields.RunningCliP2PAPIRequests,
				RunningServerP2PAPIRequests: tt.fields.RunningServerP2PAPIRequests,
				FeedbackVars:                tt.fields.FeedbackVars,
				FeedbackVarsLock:            tt.fields.FeedbackVarsLock,
				Logger:                      tt.fields.Logger,
				P2pRequestLimit:             tt.fields.P2PRequestsLimit,
				CPUPercentageLimit:          tt.fields.CPULimit,
			}
			gotLimitReached, gotLimitLevel := ass.IsP2PRequestLimitReached(tt.args.numSamples)
			if gotLimitReached != tt.wantLimitReached {
				t.Errorf("IsP2PRequestLimitReached() gotLimitReached = %v, want %v", gotLimitReached, tt.wantLimitReached)
			}
			if gotLimitLevel != tt.wantLimitLevel {
				t.Errorf("IsP2PRequestLimitReached() gotLimitLevel = %v, want %v", gotLimitLevel, tt.wantLimitLevel)
			}
		})
	}
}

func TestNewAntiSpamStrategy(t *testing.T) {
	type args struct {
		logger *log.Logger
	}
	tests := []struct {
		name string
		args args
		want *AntiSpamStrategy
	}{
		{
			name: "NewAntiSpamStrategy:success",
			args: args{
				logger: nil,
			},
			want: &AntiSpamStrategy{
				Logger:                      nil,
				CPUPercentageSamples:        make([]float64, 0, constant.FeedbackTotalSamples),
				MemUsageSamples:             make([]float64, 0, constant.FeedbackTotalSamples),
				GoRoutineSamples:            make([]int, 0, constant.FeedbackTotalSamples),
				RunningServerP2PAPIRequests: make([]int, 0, constant.FeedbackTotalSamples),
				RunningCliP2PAPIRequests:    make([]int, 0, constant.FeedbackTotalSamples),
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
				CPUPercentageLimit: 10,
				P2pRequestLimit:    11,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAntiSpamStrategy(tt.args.logger, 10, 11); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAntiSpamStrategy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAntiSpamStrategy_IsCPULimitReached(t *testing.T) {
	type fields struct {
		CPUPercentageSamples        []float64
		MemUsageSamples             []float64
		GoRoutineSamples            []int
		RunningCliP2PAPIRequests    []int
		RunningServerP2PAPIRequests []int
		FeedbackVars                map[string]interface{}
		FeedbackVarsLock            sync.RWMutex
		CPUPercentageLimit          int
		P2PRequestLimit             int
		Logger                      *log.Logger
	}
	type args struct {
		numSamples int
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantLimitReached bool
		wantLimitLevel   constant.FeedbackLimitLevel
	}{
		{
			name: "IsP2PRequestLimitReached:success-{notEnoughSamples}",
			fields: fields{
				CPUPercentageSamples: []float64{
					10,
					10,
				},
				CPUPercentageLimit: constant.FeedbackLimitCPUPercentage,
			},
			args: args{
				numSamples: 3,
			},
			wantLimitLevel:   constant.FeedbackLimitNone,
			wantLimitReached: false,
		},
		{
			name: "IsP2PRequestLimitReached:success-{noLimitReached}",
			fields: fields{
				CPUPercentageSamples: []float64{
					10,
					10,
					20,
					20,
					30,
					30,
				},
				CPUPercentageLimit: constant.FeedbackLimitCPUPercentage,
			},
			args: args{
				numSamples: 4,
			},
			wantLimitLevel:   constant.FeedbackLimitNone,
			wantLimitReached: false,
		},
		{
			name: "IsP2PRequestLimitReached:success-{criticalLimitReached}",
			fields: fields{
				CPUPercentageSamples: []float64{
					constant.FeedbackLimitCPUPercentage,
					constant.FeedbackLimitCPUPercentage,
					constant.FeedbackLimitCPUPercentage + 2,
				},
				Logger:             log.New(),
				CPUPercentageLimit: constant.FeedbackLimitCPUPercentage,
			},
			args: args{
				numSamples: 3,
			},
			wantLimitLevel:   constant.FeedbackLimitCritical,
			wantLimitReached: true,
		},
		{
			name: "IsP2PRequestLimitReached:success-{highLimitReached}",
			fields: fields{
				CPUPercentageSamples: []float64{
					constant.FeedbackLimitCPUPercentage * constant.FeedbackLimitHighPerc / 100,
					constant.FeedbackLimitCPUPercentage * constant.FeedbackLimitHighPerc / 100,
				},
				Logger:             log.New(),
				CPUPercentageLimit: constant.FeedbackLimitCPUPercentage,
			},
			args: args{
				numSamples: 2,
			},
			wantLimitLevel:   constant.FeedbackLimitHigh,
			wantLimitReached: true,
		},
		{
			name: "IsP2PRequestLimitReached:success-{mediumLimitReached}",
			fields: fields{
				CPUPercentageSamples: []float64{
					constant.FeedbackLimitCPUPercentage * constant.FeedbackLimitMediumPerc / 100,
					constant.FeedbackLimitCPUPercentage * constant.FeedbackLimitMediumPerc / 100,
				},
				Logger:             log.New(),
				CPUPercentageLimit: constant.FeedbackLimitCPUPercentage,
			},
			args: args{
				numSamples: 2,
			},
			wantLimitLevel:   constant.FeedbackLimitMedium,
			wantLimitReached: true,
		},
		{
			name: "IsP2PRequestLimitReached:success-{mediumLimitReached}",
			fields: fields{
				CPUPercentageSamples: []float64{
					constant.FeedbackLimitCPUPercentage * constant.FeedbackLimitLowPerc / 100,
					constant.FeedbackLimitCPUPercentage * constant.FeedbackLimitLowPerc / 100,
				},
				Logger:             log.New(),
				CPUPercentageLimit: constant.FeedbackLimitCPUPercentage,
			},
			args: args{
				numSamples: 2,
			},
			wantLimitLevel:   constant.FeedbackLimitLow,
			wantLimitReached: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ass := &AntiSpamStrategy{
				CPUPercentageSamples:        tt.fields.CPUPercentageSamples,
				MemUsageSamples:             tt.fields.MemUsageSamples,
				GoRoutineSamples:            tt.fields.GoRoutineSamples,
				RunningCliP2PAPIRequests:    tt.fields.RunningCliP2PAPIRequests,
				RunningServerP2PAPIRequests: tt.fields.RunningServerP2PAPIRequests,
				FeedbackVars:                tt.fields.FeedbackVars,
				FeedbackVarsLock:            tt.fields.FeedbackVarsLock,
				CPUPercentageLimit:          tt.fields.CPUPercentageLimit,
				P2pRequestLimit:             tt.fields.P2PRequestLimit,
				Logger:                      tt.fields.Logger,
			}
			gotLimitReached, gotLimitLevel := ass.IsCPULimitReached(tt.args.numSamples)
			if gotLimitReached != tt.wantLimitReached {
				t.Errorf("IsCPULimitReached() gotLimitReached = %v, want %v", gotLimitReached, tt.wantLimitReached)
			}
			if gotLimitLevel != tt.wantLimitLevel {
				t.Errorf("IsCPULimitReached() gotLimitLevel = %v, want %v", gotLimitLevel, tt.wantLimitLevel)
			}
		})
	}
}
