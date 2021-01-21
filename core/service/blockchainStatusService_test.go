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
package service

import (
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
)

func TestNewBlockchainStatusService(t *testing.T) {
	type args struct {
		lockSmithing bool
		logger       *log.Logger
	}
	logTest := log.New()
	tests := []struct {
		name string
		args args
		want *BlockchainStatusService
	}{
		{
			name: "NewBlockchainStatusService",
			args: args{
				lockSmithing: true,
				logger:       logTest,
			},
			want: &BlockchainStatusService{
				Logger: logTest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewBlockchainStatusService(tt.args.lockSmithing, tt.args.logger)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockchainStatusService() = %v, want %v", got, tt.want)
			}
			if got.IsSmithingLocked() != tt.args.lockSmithing {
				t.Errorf("NewBlockchainStatusService() lockSmithing = %v, want %v", got.IsSmithingLocked(),
					tt.args.lockSmithing)
			}
		})
	}
}

func TestBlockchainStatusService_SetFirstDownloadFinished(t *testing.T) {
	type fields struct {
		Logger *log.Logger
	}
	type args struct {
		ct       chaintype.ChainType
		finished bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "SetFirstDownloadFinished",
			args: args{
				ct:       &chaintype.MainChain{},
				finished: true,
			},
			fields: fields{
				Logger: log.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btss := &BlockchainStatusService{
				Logger: tt.fields.Logger,
			}
			// test map concurrency r/w
			for i := 0; i < 10; i++ {
				go btss.SetFirstDownloadFinished(tt.args.ct, tt.args.finished)
				go btss.IsFirstDownloadFinished(tt.args.ct)
			}
		})
	}
}

func TestBlockchainStatusService_IsFirstDownloadFinished(t *testing.T) {
	type fields struct {
		Logger *log.Logger
	}
	type args struct {
		ct     chaintype.ChainType
		setVal bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "IsFirstDownloadFinished",
			fields: fields{
				Logger: log.New(),
			},
			args: args{
				ct:     &chaintype.MainChain{},
				setVal: true,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btss := &BlockchainStatusService{
				Logger: tt.fields.Logger,
			}
			btss.SetFirstDownloadFinished(tt.args.ct, tt.args.setVal)
			if got := btss.IsFirstDownloadFinished(tt.args.ct); got != tt.want {
				t.Errorf("BlockchainStatusService.IsFirstDownloadFinished() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchainStatusService_SetIsDownloading(t *testing.T) {
	type fields struct {
		Logger *log.Logger
	}
	type args struct {
		ct          chaintype.ChainType
		downloading bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "SetIsDownloading",
			args: args{
				ct:          &chaintype.MainChain{},
				downloading: true,
			},
			fields: fields{
				Logger: log.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btss := &BlockchainStatusService{
				Logger: tt.fields.Logger,
			}
			// test map concurrency r/w
			for i := 0; i < 10; i++ {
				go btss.SetIsDownloading(tt.args.ct, tt.args.downloading)
				go btss.IsDownloading(tt.args.ct)
			}
		})
	}
}

func TestBlockchainStatusService_IsDownloading(t *testing.T) {
	type fields struct {
		Logger *log.Logger
	}
	type args struct {
		ct     chaintype.ChainType
		setVal bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "IsDownloading",
			fields: fields{
				Logger: log.New(),
			},
			args: args{
				ct:     &chaintype.MainChain{},
				setVal: true,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btss := &BlockchainStatusService{
				Logger: tt.fields.Logger,
			}
			btss.SetIsDownloading(tt.args.ct, tt.args.setVal)
			if got := btss.IsDownloading(tt.args.ct); got != tt.want {
				t.Errorf("BlockchainStatusService.IsDownloading() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchainStatusService_SetIsSmithingLocked(t *testing.T) {
	type fields struct {
		Logger *log.Logger
	}
	type args struct {
		smithingLocked bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "SetIsDownloading",
			args: args{
				smithingLocked: true,
			},
			fields: fields{
				Logger: log.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btss := &BlockchainStatusService{
				Logger: tt.fields.Logger,
			}
			btss.SetIsSmithingLocked(tt.args.smithingLocked)
		})
	}
}

func TestBlockchainStatusService_IsSmithingLocked(t *testing.T) {
	type fields struct {
		Logger *log.Logger
	}
	type args struct {
		setVal bool
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		want   bool
	}{
		{
			name: "IsDownloading",
			fields: fields{
				Logger: log.New(),
			},
			args: args{
				setVal: true,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btss := &BlockchainStatusService{
				Logger: tt.fields.Logger,
			}
			btss.SetIsSmithingLocked(tt.args.setVal)
			if got := btss.IsSmithingLocked(); got != tt.want {
				t.Errorf("BlockchainStatusService.IsSmithingLocked() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchainStatusService_SetIsSmithing(t *testing.T) {
	type fields struct {
		Logger *log.Logger
	}
	type args struct {
		ct       chaintype.ChainType
		smithing bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "SetIsDownloading",
			args: args{
				ct:       &chaintype.MainChain{},
				smithing: true,
			},
			fields: fields{
				Logger: log.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btss := &BlockchainStatusService{
				Logger: tt.fields.Logger,
			}
			// test map concurrency r/w
			for i := 0; i < 10; i++ {
				go btss.SetIsSmithing(tt.args.ct, tt.args.smithing)
				go btss.IsSmithing(tt.args.ct)
			}
		})
	}
}

func TestBlockchainStatusService_IsSmithing(t *testing.T) {
	type fields struct {
		Logger *log.Logger
	}
	type args struct {
		ct     chaintype.ChainType
		setVal bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "IsSmithing",
			fields: fields{
				Logger: log.New(),
			},
			args: args{
				ct:     &chaintype.MainChain{},
				setVal: true,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btss := &BlockchainStatusService{
				Logger: tt.fields.Logger,
			}
			btss.SetIsSmithing(tt.args.ct, tt.args.setVal)
			if got := btss.IsSmithing(tt.args.ct); got != tt.want {
				t.Errorf("BlockchainStatusService.IsSmithing() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockchainStatusService_SetIsDownloadingSnapshot(t *testing.T) {
	type fields struct {
		Logger *log.Logger
	}
	type args struct {
		ct                  chaintype.ChainType
		downloadingSnapshot bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "SetIsDownloading",
			args: args{
				ct:                  &chaintype.MainChain{},
				downloadingSnapshot: true,
			},
			fields: fields{
				Logger: log.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btss := &BlockchainStatusService{
				Logger: tt.fields.Logger,
			}
			// test map concurrency r/w
			for i := 0; i < 10; i++ {
				go btss.SetIsDownloadingSnapshot(tt.args.ct, tt.args.downloadingSnapshot)
				go btss.IsDownloadingSnapshot(tt.args.ct)
			}
		})
	}
}

func TestBlockchainStatusService_IsDownloadingSnapshot(t *testing.T) {
	type fields struct {
		Logger *log.Logger
	}
	type args struct {
		ct     chaintype.ChainType
		setVal bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "IsSmithing",
			fields: fields{
				Logger: log.New(),
			},
			args: args{
				ct:     &chaintype.MainChain{},
				setVal: true,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btss := &BlockchainStatusService{
				Logger: tt.fields.Logger,
			}
			btss.SetIsDownloadingSnapshot(tt.args.ct, tt.args.setVal)
			if got := btss.IsDownloadingSnapshot(tt.args.ct); got != tt.want {
				t.Errorf("BlockchainStatusService.IsDownloadingSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}
