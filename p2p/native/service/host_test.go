package service

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

func changeMaxUnresolvedPeers(hostServiceInstance *HostService, newValue int32) {
	hostServiceInstance.MaxUnresolvedPeers = newValue
}

func changeMaxResolvedPeers(hostServiceInstance *HostService, newValue int32) {
	hostServiceInstance.MaxResolvedPeers = newValue
}

var goodHostInstance = &model.Host{
	Info: &model.Node{
		SharedAddress: "127.0.0.1",
		Address:       "127.0.0.1",
		Port:          8000,
	},
	KnownPeers: map[string]*model.Peer{
		"127.0.0.1:3000": {
			Info: &model.Node{
				SharedAddress: "127.0.0.1",
				Address:       "127.0.0.1",
				Port:          3000,
			},
		},
	},
	ResolvedPeers: map[string]*model.Peer{
		"127.0.0.1:3000": {
			Info: &model.Node{
				SharedAddress: "127.0.0.1",
				Address:       "127.0.0.1",
				Port:          3000,
			},
		},
	},
	UnresolvedPeers: map[string]*model.Peer{
		"127.0.0.1:3000": {
			Info: &model.Node{
				SharedAddress: "127.0.0.1",
				Address:       "127.0.0.1",
				Port:          3000,
			},
		},
	},
	BlacklistedPeers: map[string]*model.Peer{
		"127.0.0.1:3000": {
			Info: &model.Node{
				SharedAddress: "127.0.0.1",
				Address:       "127.0.0.1",
				Port:          3000,
			},
		},
	},
}

