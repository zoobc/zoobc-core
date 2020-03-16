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
