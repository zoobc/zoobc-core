package util

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {

	type args struct {
		path      string
		name      string
		extension string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ConfigNotFound",
			args: args{
				path:      "./resource",
				name:      "config",
				extension: "toml",
			},
			wantErr: true,
		},
		{
			name: "MustError",
			args: args{
				path:      "./resource",
				name:      "fail",
				extension: "toml",
			},
			wantErr: true,
		},
		{
			name: "MustError:{len path, name, or extension < 1}",
			args: args{
				path:      "",
				name:      "",
				extension: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoadConfig(tt.args.path, tt.args.name, tt.args.extension); (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
