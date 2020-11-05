package util

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestInitLogger(t *testing.T) {
	type args struct {
		path     string
		filename string
		levels   []string
		logOnCLI bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "wantError",
			args: args{
				path:     "",
				filename: "",
			},
			wantErr: true,
		},
		{
			name: "wantSuccess:FullLevels",
			args: args{
				path:     "./testdata/",
				filename: "test.log",
				levels:   []string{"info", "warn", "error", "fatal", "panic"},
				logOnCLI: true,
			},

			wantErr: false,
		},
		{
			name: "wantSuccess:NoLevels",
			args: args{
				path:     "./testdata/",
				filename: "test.log",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := InitLogger(tt.args.path, tt.args.filename, tt.args.levels, tt.args.logOnCLI); (got == nil) != tt.wantErr {
				t.Errorf("InitLogger() = %v, wantError %v", tt.name, tt.wantErr)
			}
		})
	}
}

func Test_hooker_Fire(t *testing.T) {
	type fields struct {
		Writer      *os.File
		EntryLevels []logrus.Level
	}
	type args struct {
		entry *logrus.Entry
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "wantFail:Write",
			fields: fields{},
			args: args{
				entry: &logrus.Entry{
					Logger: logrus.New(),
				},
			},
			wantErr: true,
		},
		{
			name:   "wantFail:EntryToString",
			fields: fields{},
			args: args{
				entry: &logrus.Entry{
					Logger: &logrus.Logger{
						Formatter: &logrus.TextFormatter{},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := hooker{
				Writer:      tt.fields.Writer,
				EntryLevels: tt.fields.EntryLevels,
			}
			if err := h.Fire(tt.args.entry); (err != nil) != tt.wantErr {
				t.Errorf("hooker.Fire() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
