package handler

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	service2 "github.com/zoobc/zoobc-core/p2p/service"
)

type (
	mockGetPeerInfoError struct {
		service2.P2PServerServiceInterface
	}
	mockGetPeerInfoSuccess struct {
		service2.P2PServerServiceInterface
	}
)

func (*mockGetPeerInfoError) GetPeerInfo(ctx context.Context, req *model.GetPeerInfoRequest) (*model.GetPeerInfoResponse, error) {
	return nil, errors.New("Error GetPeerInfo")
}
func (*mockGetPeerInfoSuccess) GetPeerInfo(ctx context.Context, req *model.GetPeerInfoRequest) (*model.GetPeerInfoResponse, error) {
	return &model.GetPeerInfoResponse{}, nil
}

func TestP2PServerHandler_GetPeerInfo(t *testing.T) {
	type fields struct {
		Service service2.P2PServerServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetPeerInfoRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetPeerInfoResponse
		wantErr bool
	}{
		{
			name: "GetPeerInfo:Error",
			fields: fields{
				Service: &mockGetPeerInfoError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPeerInfo:Success",
			fields: fields{
				Service: &mockGetPeerInfoSuccess{},
			},
			want:    &model.GetPeerInfoResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &P2PServerHandler{
				Service: tt.fields.Service,
			}
			got, err := ss.GetPeerInfo(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerHandler.GetPeerInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerHandler.GetPeerInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

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

type (
	mockGetCumulativeDifficultyError struct {
		service2.P2PServerServiceInterface
	}
	mockGetCumulativeDifficultySuccess struct {
		service2.P2PServerServiceInterface
	}
)

func (*mockGetCumulativeDifficultyError) GetCumulativeDifficulty(ctx context.Context, chainType chaintype.ChainType,
) (*model.GetCumulativeDifficultyResponse, error) {
	return nil, errors.New("Error GetCumulativeDifficulty")
}

func (*mockGetCumulativeDifficultySuccess) GetCumulativeDifficulty(ctx context.Context, chainType chaintype.ChainType,
) (*model.GetCumulativeDifficultyResponse, error) {
	return &model.GetCumulativeDifficultyResponse{}, nil
}

func TestP2PServerHandler_GetCumulativeDifficulty(t *testing.T) {
	type fields struct {
		Service service2.P2PServerServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetCumulativeDifficultyRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetCumulativeDifficultyResponse
		wantErr bool
	}{
		{
			name: "GetCumulativeDifficulty:Error",
			fields: fields{
				Service: &mockGetCumulativeDifficultyError{},
			},
			args: args{
				req: &model.GetCumulativeDifficultyRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetCumulativeDifficulty:Success",
			fields: fields{
				Service: &mockGetCumulativeDifficultySuccess{},
			},
			args: args{
				req: &model.GetCumulativeDifficultyRequest{},
			},
			want:    &model.GetCumulativeDifficultyResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &P2PServerHandler{
				Service: tt.fields.Service,
			}
			got, err := ss.GetCumulativeDifficulty(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerHandler.GetCumulativeDifficulty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerHandler.GetCumulativeDifficulty() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetCommonMilestoneBlockIDsSuccess struct {
		service2.P2PServerServiceInterface
	}
)

func (*mockGetCommonMilestoneBlockIDsSuccess) GetCommonMilestoneBlockIDs(ctx context.Context, chainType chaintype.ChainType,
	lastBlockID, lastMilestoneBlockID int64) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	return &model.GetCommonMilestoneBlockIdsResponse{}, nil
}
func TestP2PServerHandler_GetCommonMilestoneBlockIDs(t *testing.T) {
	type fields struct {
		Service service2.P2PServerServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetCommonMilestoneBlockIdsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetCommonMilestoneBlockIdsResponse
		wantErr bool
	}{
		{
			name: "GetCommonMilestoneBlockIDs:Error",
			args: args{
				req: &model.GetCommonMilestoneBlockIdsRequest{
					ChainType:            int32(0),
					LastBlockID:          int64(0),
					LastMilestoneBlockID: int64(0),
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetCommonMilestoneBlockIDs:Success",
			args: args{
				req: &model.GetCommonMilestoneBlockIdsRequest{
					ChainType:            int32(1),
					LastBlockID:          int64(1),
					LastMilestoneBlockID: int64(1),
				},
			},
			fields: fields{
				Service: &mockGetCommonMilestoneBlockIDsSuccess{},
			},
			want:    &model.GetCommonMilestoneBlockIdsResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &P2PServerHandler{
				Service: tt.fields.Service,
			}
			got, err := ss.GetCommonMilestoneBlockIDs(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerHandler.GetCommonMilestoneBlockIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerHandler.GetCommonMilestoneBlockIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetNextBlockIDsError struct {
		service2.P2PServerServiceInterface
	}
	mockGetNextBlockIDsSuccess struct {
		service2.P2PServerServiceInterface
	}
)

func (*mockGetNextBlockIDsError) GetNextBlockIDs(ctx context.Context, chainType chaintype.ChainType,
	reqLimit uint32, reqBlockID int64) ([]int64, error) {
	return nil, errors.New("Error GetNextBlockIDs")
}

func (*mockGetNextBlockIDsSuccess) GetNextBlockIDs(ctx context.Context, chainType chaintype.ChainType,
	reqLimit uint32, reqBlockID int64) ([]int64, error) {
	return []int64{}, nil
}

func TestP2PServerHandler_GetNextBlockIDs(t *testing.T) {
	type fields struct {
		Service service2.P2PServerServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetNextBlockIdsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.BlockIdsResponse
		wantErr bool
	}{
		{
			name: "GetNextBlockIDs:Error",
			fields: fields{
				Service: &mockGetNextBlockIDsError{},
			},
			args: args{
				req: &model.GetNextBlockIdsRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetNextBlockIDs:Success",
			fields: fields{
				Service: &mockGetNextBlockIDsSuccess{},
			},
			args: args{
				req: &model.GetNextBlockIdsRequest{},
			},
			want: &model.BlockIdsResponse{
				BlockIds: []int64{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &P2PServerHandler{
				Service: tt.fields.Service,
			}
			got, err := ss.GetNextBlockIDs(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerHandler.GetNextBlockIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerHandler.GetNextBlockIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetNextBlocksError struct {
		service2.P2PServerServiceInterface
	}
	mockGetNextBlocksSuccess struct {
		service2.P2PServerServiceInterface
	}
)

func (*mockGetNextBlocksError) GetNextBlocks(ctx context.Context, chainType chaintype.ChainType, blockID int64,
	blockIDList []int64) (*model.BlocksData, error) {
	return nil, errors.New("Error GetNextBlocks")
}
func (*mockGetNextBlocksSuccess) GetNextBlocks(ctx context.Context, chainType chaintype.ChainType, blockID int64,
	blockIDList []int64) (*model.BlocksData, error) {
	return &model.BlocksData{}, nil
}

func TestP2PServerHandler_GetNextBlocks(t *testing.T) {
	type fields struct {
		Service service2.P2PServerServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetNextBlocksRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.BlocksData
		wantErr bool
	}{
		{
			name: "GetNextBlocks:Error",
			fields: fields{
				Service: &mockGetNextBlocksError{},
			},
			args: args{
				req: &model.GetNextBlocksRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetNextBlocks:Success",
			fields: fields{
				Service: &mockGetNextBlocksSuccess{},
			},
			args: args{
				req: &model.GetNextBlocksRequest{},
			},
			want:    &model.BlocksData{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &P2PServerHandler{
				Service: tt.fields.Service,
			}
			got, err := ss.GetNextBlocks(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerHandler.GetNextBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerHandler.GetNextBlocks() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockSendBlockError struct {
		service2.P2PServerServiceInterface
	}
	mockSendBlockSuccess struct {
		service2.P2PServerServiceInterface
	}
)

func (*mockSendBlockError) SendBlock(ctx context.Context, chainType chaintype.ChainType, block *model.Block,
	senderPublicKey []byte) (*model.SendBlockResponse, error) {
	return nil, errors.New("Error SendBlock")
}
func (*mockSendBlockSuccess) SendBlock(ctx context.Context, chainType chaintype.ChainType, block *model.Block,
	senderPublicKey []byte) (*model.SendBlockResponse, error) {
	return &model.SendBlockResponse{}, nil
}

func TestP2PServerHandler_SendBlock(t *testing.T) {
	type fields struct {
		Service service2.P2PServerServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.SendBlockRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.SendBlockResponse
		wantErr bool
	}{
		{
			name: "SendBlock:Error",
			fields: fields{
				Service: &mockSendBlockError{},
			},
			args: args{
				req: &model.SendBlockRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "SendBlock:Success",
			fields: fields{
				Service: &mockSendBlockSuccess{},
			},
			args: args{
				req: &model.SendBlockRequest{},
			},
			want:    &model.SendBlockResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &P2PServerHandler{
				Service: tt.fields.Service,
			}
			got, err := ss.SendBlock(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerHandler.SendBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerHandler.SendBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockSendTransactionError struct {
		service2.P2PServerServiceInterface
	}
	mockSendTransactionSuccess struct {
		service2.P2PServerServiceInterface
	}
)

func (*mockSendTransactionError) SendTransaction(ctx context.Context, chainType chaintype.ChainType,
	transactionBytes, senderPublicKey []byte) (*model.SendTransactionResponse, error) {
	return nil, errors.New("Error SendTransaction")
}

func (*mockSendTransactionSuccess) SendTransaction(ctx context.Context, chainType chaintype.ChainType,
	transactionBytes, senderPublicKey []byte) (*model.SendTransactionResponse, error) {
	return &model.SendTransactionResponse{}, nil
}

func TestP2PServerHandler_SendTransaction(t *testing.T) {
	type fields struct {
		Service service2.P2PServerServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.SendTransactionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.SendTransactionResponse
		wantErr bool
	}{
		{
			name: "SendTransaction:Error",
			fields: fields{
				Service: &mockSendTransactionError{},
			},
			args: args{
				req: &model.SendTransactionRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "SendTransaction:Success",
			fields: fields{
				Service: &mockSendTransactionSuccess{},
			},
			args: args{
				req: &model.SendTransactionRequest{},
			},
			want:    &model.SendTransactionResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &P2PServerHandler{
				Service: tt.fields.Service,
			}
			got, err := ss.SendTransaction(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerHandler.SendTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerHandler.SendTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockSendBlockTransactionsError struct {
		service2.P2PServerServiceInterface
	}
	mockSendBlockTransactionsSuccess struct {
		service2.P2PServerServiceInterface
	}
)

func (*mockSendBlockTransactionsError) SendBlockTransactions(ctx context.Context, chainType chaintype.ChainType,
	transactionsBytes [][]byte, senderPublicKey []byte) (*model.SendBlockTransactionsResponse, error) {
	return nil, errors.New("Error SendBlockTransactions")
}

func (*mockSendBlockTransactionsSuccess) SendBlockTransactions(ctx context.Context, chainType chaintype.ChainType,
	transactionsBytes [][]byte, senderPublicKey []byte) (*model.SendBlockTransactionsResponse, error) {
	return &model.SendBlockTransactionsResponse{}, nil
}

func TestP2PServerHandler_SendBlockTransactions(t *testing.T) {
	type fields struct {
		Service service2.P2PServerServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.SendBlockTransactionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.SendBlockTransactionsResponse
		wantErr bool
	}{
		{
			name: "SendBlockTransactions:Error",
			fields: fields{
				Service: &mockSendBlockTransactionsError{},
			},
			args: args{
				req: &model.SendBlockTransactionsRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "SendBlockTransactions:Success",
			fields: fields{
				Service: &mockSendBlockTransactionsSuccess{},
			},
			args: args{
				req: &model.SendBlockTransactionsRequest{},
			},
			want:    &model.SendBlockTransactionsResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &P2PServerHandler{
				Service: tt.fields.Service,
			}
			got, err := ss.SendBlockTransactions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerHandler.SendBlockTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerHandler.SendBlockTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockRequestBlockTransactionsError struct {
		service2.P2PServerServiceInterface
	}
	mockRequestBlockTransactionsSuccess struct {
		service2.P2PServerServiceInterface
	}
)

func (*mockRequestBlockTransactionsError) RequestBlockTransactions(ctx context.Context, chainType chaintype.ChainType,
	blockID int64, transactionsIDs []int64) (*model.Empty, error) {
	return nil, errors.New("Error RequestBlockTransactions")
}

func (*mockRequestBlockTransactionsSuccess) RequestBlockTransactions(ctx context.Context, chainType chaintype.ChainType,
	blockID int64, transactionsIDs []int64) (*model.Empty, error) {
	return &model.Empty{}, nil
}

func TestP2PServerHandler_RequestBlockTransactions(t *testing.T) {
	type fields struct {
		Service service2.P2PServerServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.RequestBlockTransactionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Empty
		wantErr bool
	}{
		{
			name: "RequestBlockTransactions:Error",
			fields: fields{
				Service: &mockRequestBlockTransactionsError{},
			},
			args: args{
				req: &model.RequestBlockTransactionsRequest{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "RequestBlockTransactions:Success",
			fields: fields{
				Service: &mockRequestBlockTransactionsSuccess{},
			},
			args: args{
				req: &model.RequestBlockTransactionsRequest{},
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
			got, err := ss.RequestBlockTransactions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerHandler.RequestBlockTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerHandler.RequestBlockTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockRequestDownloadFileError struct {
		service2.P2PServerServiceInterface
	}
	mockRequestDownloadFileSuccess struct {
		service2.P2PServerServiceInterface
	}
)

func (*mockRequestDownloadFileError) RequestDownloadFile(context.Context, []byte, []string) (*model.FileDownloadResponse, error) {
	return nil, errors.New("Error RequestDownloadFile")
}

func (*mockRequestDownloadFileSuccess) RequestDownloadFile(context.Context, []byte, []string) (*model.FileDownloadResponse, error) {
	return &model.FileDownloadResponse{}, nil
}

func TestP2PServerHandler_RequestFileDownload(t *testing.T) {
	type fields struct {
		Service service2.P2PServerServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.FileDownloadRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.FileDownloadResponse
		wantErr bool
	}{
		{
			name: "RequestFileDownload:NotContainAnyFileName",
			args: args{
				req: &model.FileDownloadRequest{
					FileChunkNames: []string{},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "RequestFileDownload:Error",
			args: args{
				req: &model.FileDownloadRequest{
					FileChunkNames: []string{"mockName"},
				},
			},
			fields: fields{
				Service: &mockRequestDownloadFileError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "RequestFileDownload:Success",
			args: args{
				req: &model.FileDownloadRequest{
					FileChunkNames: []string{"mockName"},
				},
			},
			fields: fields{
				Service: &mockRequestDownloadFileSuccess{},
			},
			want:    &model.FileDownloadResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &P2PServerHandler{
				Service: tt.fields.Service,
			}
			got, err := ss.RequestFileDownload(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("P2PServerHandler.RequestFileDownload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("P2PServerHandler.RequestFileDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}
