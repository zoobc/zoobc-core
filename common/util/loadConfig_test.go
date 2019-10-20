package util

import (
	"os"
	"testing"

	"github.com/spf13/viper"
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
			name: "MustSucceed",
			args: args{
				path:      "./resource",
				name:      "config",
				extension: "toml",
			},
			wantErr: false,
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

func TestOverrideConfigKeyArray(t *testing.T) {
	type args struct {
		envKey        string
		cfgFileKey    string
		envValue      string
		envValueCount int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "OverrideConfigKeyArray: success-{oneValue}",
			args: args{
				envKey:        "TEST_WELLKNOWN_PEERS",
				cfgFileKey:    "testWellKnownPeers",
				envValue:      "192.168.21.254:8001",
				envValueCount: 1,
			},
		},
		{
			name: "OverrideConfigKeyArray: success-{multiValue}",
			args: args{
				envKey:        "TEST_WELLKNOWN_PEERS",
				cfgFileKey:    "testWellKnownPeers",
				envValue:      "192.168.21.254:8001,192.168.21.253:8001,192.168.21.252:8001,192.168.21.251:8001,192.168.21.250:8001,192.168.21.249:8001",
				envValueCount: 6,
			},
		},
	}
	for _, tt := range tests {
		os.Setenv(tt.args.envKey, tt.args.envValue)
		t.Run(tt.name, func(t *testing.T) {
			OverrideConfigKeyArray(tt.args.envKey, tt.args.cfgFileKey)
			os.Unsetenv(tt.args.envKey)
			got := viper.GetStringSlice(tt.args.cfgFileKey)
			if len(got) == 0 || len(got) > tt.args.envValueCount {
				t.Errorf("error parsing env value")
			}
			if got[0] != "192.168.21.254:8001" {
				t.Errorf("config value diffs from env value")
			}
		})
	}
}
