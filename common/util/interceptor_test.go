package util

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewServerInterceptor(t *testing.T) {

	type args struct {
		logger *logrus.Logger
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
				logger: logrus.New(),
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
			got := NewServerInterceptor(tt.args.logger)
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
		logger *logrus.Logger
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
			got := NewClientInterceptor(tt.args.logger)
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
