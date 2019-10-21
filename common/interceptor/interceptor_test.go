package interceptor

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	rpcModel "github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	mockOwnerAddress = "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE"
	mockOwnerSeed    = "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"
)

type (
	mockServerStreamSuccess struct{}
	mockServerStreamNoAuth  struct {
		mockServerStreamSuccess
	}
	mockServerStreamNoMetadata struct {
		mockServerStreamSuccess
	}
	mockServerStreamInvalidAuth struct {
		mockServerStreamSuccess
	}
)

func (*mockServerStreamSuccess) SetHeader(metadata.MD) error  { return nil }
func (*mockServerStreamSuccess) SendHeader(metadata.MD) error { return nil }
func (*mockServerStreamSuccess) SetTrailer(metadata.MD)       {}
func (*mockServerStreamSuccess) Context() context.Context {
	currentTime := uint64(time.Now().Unix())
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(uint32(rpcModel.RequestType_GetNodeHardware)))
	buffer.Write(util.ConvertUint64ToBytes(currentTime))
	sig := (crypto.NewSignature()).Sign(
		buffer.Bytes(),
		constant.SignatureTypeDefault,
		mockOwnerSeed,
	)
	buffer.Write(sig)
	ctx := context.Background()
	md := metadata.Pairs("authorization", base64.StdEncoding.EncodeToString(buffer.Bytes()))
	ctx = metadata.NewIncomingContext(ctx, md)
	return ctx
}
func (*mockServerStreamSuccess) SendMsg(m interface{}) error { return nil }
func (*mockServerStreamSuccess) RecvMsg(m interface{}) error { return nil }

func (*mockServerStreamNoAuth) Context() context.Context {
	ctx := context.Background()
	md := metadata.Pairs("foo", "bar")
	ctx = metadata.NewIncomingContext(ctx, md)
	return ctx
}

func (*mockServerStreamNoMetadata) Context() context.Context {
	ctx := context.Background()
	return ctx
}

func (*mockServerStreamInvalidAuth) Context() context.Context {
	// sign with invalid request type parameter
	currentTime := uint64(time.Now().Unix())
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(uint32(1435)))
	buffer.Write(util.ConvertUint64ToBytes(currentTime))
	sig := (crypto.NewSignature()).Sign(
		buffer.Bytes(),
		constant.SignatureTypeDefault,
		mockOwnerSeed,
	)
	buffer.Write(sig)
	ctx := context.Background()
	md := metadata.Pairs("authorization", base64.StdEncoding.EncodeToString(buffer.Bytes()))
	ctx = metadata.NewIncomingContext(ctx, md)
	return ctx
}

func TestNewServerInterceptor(t *testing.T) {
	type args struct {
		logger              *logrus.Logger
		ownerAccountAddress string
	}
	tests := []struct {
		name        string
		args        args
		want        grpc.UnaryServerInterceptor
		wantRecover bool
	}{
		{
			name: "wantRecover",
			args: args{
				logger:              logrus.New(),
				ownerAccountAddress: "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
			},
			want: func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				return nil, status.Errorf(codes.Internal, "there's something wrong")
			},
			wantRecover: true,
		},
		{
			name: "wantNotRecover",
			args: args{
				logger: logrus.New(),
			},
			want: func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				return nil, status.Errorf(codes.Internal, "there's something wrong")
			},
			wantRecover: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewServerInterceptor(tt.args.logger, tt.args.ownerAccountAddress)
			if cmp.Equal(got, tt.want) {
				t.Errorf("NewInterceptor() = %v, want %v", got, tt.want)
			}
			testServerInterceptor(got, tt.wantRecover)
		})
	}
}