func TestCreateHostService(t *testing.T) {
	type args struct {
		hostInstance *model.Host
	}
	tests := []struct {
		name string
		args args
		want *HostService
	}{
		{
			name: "Host:CreateHostService",
			args: args{
				hostInstance: &model.Host{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3000,
					},
				},
			},
			want: &HostService{
				Host: &model.Host{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3000,
					},
				},
				MaxUnresolvedPeers: constant.MaxUnresolvedPeers,
				MaxResolvedPeers:   constant.MaxResolvedPeers,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateHostService(tt.args.hostInstance); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateHostService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetHostService(t *testing.T) {
	type args struct {
		hostInstance *model.Host
	}
	tests := []struct {
		name    string
		args    args
		want    *HostService
		wantErr bool
	}{
		{
			name: "Host:GetHostService success",
			args: args{
				hostInstance: &model.Host{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3001,
					},
				},
			},
			want: &HostService{
				Host: &model.Host{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3001,
					},
				},
				MaxUnresolvedPeers: constant.MaxUnresolvedPeers,
				MaxResolvedPeers:   constant.MaxResolvedPeers,
			},
			wantErr: false,
		},
		{
			name: "Host:GetHostService fails",
			args: args{
				hostInstance: nil,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CreateHostService(tt.args.hostInstance)
			got, err := GetHostService()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHostService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetHostService() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestGetResolvedPeers(t *testing.T) {
	type args struct {
		hostInstance *model.Host
	}
	tests := []struct {
		name string
		args args
		want map[string]*model.Peer
	}{
		{
			name: "Host:GetResolvedPeersTest",
			args: args{
				hostInstance: goodHostInstance,
			},
			want: map[string]*model.Peer{
				"127.0.0.1:3000": {
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3000,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := CreateHostService(tt.args.hostInstance)
			if got := hs.GetResolvedPeers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetResolvedPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAnyResolvedPeer(t *testing.T) {
	type args struct {
		hostInstance *model.Host
	}
	tests := []struct {
		name string
		args args
		want *model.Peer
	}{
		{
			name: "Host:GetAnyResolvedPeerTest",
			args: args{
				hostInstance: goodHostInstance,
			},
			want: &model.Peer{
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3000,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := CreateHostService(tt.args.hostInstance)
			if got := hs.GetAnyResolvedPeer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAnyResolvedPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddToResolvedPeer(t *testing.T) {
	type args struct {
		hostInstance *model.Host
		newPeer      *model.Peer
	}
	tests := []struct {
		name        string
		args        args
		wantContain *model.Peer
		wantErr     bool
	}{
		{
			name: "Host:AddToResolvedPeer success",
			args: args{
				hostInstance: goodHostInstance,
				newPeer: &model.Peer{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3001,
					},
				},
			},
			wantContain: &model.Peer{
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3001,
				},
			},
			wantErr: false,
		},
		{
			name: "Host:AddToResolvedPeer fails",
			args: args{
				hostInstance: nil,
				newPeer:      nil,
			},
			wantContain: nil,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := CreateHostService(tt.args.hostInstance)
			err := hs.AddToResolvedPeer(tt.args.newPeer)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddToResolvedPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				peers := hs.GetResolvedPeers()
				for _, peer := range peers {
					if reflect.DeepEqual(peer, tt.wantContain) {
						return
					}
				}
				t.Errorf("AddToResolvedPeer() = %v, want %v", peers, tt.wantContain)
			}
		})
	}
}

func TestRemoveResolvedPeer(t *testing.T) {
	type args struct {
		hostInstance *model.Host
		peerToRemove *model.Peer
	}
	tests := []struct {
		name           string
		args           args
		wantNotContain *model.Peer
		wantErr        bool
	}{
		{
			name: "Host:RemoveResolvedPeer success",
			args: args{
				hostInstance: goodHostInstance,
				peerToRemove: &model.Peer{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3000,
					},
				},
			},
			wantNotContain: &model.Peer{
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3000,
				},
			},
			wantErr: false,
		},
		{
			name: "Host:RemoveResolvedPeer fails",
			args: args{
				hostInstance: nil,
				peerToRemove: nil,
			},
			wantNotContain: nil,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := CreateHostService(tt.args.hostInstance)
			err := hs.RemoveResolvedPeer(tt.args.peerToRemove)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveResolvedPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				peers := hs.GetResolvedPeers()
				for _, peer := range peers {
					if reflect.DeepEqual(peer, tt.wantNotContain) {
						t.Errorf("RemoveResolvedPeer() = %v, want %v", peers, tt.wantNotContain)
					}
				}
			}
		})
	}
}

func TestGetUnresolvedPeers(t *testing.T) {
	type args struct {
		hostInstance *model.Host
	}
	tests := []struct {
		name string
		args args
		want map[string]*model.Peer
	}{
		{
			name: "Host:GetUnresolvedPeersTest",
			args: args{
				hostInstance: goodHostInstance,
			},
			want: map[string]*model.Peer{
				"127.0.0.1:3000": {
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3000,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := CreateHostService(tt.args.hostInstance)
			if got := hs.GetUnresolvedPeers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUnresolvedPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAnyUnresolvedPeer(t *testing.T) {
	type args struct {
		hostInstance *model.Host
	}
	tests := []struct {
		name string
		args args
		want *model.Peer
	}{
		{
			name: "Host:GetAnyUnresolvedPeerTest",
			args: args{
				hostInstance: goodHostInstance,
			},
			want: &model.Peer{
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3000,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := CreateHostService(tt.args.hostInstance)
			if got := hs.GetAnyUnresolvedPeer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAnyUnresolvedPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddToUnresolvedPeer(t *testing.T) {
	type args struct {
		hostInstance *model.Host
		newPeer      *model.Peer
	}
	tests := []struct {
		name        string
		args        args
		wantContain *model.Peer
		wantErr     bool
	}{
		{
			name: "Host:AddToUnresolvedPeer success",
			args: args{
				hostInstance: goodHostInstance,
				newPeer: &model.Peer{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3001,
					},
				},
			},
			wantContain: &model.Peer{
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3001,
				},
			},
			wantErr: false,
		},
		{
			name: "Host:AddToUnresolvedPeer fails",
			args: args{
				hostInstance: nil,
				newPeer:      nil,
			},
			wantContain: nil,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := CreateHostService(tt.args.hostInstance)
			err := hs.AddToUnresolvedPeer(tt.args.newPeer)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddToUnresolvedPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				peers := hs.GetUnresolvedPeers()
				for _, peer := range peers {
					if reflect.DeepEqual(peer, tt.wantContain) {
						return
					}
				}
				t.Errorf("AddToUnresolvedPeer() = %v, want %v", peers, tt.wantContain)
			}
		})
	}
}

func TestAddToUnresolvedPeers(t *testing.T) {
	type args struct {
		hostInstance       *model.Host
		newNode            *model.Node
		MaxUnresolvedPeers int32
		toForceAdd         bool
	}
	tests := []struct {
		name        string
		args        args
		wantContain *model.Peer
		wantErr     bool
	}{
		{
			name: "Host:AddToUnresolvedPeers success",
			args: args{
				hostInstance: goodHostInstance,
				newNode: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3001,
				},
				MaxUnresolvedPeers: 100,
				toForceAdd:         true,
			},
			wantContain: &model.Peer{
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3001,
				},
			},
			wantErr: false,
		},
		{
			name: "Host:AddToUnresolvedPeers fail",
			args: args{
				hostInstance: goodHostInstance,
				newNode: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3001,
				},
				MaxUnresolvedPeers: 1,
				toForceAdd:         false,
			},
			wantContain: nil,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := CreateHostService(tt.args.hostInstance)
			changeMaxUnresolvedPeers(hs, tt.args.MaxUnresolvedPeers)
			err := hs.AddToUnresolvedPeers([]*model.Node{tt.args.newNode}, tt.args.toForceAdd)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddToUnresolvedPeers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				peers := hs.GetUnresolvedPeers()
				for _, peer := range peers {
					if reflect.DeepEqual(peer, tt.wantContain) {
						return
					}
				}
				t.Errorf("AddToUnresolvedPeers() = %v, want %v", peers, tt.wantContain)
			}
		})
	}
}

func TestRemoveUnresolvedPeer(t *testing.T) {
	type args struct {
		hostInstance *model.Host
		peerToRemove *model.Peer
	}
	tests := []struct {
		name           string
		args           args
		wantNotContain *model.Peer
		wantErr        bool
	}{
		{
			name: "Host:RemoveUnresolvedPeer success",
			args: args{
				hostInstance: goodHostInstance,
				peerToRemove: &model.Peer{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3000,
					},
				},
			},
			wantNotContain: &model.Peer{
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3000,
				},
			},
			wantErr: false,
		},
		{
			name: "Host:RemoveUnresolvedPeer fails",
			args: args{
				hostInstance: nil,
				peerToRemove: nil,
			},
			wantNotContain: nil,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := CreateHostService(tt.args.hostInstance)
			err := hs.RemoveUnresolvedPeer(tt.args.peerToRemove)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveUnresolvedPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				peers := hs.GetUnresolvedPeers()
				for _, peer := range peers {
					if reflect.DeepEqual(peer, tt.wantNotContain) {
						t.Errorf("RemoveUnresolvedPeer() = %v, want %v", peers, tt.wantNotContain)
					}
				}
			}
		})
	}
}

func TestGetBlacklistedPeers(t *testing.T) {
	type args struct {
		hostInstance *model.Host
	}
	tests := []struct {
		name string
		args args
		want map[string]*model.Peer
	}{
		{
			name: "Host:GetBlacklistedPeersTest",
			args: args{
				hostInstance: goodHostInstance,
			},
			want: map[string]*model.Peer{
				"127.0.0.1:3000": {
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3000,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := CreateHostService(tt.args.hostInstance)
			if got := hs.GetBlacklistedPeers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlacklistedPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddToBlacklistedPeer(t *testing.T) {
	type args struct {
		hostInstance *model.Host
		newPeer      *model.Peer
	}
	tests := []struct {
		name        string
		args        args
		wantContain *model.Peer
		wantErr     bool
	}{
		{
			name: "Host:AddToBlacklistedPeer success",
			args: args{
				hostInstance: goodHostInstance,
				newPeer: &model.Peer{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3001,
					},
				},
			},
			wantContain: &model.Peer{
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3001,
				},
			},
			wantErr: false,
		},
		{
			name: "Host:AddToBlacklistedPeer fails",
			args: args{
				hostInstance: nil,
				newPeer:      nil,
			},
			wantContain: nil,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := CreateHostService(tt.args.hostInstance)
			err := hs.AddToBlacklistedPeer(tt.args.newPeer)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddToBlacklistedPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				peers := hs.GetBlacklistedPeers()
				for _, peer := range peers {
					if reflect.DeepEqual(peer, tt.wantContain) {
						return
					}
				}
				t.Errorf("AddToBlacklistedPeer() = %v, want %v", peers, tt.wantContain)
			}
		})
	}
}

func TestRemoveBlacklistedPeer(t *testing.T) {
	type args struct {
		hostInstance *model.Host
		peerToRemove *model.Peer
	}
	tests := []struct {
		name           string
		args           args
		wantNotContain *model.Peer
		wantErr        bool
	}{
		{
			name: "Host:RemoveBlacklistedPeer success",
			args: args{
				hostInstance: goodHostInstance,
				peerToRemove: &model.Peer{
					Info: &model.Node{
						SharedAddress: "127.0.0.1",
						Address:       "127.0.0.1",
						Port:          3000,
					},
				},
			},
			wantNotContain: &model.Peer{
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3000,
				},
			},
			wantErr: false,
		},
		{
			name: "Host:RemoveBlacklistedPeer fails",
			args: args{
				hostInstance: nil,
				peerToRemove: nil,
			},
			wantNotContain: nil,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := CreateHostService(tt.args.hostInstance)
			err := hs.RemoveBlacklistedPeer(tt.args.peerToRemove)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveBlacklistedPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				peers := hs.GetBlacklistedPeers()
				for _, peer := range peers {
					if reflect.DeepEqual(peer, tt.wantNotContain) {
						t.Errorf("RemoveBlacklistedPeer() = %v, want %v", peers, tt.wantNotContain)
					}
				}
			}
		})
	}
}

