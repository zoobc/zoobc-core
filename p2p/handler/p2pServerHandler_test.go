package handler

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
	service2 "github.com/zoobc/zoobc-core/p2p/service"
)

type (
	mockGetMorePeersError struct {
		service2.P2PServerServiceInterface
	}
	mockGetMorePeersSuccess struct {
		service2.P2PServerServiceInterface
	}
)

func (*mockGetMorePeersError) GetMorePeers(ctx context.Context, req *model.Empty) ([]*model.Node, error) {
	return nil, errors.New("Error GetMorePeers")
}

func (*mockGetMorePeersSuccess) GetMorePeers(ctx context.Context, req *model.Empty) ([]*model.Node, error) {
	return []*model.Node{}, nil
}

func TestP2PServerHandler_GetMorePeers(t *testing.T) {
	type fields struct {
		Service service2.P2PServerServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.Empty
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetMorePeersResponse
		wantErr bool
	}{
		{
			name: "GetMorePeers:Error",
			fields: fields{
				Service: &mockGetMorePeersError{},
			},
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetMorePeers:Success",
			fields: fields{
				Service: &mockGetMorePeersSuccess{},
			},
			args: args{},
			want: &model.GetMorePeersResponse{
				Peers: []*model.Node{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &P2PServerHandler{
				Service: tt.fields.Service,
			}
			got, err := ss.GetMorePeers(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerHandler.GetMorePeers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerHandler.GetMorePeers() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockSendPeersError struct {
		service2.P2PServerServiceInterface
	}
	mockSendPeersSuccess struct {
		service2.P2PServerServiceInterface
	}
)

func (*mockSendPeersError) SendPeers(ctx context.Context, peers []*model.Node) (*model.Empty, error) {
	return nil, errors.New("Error SendPeers")
}
func (*mockSendPeersSuccess) SendPeers(ctx context.Context, peers []*model.Node) (*model.Empty, error) {
	return &model.Empty{}, nil
}

func TestP2PServerHandler_SendPeers(t *testing.T) {
	type fields struct {
		Service service2.P2PServerServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.SendPeersRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Empty
		wantErr bool
	}{
		{
			name: "SendPeers:PeersIsNil",
			args: args{
				req: &model.SendPeersRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "SendPeers:Error",
			args: args{
				req: &model.SendPeersRequest{
					Peers: []*model.Node{},
				},
			},
			fields: fields{
				Service: &mockSendPeersError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "SendPeers:Success",
			args: args{
				req: &model.SendPeersRequest{
					Peers: []*model.Node{},
				},
			},
			fields: fields{
				Service: &mockSendPeersSuccess{},
			},
			want:    &model.Empty{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &P2PServerHandler{
				Service: tt.fields.Service,
			}
			got, err := ss.SendPeers(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerHandler.SendPeers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerHandler.SendPeers() = %v, want %v", got, tt.want)
			}
		})
	}
}