func testServerInterceptor(fn grpc.UnaryServerInterceptor, wantRecover bool) {
	var (
		handler grpc.UnaryHandler
	)
	if wantRecover {
		handler = func(ctx context.Context, req interface{}) (resp interface{}, err error) {
			panic(handler)
		}
	} else {
		handler = func(ctx context.Context, req interface{}) (resp interface{}, err error) {
			return nil, status.Errorf(codes.Internal, "there's something wrong")
		}
	}
	_, _ = fn(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
}

func TestNewClientInterceptor(t *testing.T) {
	type args struct {
		logger        *logrus.Logger
		ignoredErrors map[codes.Code]string
	}
	tests := []struct {
		name        string
		args        args
		want        grpc.UnaryClientInterceptor
		wantRecover bool
	}{
		{
			name: "wantRecover",
			args: args{
				logger:        logrus.New(),
				ignoredErrors: nil,
			},
			want: func(
				ctx context.Context,
				method string,
				req, reply interface{},
				cc *grpc.ClientConn,
				invoker grpc.UnaryInvoker,
				opts ...grpc.CallOption) error {
				return status.Errorf(codes.Internal, "there's something wrong")
			},
			wantRecover: false,
		},
		{
			name: "wantRecover",
			args: args{
				logger: logrus.New(),
			},
			want: func(
				ctx context.Context,
				method string,
				req, reply interface{},
				cc *grpc.ClientConn,
				invoker grpc.UnaryInvoker,
				opts ...grpc.CallOption) error {
				return status.Errorf(codes.Internal, "there's something wrong")
			},
			wantRecover: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewClientInterceptor(tt.args.logger, tt.args.ignoredErrors)
			if cmp.Equal(got, tt.want) {
				t.Errorf("NewClientInterceptor() = %v, want %v", got, tt.want)
			}
			testClientInterceptor(got, tt.wantRecover)
		})
	}
}

func testClientInterceptor(fn grpc.UnaryClientInterceptor, wantRecover bool) {
	var (
		invoker grpc.UnaryInvoker
	)
	if wantRecover {
		invoker = func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			panic(invoker)
		}
	} else {
		invoker = func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return status.Errorf(codes.Internal, "there's something wrong")
		}
	}

	cc, _ := grpc.Dial("127.0.0.1:8001", grpc.WithInsecure())
	_ = fn(context.Background(), "testMethod", nil, nil, cc, invoker, nil)

}

func TestNewNodeAdminAuthStreamInterceptor(t *testing.T) {

	type args struct {
		ownerAddress string
		fullMethod   string
		handler      grpc.StreamHandler
		serverStream grpc.ServerStream
	}
	tests := []struct {
		name         string
		args         args
		want         grpc.StreamServerInterceptor
		wantInnerErr bool
	}{
		{
			name: "wantErr:noMetadata",
			args: args{
				ownerAddress: mockOwnerAddress,
				fullMethod:   "/service.NodeHardwareService/GetNodeHardware",
				serverStream: &mockServerStreamNoMetadata{},
			},
			want: func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
				return nil
			},
			wantInnerErr: true,
		},
		{
			name: "wantErr:noAuth",
			args: args{
				ownerAddress: mockOwnerAddress,
				fullMethod:   "/service.NodeHardwareService/GetNodeHardware",
				serverStream: &mockServerStreamNoAuth{},
			},
			want: func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
				return nil
			},
			wantInnerErr: true,
		},
		{
			name: "wantSuccess:GetNodeHardware",
			args: args{
				ownerAddress: mockOwnerAddress,
				fullMethod:   "/service.NodeHardwareService/GetNodeHardware",
				serverStream: &mockServerStreamSuccess{},
				handler: func(srv interface{}, stream grpc.ServerStream) error {
					return nil
				},
			},
			want: func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
				return nil
			},
			wantInnerErr: false,
		},
		{
			name: "wantFail:invalidSignature",
			args: args{
				ownerAddress: mockOwnerAddress,
				fullMethod:   "/service.NodeHardwareService/GetNodeHardware",
				serverStream: &mockServerStreamInvalidAuth{},
				handler: func(srv interface{}, stream grpc.ServerStream) error {
					return nil
				},
			},
			want: func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
				return nil
			},
			wantInnerErr: true,
		},
		{
			name: "wantSuccess:unprotectedService",
			args: args{
				ownerAddress: mockOwnerAddress,
				fullMethod:   "/nonProtectedService",
				serverStream: &mockServerStreamSuccess{},
				handler: func(srv interface{}, stream grpc.ServerStream) error {
					return nil
				},
			},
			want: func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
				return nil
			},
			wantInnerErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewStreamInterceptor(tt.args.ownerAddress)
			if cmp.Equal(got, tt.want) {
				t.Errorf("NewInterceptor() = %v, want %v", got, tt.want)
			}
			info := &grpc.StreamServerInfo{
				FullMethod:     tt.args.fullMethod,
				IsClientStream: true,
				IsServerStream: true,
			}
			err := got(nil, tt.args.serverStream, info, tt.args.handler)
			if err != nil && tt.wantInnerErr != true {
				t.Error("unexpected error occurred")
			}
		})
	}
}
