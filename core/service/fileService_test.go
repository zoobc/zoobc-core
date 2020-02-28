package service

import (
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/ugorji/go/codec"
)

func TestFileService_ParseFileChunkHashes(t *testing.T) {
	type fields struct {
		Logger       *log.Logger
		h            codec.Handle
		snapshotPath string
	}
	type args struct {
		fileHashes []byte
		hashLength int
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		wantFileHashesAry [][]byte
		wantErr           bool
	}{
		{
			name: "ParseFileChunkHashes:success",
			args: args{
				hashLength: 32,
				fileHashes: make([]byte, 64),
			},
			wantFileHashesAry: [][]byte{
				make([]byte, 32),
				make([]byte, 32),
			},
		},
		{
			name: "ParseFileChunkHashes:fail-{InvalidHashesLength}",
			args: args{
				hashLength: 32,
				fileHashes: make([]byte, 65),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &FileService{
				Logger:       tt.fields.Logger,
				h:            tt.fields.h,
				snapshotPath: tt.fields.snapshotPath,
			}
			gotFileHashesAry, err := fs.ParseFileChunkHashes(tt.args.fileHashes, tt.args.hashLength)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileService.ParseFileChunkHashes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFileHashesAry, tt.wantFileHashesAry) {
				t.Errorf("FileService.ParseFileChunkHashes() = %v, want %v", gotFileHashesAry, tt.wantFileHashesAry)
			}
		})
	}
}
