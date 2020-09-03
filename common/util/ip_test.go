package util

import (
	"net"
	"testing"
)

func TestGetPublicIP(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name: "WantSuccess",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipu := &IPUtil{}
			got, err := ipu.GetPublicIP()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPublicIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if ipu.IsPublicIP(got) {
				t.Errorf("GetPublicIP() got = %v ", got)
			}
		})
	}
}

func TestIsPublicIP(t *testing.T) {
	type args struct {
		IP net.IP
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "WantPublicIP",
			args: args{
				IP: net.ParseIP("172.104.34.10"),
			},
			want: true,
		},
		{
			name: "WantPrivateIP",
			args: args{
				IP: net.ParseIP("192.168.10.1"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipu := &IPUtil{}
			if got := ipu.IsPublicIP(&tt.args.IP); got != tt.want {
				t.Errorf("IsPublicIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDomain(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "WantDomain",
			args: args{
				address: "zoobc.com",
			},
			want: true,
		},
		{
			name: "WantIP",
			args: args{
				address: "172.104.34.10",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipu := &IPUtil{}
			if got := ipu.IsDomain(tt.args.address); got != tt.want {
				t.Errorf("IsDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPublicIPDYNDNS(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name: "WantSuccess",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipu := &IPUtil{}
			got, err := ipu.GetPublicIPDYNDNS()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPublicIPDYNDNS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !ipu.IsPublicIP(got) { // perhaps is public ip
				t.Errorf("GetPublicIPDYNDNS() got = %v", got)
			}
		})
	}
}
