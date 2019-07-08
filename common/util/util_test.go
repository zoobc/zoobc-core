package util

import (
	"testing"
)

func TestInitLogger(t *testing.T) {
	type args struct {
		path     string
		filename string
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
			name: "wantSuccess",
			args: args{
				path:     "./testdata/",
				filename: "test.log",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := InitLogger(tt.args.path, tt.args.filename); (got == nil) != tt.wantErr {
				t.Errorf("InitLogger() = %v, wantError %v", tt.name, tt.wantErr)
			}
		})
	}
}