func TestGetAnyKnownPeer(t *testing.T) {
	type args struct {
		hostInstance *model.Host
	}
	tests := []struct {
		name string
		args args
		want *model.Peer
	}{
		{
			name: "Host:GetAnyKnownPeerTest",
			args: args{
				hostInstance: goodHostInstance,
			},
			want: &model.Peer{
				Info: &model.Node{
					SharedAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Port:          3000,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := CreateHostService(tt.args.hostInstance)
			if got := hs.GetAnyKnownPeer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAnyKnownPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetExceedMaxUnresolvedPeers(t *testing.T) {
	hs := CreateHostService(&model.Host{
		UnresolvedPeers: make(map[string]*model.Peer),
	})
	changeMaxUnresolvedPeers(hs, 1)

	var expectedResult, exceedMaxUnresolvedPeers int32

	expectedResult = int32(0)
	exceedMaxUnresolvedPeers = hs.GetExceedMaxUnresolvedPeers()
	if exceedMaxUnresolvedPeers != expectedResult {
		t.Errorf("GetExceedMaxUnresolvedPeers() = %v, want %v", exceedMaxUnresolvedPeers, expectedResult)
	}

	_ = hs.AddToUnresolvedPeer(&model.Peer{
		Info: &model.Node{
			SharedAddress: "127.0.0.1",
			Address:       "127.0.0.1",
			Port:          3000,
		},
	})

	expectedResult = int32(1)
	exceedMaxUnresolvedPeers = hs.GetExceedMaxUnresolvedPeers()
	if exceedMaxUnresolvedPeers != expectedResult {
		t.Errorf("GetExceedMaxUnresolvedPeers() = %v, want %v", exceedMaxUnresolvedPeers, expectedResult)
	}
}

func TestGetExceedMaxResolvedPeers(t *testing.T) {
	hs := CreateHostService(&model.Host{
		ResolvedPeers: make(map[string]*model.Peer),
	})
	changeMaxResolvedPeers(hs, 1)

	var expectedResult, exceedMaxResolvedPeers int32

	expectedResult = int32(0)
	exceedMaxResolvedPeers = hs.GetExceedMaxResolvedPeers()
	if exceedMaxResolvedPeers != expectedResult {
		t.Errorf("GetExceedMaxResolvedPeers() = %v, want %v", exceedMaxResolvedPeers, expectedResult)
	}

	_ = hs.AddToResolvedPeer(&model.Peer{
		Info: &model.Node{
			SharedAddress: "127.0.0.1",
			Address:       "127.0.0.1",
			Port:          3000,
		},
	})

	expectedResult = int32(1)
	exceedMaxResolvedPeers = hs.GetExceedMaxResolvedPeers()
	if exceedMaxResolvedPeers != expectedResult {
		t.Errorf("GetExceedMaxResolvedPeers() = %v, want %v", exceedMaxResolvedPeers, expectedResult)
	}
}
